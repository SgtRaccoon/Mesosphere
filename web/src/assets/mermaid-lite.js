/** Minimal air-gapped mermaid renderer for flowchart-style diagrams. */
export function renderMermaid(code) {
  const lines = String(code)
    .split(/\r?\n/)
    .map((l) => l.trim())
    .filter(Boolean);
  const nodes = new Map();
  const edges = [];

  function ensureNode(raw) {
    const parsed = parseNodeToken(raw);
    if (!parsed) return null;
    const prev = nodes.get(parsed.id);
    if (!prev || prev === parsed.id) nodes.set(parsed.id, parsed.label);
    else if (parsed.label !== parsed.id) nodes.set(parsed.id, parsed.label);
    else if (!nodes.has(parsed.id)) nodes.set(parsed.id, parsed.label);
    return parsed.id;
  }

  for (const line of lines) {
    if (/^(graph|flowchart)\b/i.test(line)) continue;
    if (/^subgraph\b/i.test(line) || line === "end") continue;

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

  const ids = [...nodes.keys()];
  if (!ids.length) {
    return `<pre class="mermaid-fallback"><code>${escapeXml(code)}</code></pre>`;
  }
  const w = Math.max(240, ids.length * 140);
  const h = 140 + edges.length * 12;
  const boxes = ids
    .map((id, i) => {
      const x = 20 + i * 130;
      const y = 40;
      const label = escapeXml(nodes.get(id));
      return `<g class="mermaid-node"><rect x="${x}" y="${y}" width="110" height="40" rx="6" fill="#1a222c" stroke="#7aa2f7"/>
        <text x="${x + 55}" y="${y + 25}" text-anchor="middle" fill="#e7ecf1" font-size="12">${label}</text></g>`;
    })
    .join("");
  const index = Object.fromEntries(ids.map((id, i) => [id, i]));
  const arrows = edges
    .map((e) => {
      const a = index[e.from];
      const b = index[e.to];
      if (a == null || b == null) return "";
      const x1 = 20 + a * 130 + (a < b ? 110 : 0);
      const x2 = 20 + b * 130 + (a < b ? 0 : 110);
      const label = e.label
        ? `<text x="${(x1 + x2) / 2}" y="52" text-anchor="middle" fill="#9aa8b8" font-size="10">${escapeXml(e.label)}</text>`
        : "";
      return `<line x1="${x1}" y1="60" x2="${x2}" y2="60" stroke="#7aa2f7" marker-end="url(#arrow)"/>${label}`;
    })
    .join("");
  return `<svg class="mermaid-svg" xmlns="http://www.w3.org/2000/svg" width="${w}" height="${h}" viewBox="0 0 ${w} ${h}">
    <defs><marker id="arrow" markerWidth="8" markerHeight="8" refX="8" refY="4" orient="auto"><path d="M0,0 L8,4 L0,8 z" fill="#7aa2f7"/></marker></defs>
    ${arrows}${boxes}</svg>`;
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
