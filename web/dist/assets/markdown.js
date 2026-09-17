import { renderMermaid } from "./mermaid-lite.js";

export function markdownToHtml(md, docPath) {
  if (!md) return "";
  const parts = [];
  const fence = /```(\w*)\n([\s\S]*?)```/g;
  let last = 0;
  let m;
  while ((m = fence.exec(md))) {
    parts.push(inlineMarkdown(md.slice(last, m.index), docPath));
    const lang = (m[1] || "").toLowerCase();
    const body = m[2].replace(/\n$/, "");
    if (lang === "mermaid") {
      parts.push(renderMermaid(body));
    } else {
      parts.push(`<pre><code>${escapeHtml(body)}</code></pre>`);
    }
    last = m.index + m[0].length;
  }
  parts.push(inlineMarkdown(md.slice(last), docPath));
  return parts.join("");
}

function inlineMarkdown(src, docPath) {
  const lines = src.split("\n");
  const out = [];
  for (const line of lines) {
    if (/^### /.test(line)) out.push(`<h3>${inline(line.slice(4), docPath)}</h3>`);
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
