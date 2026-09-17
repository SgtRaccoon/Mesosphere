import { renderMermaid } from "./mermaid-lite.js";

export function markdownToHtml(md, docPath) {
  if (!md) return "";
  let rest = md;
  let metaHtml = "";
  const fm = rest.match(/^---\r?\n([\s\S]*?)\r?\n---\r?\n?/);
  if (fm) {
    metaHtml = renderFrontmatter(fm[1]);
    rest = rest.slice(fm[0].length);
  }
  const parts = [metaHtml];
  const fence = /```(\w*)\r?\n([\s\S]*?)```/g;
  let last = 0;
  let m;
  while ((m = fence.exec(rest))) {
    parts.push(inlineMarkdown(rest.slice(last, m.index), docPath));
    const lang = (m[1] || "").toLowerCase();
    const body = m[2].replace(/\n$/, "");
    if (lang === "mermaid") {
      parts.push(renderMermaid(body));
    } else {
      parts.push(`<pre><code>${escapeHtml(body)}</code></pre>`);
    }
    last = m.index + m[0].length;
  }
  parts.push(inlineMarkdown(rest.slice(last), docPath));
  return parts.join("");
}

function renderFrontmatter(body) {
  const rows = String(body)
    .split(/\r?\n/)
    .map((l) => l.trim())
    .filter(Boolean)
    .map((line) => {
      const i = line.indexOf(":");
      if (i === -1) {
        return `<div class="md-meta-line">${escapeHtml(line)}</div>`;
      }
      const k = line.slice(0, i).trim();
      const v = line.slice(i + 1).trim();
      return `<div class="md-meta-row"><span class="md-meta-key">${escapeHtml(k)}</span><span class="md-meta-val">${escapeHtml(v)}</span></div>`;
    })
    .join("");
  return `<aside class="md-meta-box">${rows}</aside>`;
}

function isTableRow(line) {
  const t = line.trim();
  return t.includes("|") && !t.startsWith("```");
}

function isTableSep(line) {
  const cells = splitRow(line);
  return cells.length > 0 && cells.every((c) => /^:?-{3,}:?$/.test(c.replace(/\s/g, "")) || c === "");
}

function splitRow(line) {
  let t = line.trim();
  if (t.startsWith("|")) t = t.slice(1);
  if (t.endsWith("|")) t = t.slice(0, -1);
  return t.split("|").map((c) => c.trim());
}

function renderTable(rows, docPath) {
  const body = rows.filter((r) => !isTableSep(r));
  if (!body.length) return "";
  const head = splitRow(body[0]);
  const rest = body.slice(1);
  const th = head.map((c) => `<th>${inline(c, docPath)}</th>`).join("");
  const trs = rest
    .map((r) => `<tr>${splitRow(r).map((c) => `<td>${inline(c, docPath)}</td>`).join("")}</tr>`)
    .join("");
  return `<table><thead><tr>${th}</tr></thead><tbody>${trs}</tbody></table>`;
}

function inlineMarkdown(src, docPath) {
  const lines = src.split("\n");
  const out = [];
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    if (isTableRow(line)) {
      const block = [line];
      while (i + 1 < lines.length && isTableRow(lines[i + 1])) {
        i++;
        block.push(lines[i]);
      }
      out.push(renderTable(block, docPath));
    } else if (/^### /.test(line)) out.push(`<h3>${inline(line.slice(4), docPath)}</h3>`);
    else if (/^## /.test(line)) out.push(`<h2>${inline(line.slice(3), docPath)}</h2>`);
    else if (/^# /.test(line)) out.push(`<h1>${inline(line.slice(2), docPath)}</h1>`);
    else if (/^[-*] /.test(line)) out.push(`<li>${inline(line.slice(2), docPath)}</li>`);
    else if (line.trim() === "") out.push("");
    else out.push(`<p>${inline(line, docPath)}</p>`);
  }
  return out.join("\n");
}

function inline(text, docPath) {
  let s = escapeHtml(text);
  s = s.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, (_, alt, src) => {
    const url = resolveAsset(docPath, src);
    return `<img alt="${alt}" src="${url}" />`;
  });
  s = s.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2">$1</a>');
  s = s.replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>");
  return s;
}

function resolveAsset(docPath, src) {
  if (/^[a-z]+:/i.test(src) || src.startsWith("/")) return src;
  const dir = (docPath || "").split("/").slice(0, -1).join("/");
  const joined = dir ? `${dir}/${src}` : src;
  return joined.replace(/\/\.\//g, "/");
}

function escapeHtml(s) {
  return String(s)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}
