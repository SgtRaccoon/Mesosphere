import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const dir = dirname(fileURLToPath(import.meta.url));
const selector = readFileSync(join(dir, "RepoSelector.vue"), "utf8");
const topBar = readFileSync(join(dir, "TopBar.vue"), "utf8");
const paneBar = readFileSync(join(dir, "PaneBar.vue"), "utf8");
const docsView = readFileSync(join(dir, "DocsView.vue"), "utf8");
const mermaidView = readFileSync(join(dir, "MermaidRenderer.vue"), "utf8");

function assert(cond, msg) {
  if (!cond) {
    console.error(msg);
    process.exit(1);
  }
}

assert(selector.includes("repo-grid"), "RepoSelector missing box-grid");
assert(selector.includes("repos"), "RepoSelector must list repos");
assert(selector.includes("update:modelValue"), "RepoSelector must emit selection");
assert(selector.includes("repo-card-meta"), "RepoSelector shows doc/task counts");
assert(selector.includes(":hover"), "RepoSelector hover styles");
assert(selector.includes("text-overflow"), "RepoSelector truncates overflowing card text");
assert(topBar.includes("top-bar"), "TopBar missing top-bar");
assert(topBar.includes("Mesosphere"), "TopBar must show app name");
assert(!topBar.includes("Docs"), "TopBar should not include Docs tab");
assert(paneBar.includes("Docs") && paneBar.includes("Tasks"), "PaneBar missing Docs/Tasks tabs");
assert(paneBar.includes('name="file-text"') && paneBar.includes('name="list-todo"'), "PaneBar Docs/Tasks lucide icons");
assert(paneBar.includes("repo.name"), "PaneBar must show selected repo name");
assert(paneBar.includes("open-repos"), "PaneBar click opens repo cards");
assert(docsView.includes("docs-nav"), "DocsView missing left-hand menu");
assert(docsView.includes("docs-version"), "DocsView missing version dropdown");
assert(docsView.includes("DocTreeNode"), "DocsView missing directory tree");
assert(docsView.includes("markdownToHtml"), "DocsView must render markdown");
assert(mermaidView.includes("renderMermaid"), "MermaidRenderer must render diagrams");

const { markdownToHtml } = await import("../assets/markdown.js");
const html = markdownToHtml("# Title\n\n![img](pic.png)\n\n```mermaid\ngraph TD\nA-->B\n```\n", "docs/a.md");
assert(html.includes("<h1>"), "markdown heading");
assert(html.includes("<svg"), "mermaid svg");
assert(html.includes("docs/pic.png"), "relative image");
const tableHtml = markdownToHtml("| A | B |\n| --- | --- |\n| 1 | 2 |\n", "x.md");
assert(tableHtml.includes("<table>"), "markdown table");
assert(tableHtml.includes("<th>"), "markdown table header");
const labeledMermaid = markdownToHtml(
  "```mermaid\nflowchart LR\nA[Start] --> B{Decision}\nB --> C[Done]\n```\n",
  "x.md",
);
assert(labeledMermaid.includes("<svg"), "labeled mermaid svg");
assert(labeledMermaid.includes("Start"), "mermaid node label");
const { renderMermaid } = await import("../assets/mermaid-lite.js");
const tb = renderMermaid("flowchart TB\nA[Top] --> B[Bottom]");
assert(tb.includes("Top") && tb.includes("Bottom"), "tb labels");
const aY = Number((tb.match(/<rect x="[^"]+" y="([^"]+)"/) || [])[1]);
const rects = [...tb.matchAll(/<rect x="[^"]+" y="([^"]+)"/g)].map((m) => Number(m[1]));
assert(rects.length >= 2 && rects[1] > rects[0], "TB stacks top-to-bottom");
const br = renderMermaid("flowchart LR\nA[Hello<br/>World] --> B[Next]");
assert(br.includes("<tspan") && br.includes("Hello") && br.includes("World"), "br becomes tspans");
assert(!br.includes("&lt;br"), "br not shown as text");
void aY;
const metaHtml = markdownToHtml("---\ntitle: Hello\nstatus: draft\n---\n# Body\n", "x.md");
assert(metaHtml.includes("md-meta-box"), "frontmatter box");
assert(metaHtml.includes("title"), "frontmatter key");
assert(metaHtml.includes("<h1>"), "body after frontmatter");
const hrHtml = markdownToHtml("# A\n\n---\n\n# B\n", "x.md");
assert(hrHtml.includes("<hr"), "horizontal rule");
const listHtml = markdownToHtml("- a\n  - nested\n- b\n", "x.md");
assert(listHtml.includes("<ul>"), "unordered list wrap");
assert((listHtml.match(/<ul>/g) || []).length >= 2, "nested unordered list");
const nestedSg = renderMermaid("flowchart TB\nsubgraph outer [Outer]\nsubgraph inner [Inner]\nA[One] --> B[Two]\nend\nend");
assert(nestedSg.includes("Outer") && nestedSg.includes("Inner"), "nested subgraph titles");
assert((nestedSg.match(/mermaid-subgraph/g) || []).length >= 2, "nested subgraph boxes");
const { buildDocTree, childDirs } = await import("../assets/docs-tree.js");
const tree = buildDocTree([{ path: "docs/a.md" }, { path: "docs/sub/b.md" }]);
assert(childDirs(tree)[0].name === "docs", "tree root folder");
assert(childDirs(childDirs(tree)[0]).length === 1, "nested folder");

const editor = readFileSync(join(dir, "DocsEditor.vue"), "utf8");
const publish = readFileSync(join(dir, "PublishButton.vue"), "utf8");
assert(editor.includes("docs-edit-fab"), "edit fab");
assert(editor.includes("Save"), "save control");
assert(editor.includes("'save'"), "save lucide icon");
assert(publish.includes("Publishing"), "publish loading state");
assert(publish.includes('name="upload"'), "publish lucide icon");
assert(publish.includes("publish-modal"), "conflict modal");

const board = readFileSync(join(dir, "KanbanBoard.vue"), "utf8");
const card = readFileSync(join(dir, "TaskCard.vue"), "utf8");
assert(board.includes("kanban"), "kanban board");
assert(card.includes("task.card_id"), "card id");
assert(card.includes("truncated"), "truncated scope");

const { deriveColumns, truncateScope, tasksByColumn } = await import("../assets/kanban.js");
const bin = deriveColumns([{ status: "TO DO" }, { status: "DONE" }]);
assert(bin.length === 2 && bin[0] === "TO DO" && bin[1] === "DONE", "binary columns");
const multi = deriveColumns([{ status: "TO DO" }, { status: "IN PROGRESS" }]);
assert(multi.includes("IN PROGRESS"), "explicit status columns");
assert(truncateScope("x".repeat(100)).endsWith("…"), "scope truncation");
const grouped = tasksByColumn([{ card_id: "1", status: "DONE", is_blocked: false }]);
assert(grouped.map["DONE"].length === 1, "done column");

const editorModal = readFileSync(join(dir, "TaskEditorModal.vue"), "utf8");
assert(editorModal.includes("task-json"), "raw json textarea");
assert(editorModal.includes("save"), "save emit");
assert(editorModal.includes("publish"), "publish from task modal");

const split = readFileSync(join(dir, "SplitView.vue"), "utf8");
const appVue = readFileSync(join(dir, "../App.vue"), "utf8");
assert(split.includes("split-view"), "split view class");
assert(split.includes("row") && split.includes("col"), "row/col orientation");
assert(appVue.includes("Split View"), "split toggle");
assert(appVue.includes('name="columns-two"'), "split lucide icon");
assert(appVue.includes("PaneWorkspace"), "independent pane workspaces");
const { splitOrientation } = await import("../assets/split.js");
assert(splitOrientation(1200, 800) === "row", "wide viewport is row");
assert(splitOrientation(600, 900) === "col", "tall viewport is col");

const blocked = readFileSync(join(dir, "BlockedIndicator.vue"), "utf8");
const blockerModal = readFileSync(join(dir, "BlockerModal.vue"), "utf8");
assert(blocked.includes("blocked-indicator"), "blocked badge");
assert(blocked.includes("title"), "hover tooltip");
assert(blockerModal.includes("Blocked by"), "blocker modal");
const { columnForTask } = await import("../assets/kanban.js");
assert(columnForTask({ is_blocked: true, status: "IN PROGRESS" }) === "TO DO", "blocked stays in TO DO");

const remote = readFileSync(join(dir, "RemoteStatus.vue"), "utf8");
assert(remote.includes("refresh-icon"), "refresh icon");
assert(remote.includes("behind"), "behind prop");
assert(remote.includes("syncing"), "spinning feedback");
console.log("ok");
