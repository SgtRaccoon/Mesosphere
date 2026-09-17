/** Minimal air-gapped mermaid renderer for flowchart-style diagrams. */
export function renderMermaid(code) {
  const lines = String(code)
    .split("\n")
    .map((l) => l.trim())
    .filter(Boolean);
  const nodes = new Map();
  const edges = [];
  for (const line of lines) {
    if (/^(graph|flowchart)\b/i.test(line)) continue;
    const edge = line.match(/^(\w+)\s*-->\s*(\w+)(?:\s*:\s*(.*))?$/);
    if (edge) {
      nodes.set(edge[1], nodes.get(edge[1]) || edge[1]);
      nodes.set(edge[2], nodes.get(edge[2]) || edge[2]);
      edges.push({ from: edge[1], to: edge[2], label: edge[3] || "" });
      continue;
    }
    const labeled = line.match(/^(\w+)\[([^\]]+)\]$/);
    if (labeled) nodes.set(labeled[1], labeled[2]);
  }
  const ids = [...nodes.keys()];
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
      const x1 = 20 + a * 130 + 110;
      const x2 = 20 + b * 130;
      return `<line x1="${x1}" y1="60" x2="${x2}" y2="60" stroke="#7aa2f7" marker-end="url(#arrow)"/>`;
    })
    .join("");
  return `<svg class="mermaid-svg" xmlns="http://www.w3.org/2000/svg" width="${w}" height="${h}" viewBox="0 0 ${w} ${h}">
    <defs><marker id="arrow" markerWidth="8" markerHeight="8" refX="8" refY="4" orient="auto"><path d="M0,0 L8,4 L0,8 z" fill="#7aa2f7"/></marker></defs>
    ${arrows}${boxes}</svg>`;
}

function escapeXml(s) {
  return String(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}
