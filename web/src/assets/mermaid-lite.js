/** Minimal air-gapped mermaid renderer for flowchart-style diagrams. */
export function renderMermaid(code) {
  const parsed = parseDiagram(code);
  if (!parsed.ids.length) {
    return `<pre class="mermaid-fallback"><code>${escapeXml(code)}</code></pre>`;
  }
  const laid = layout(parsed);
  return `<svg class="mermaid-svg" xmlns="http://www.w3.org/2000/svg" width="${laid.w}" height="${laid.h}" viewBox="0 0 ${laid.w} ${laid.h}">
    <defs><marker id="arrow" markerWidth="8" markerHeight="8" refX="8" refY="4" orient="auto"><path d="M0,0 L8,4 L0,8 z" fill="#7aa2f7"/></marker></defs>
    ${laid.groups}${laid.arrows}${laid.boxes}</svg>`;
}

function parseDiagram(code) {
  const lines = String(code)
    .split(/\r?\n/)
    .map((l) => l.trim())
    .filter(Boolean);
  const nodes = new Map();
  const edges = [];
  const subgraphs = new Map();
  const rootOrder = [];
  const stack = [];
  const seenNode = new Set();
  let vertical = true;

  function currentParent() {
    return stack.length ? stack[stack.length - 1] : null;
  }

  function attach(kind, id) {
    const parent = currentParent();
    const item = { kind, id };
    if (parent) subgraphs.get(parent).childOrder.push(item);
    else rootOrder.push(item);
  }

  function ensureNode(raw) {
    const parsed = parseNodeToken(raw);
    if (!parsed) return null;
    const prev = nodes.get(parsed.id);
    if (!prev || prev === parsed.id) nodes.set(parsed.id, parsed.label);
    else if (parsed.label !== parsed.id) nodes.set(parsed.id, parsed.label);
    else if (!nodes.has(parsed.id)) nodes.set(parsed.id, parsed.label);
    if (!seenNode.has(parsed.id)) {
      seenNode.add(parsed.id);
      attach("node", parsed.id);
    }
    return parsed.id;
  }

  for (const line of lines) {
    const dir = line.match(/^(graph|flowchart)\s+(TB|TD|BT|LR|RL)\b/i);
    if (dir) {
      const d = dir[2].toUpperCase();
      vertical = d === "TB" || d === "TD" || d === "BT";
      continue;
    }
    if (/^(graph|flowchart)\b/i.test(line)) continue;

    const sg = line.match(/^subgraph\s+(?:(\w+)\s*)?(?:\[([^\]]+)\]|"([^"]+)"|(.+))?$/i);
    if (sg && /^subgraph\b/i.test(line)) {
      const title = (sg[2] || sg[3] || sg[4] || sg[1] || "group").trim();
      const id = (sg[1] || `sg${subgraphs.size}`).trim() || `sg${subgraphs.size}`;
      subgraphs.set(id, { id, title, childOrder: [] });
      attach("sg", id);
      stack.push(id);
      continue;
    }
    if (/^end$/i.test(line)) {
      stack.pop();
      continue;
    }

    const edge = line.match(/^(.+?)\s*(-->|---|==>)\s*(?:\|([^|]+)\|\s*)?(.+)$/);
    if (edge) {
      const from = ensureNode(edge[1]);
      const to = ensureNode(edge[4]);
      if (from && to) edges.push({ from, to, label: (edge[3] || "").trim() });
      continue;
    }
    const labeled = parseNodeToken(line);
    if (labeled) ensureNode(line);
  }

  return { nodes, edges, subgraphs, rootOrder, vertical, ids: [...nodes.keys()] };
}

function layout(parsed) {
  const { nodes, edges, subgraphs, rootOrder, vertical, ids } = parsed;
  const pad = 20;
  const items = rootOrder.length ? rootOrder : ids.map((id) => ({ kind: "node", id }));
  const placed = placeItems(items, nodes, subgraphs, vertical, pad, pad);
  const w = placed.x + placed.w + pad;
  const h = placed.y + placed.h + pad;

  const posById = {};
  collectNodePos(placed, posById);

  const boxes = Object.values(posById)
    .map((p) => nodeSvg(p))
    .join("");
  const groups = collectGroupSvg(placed).join("");

  const arrows = edges
    .map((e) => {
      const pa = posById[e.from];
      const pb = posById[e.to];
      if (!pa || !pb) return "";
      let x1, y1, x2, y2;
      if (vertical) {
        x1 = pa.x + pa.w / 2;
        x2 = pb.x + pb.w / 2;
        if (pa.y < pb.y) {
          y1 = pa.y + pa.h;
          y2 = pb.y;
        } else {
          y1 = pa.y;
          y2 = pb.y + pb.h;
        }
      } else {
        y1 = pa.y + pa.h / 2;
        y2 = pb.y + pb.h / 2;
        if (pa.x < pb.x) {
          x1 = pa.x + pa.w;
          x2 = pb.x;
        } else {
          x1 = pa.x;
          x2 = pb.x + pb.w;
        }
      }
      const label = e.label
        ? `<text x="${(x1 + x2) / 2}" y="${(y1 + y2) / 2 - 4}" text-anchor="middle" fill="#9aa8b8" font-size="10">${escapeXml(e.label)}</text>`
        : "";
      return `<line x1="${x1}" y1="${y1}" x2="${x2}" y2="${y2}" stroke="#7aa2f7" marker-end="url(#arrow)"/>${label}`;
    })
    .join("");

  return { w, h, boxes, arrows, groups };
}

function placeItems(items, nodes, subgraphs, vertical, x, y) {
  const gap = 36;
  const innerPad = 14;
  const titleH = 20;
  const children = [];
  let cursor = 0;
  let maxCross = 0;

  for (const item of items) {
    if (item.kind === "sg") {
      const sg = subgraphs.get(item.id);
      const innerX = x + innerPad;
      const innerY = y + innerPad + titleH + (vertical ? cursor : 0);
      const ox = vertical ? innerX : x + innerPad + cursor;
      const oy = vertical ? innerY : y + innerPad + titleH;
      const nested = placeItems(sg.childOrder, nodes, subgraphs, vertical, ox, oy);
      const gw = nested.w + innerPad * 2;
      const gh = nested.h + innerPad * 2 + titleH;
      const gx = vertical ? x : x + cursor;
      const gy = vertical ? y + cursor : y;
      children.push({
        kind: "sg",
        id: item.id,
        title: sg.title,
        x: gx,
        y: gy,
        w: gw,
        h: gh,
        children: nested.children,
      });
      if (vertical) {
        cursor += gh + gap;
        maxCross = Math.max(maxCross, gw);
      } else {
        cursor += gw + gap;
        maxCross = Math.max(maxCross, gh);
      }
    } else {
      const b = measureBox(nodes.get(item.id));
      const nx = vertical ? x : x + cursor;
      const ny = vertical ? y + cursor : y;
      children.push({ kind: "node", id: item.id, x: nx, y: ny, w: b.w, h: b.h, lines: b.lines });
      if (vertical) {
        cursor += b.h + gap;
        maxCross = Math.max(maxCross, b.w);
      } else {
        cursor += b.w + gap;
        maxCross = Math.max(maxCross, b.h);
      }
    }
  }

  if (cursor > 0) cursor -= gap;
  const w = vertical ? maxCross : cursor;
  const h = vertical ? cursor : maxCross;
  return { x, y, w: Math.max(w, 0), h: Math.max(h, 0), children };
}

function collectNodePos(placed, out) {
  for (const c of placed.children || []) {
    if (c.kind === "node") out[c.id] = c;
    else collectNodePos(c, out);
  }
}

function collectGroupSvg(placed, acc = []) {
  for (const c of placed.children || []) {
    if (c.kind === "sg") {
      acc.push(
        `<g class="mermaid-subgraph"><rect x="${c.x}" y="${c.y}" width="${c.w}" height="${c.h}" rx="8" fill="#151c24" stroke="#3d4f66" stroke-dasharray="4 3"/>
        <text x="${c.x + 10}" y="${c.y + 16}" fill="#9aa8b8" font-size="11">${escapeXml(c.title)}</text></g>`,
      );
      collectGroupSvg(c, acc);
    }
  }
  return acc;
}

function nodeSvg(p) {
  const text = p.lines
    .map((line, li) => {
      const ty = p.y + 18 + li * 14;
      return `<tspan x="${p.x + p.w / 2}" y="${ty}">${escapeXml(line)}</tspan>`;
    })
    .join("");
  return `<g class="mermaid-node"><rect x="${p.x}" y="${p.y}" width="${p.w}" height="${p.h}" rx="6" fill="#1a222c" stroke="#7aa2f7"/>
        <text text-anchor="middle" fill="#e7ecf1" font-size="12">${text}</text></g>`;
}

function measureBox(label) {
  const lines = wrapLabel(label);
  const longest = lines.reduce((n, l) => Math.max(n, l.length), 0);
  const w = Math.max(110, Math.min(280, 12 + longest * 7.2));
  const h = Math.max(40, 12 + lines.length * 14 + 10);
  return { w, h, lines };
}

function wrapLabel(label) {
  const raw = String(label || "")
    .replace(/<br\s*\/?>/gi, "\n")
    .split(/\n/);
  const out = [];
  const max = 28;
  for (const part of raw) {
    const words = part.split(/\s+/).filter(Boolean);
    if (!words.length) {
      out.push("");
      continue;
    }
    let cur = "";
    for (const w of words) {
      const next = cur ? `${cur} ${w}` : w;
      if (next.length > max && cur) {
        out.push(cur);
        cur = w;
      } else {
        cur = next;
      }
    }
    if (cur) out.push(cur);
  }
  return out.length ? out : [""];
}

function parseNodeToken(raw) {
  const s = String(raw).trim().replace(/;$/, "");
  let m = s.match(/^(\w+)\s*\[([^\]]+)\]$/);
  if (m) return { id: m[1], label: m[2] };
  m = s.match(/^(\w+)\s*\{([^}]+)\}$/);
  if (m) return { id: m[1], label: m[2] };
  m = s.match(/^(\w+)\s*\(([^)]+)\)$/);
  if (m) return { id: m[1], label: m[2] };
  m = s.match(/^(\w+)\s*\(\(([^)]+)\)\)$/);
  if (m) return { id: m[1], label: m[2] };
  m = s.match(/^(\w+)$/);
  if (m) return { id: m[1], label: m[1] };
  m = s.match(/^(\w+)/);
  if (m) return { id: m[1], label: m[1] };
  return null;
}

function escapeXml(s) {
  return String(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}
