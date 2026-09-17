export function buildDocTree(docs) {
  const root = { name: "", children: {}, files: [] };
  for (const doc of docs || []) {
    const parts = String(doc.path || "").split("/").filter(Boolean);
    if (!parts.length) continue;
    let node = root;
    for (let i = 0; i < parts.length - 1; i++) {
      const name = parts[i];
      if (!node.children[name]) node.children[name] = { name, children: {}, files: [] };
      node = node.children[name];
    }
    node.files.push({ ...doc, name: parts[parts.length - 1] });
  }
  return root;
}

export function childDirs(node) {
  return Object.values(node.children || {}).sort((a, b) => a.name.localeCompare(b.name));
}
