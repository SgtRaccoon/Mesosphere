import { markdownToHtml } from "./markdown.js";
import { tasksByColumn, truncateScope } from "./kanban.js";
import { splitOrientation } from "./split.js";
import { buildDocTree, childDirs } from "./docs-tree.js";

const app = document.getElementById("app");

function emptyPane() {
  return {
    repo: null,
    picking: true,
    tab: "docs",
    docs: [],
    tasks: [],
    activePath: "",
    content: "",
    editing: false,
    draft: "",
    versions: [],
    version: "",
    unpublished: false,
    behind: 0,
    syncing: false,
    activeTask: null,
    taskDraft: "",
    taskDirty: false,
    blockerTask: null,
    openFolders: {},
  };
}

const state = {
  repos: [],
  repoStats: {},
  addingRepo: false,
  addRepoError: "",
  split: false,
  panes: [emptyPane(), emptyPane()],
  publishing: false,
  publishError: "",
  pollTimer: null,
};

function icon(name) {
  const inner = {
    "columns-two": '<rect width="18" height="18" x="3" y="3" rx="2" /><path d="M12 3v18" />',
    "file-text": '<path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z" /><path d="M14 2v4a2 2 0 0 0 2 2h4" /><path d="M10 9H8" /><path d="M16 13H8" /><path d="M16 17H8" />',
    "list-todo": '<rect x="3" y="5" width="6" height="6" rx="1" /><path d="m3 17 2 2 4-4" /><path d="M13 6h8" /><path d="M13 12h8" /><path d="M13 18h8" />',
    upload:
      '<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" /><polyline points="17 8 12 3 7 8" /><line x1="12" x2="12" y1="3" y2="15" />',
    save: '<path d="M15.2 3a2 2 0 0 1 1.4.6l3.8 3.8a2 2 0 0 1 .6 1.4V19a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2z" /><path d="M17 21v-7a1 1 0 0 0-1-1H8a1 1 0 0 0-1 1v7" /><path d="M7 3v4a1 1 0 0 0 1 1h7" />',
  }[name];
  return `<svg class="icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${inner}</svg>`;
}

function escapeHtml(s) {
  return String(s)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

function paneCount() {
  return state.split ? 2 : 1;
}

function renderTree(node, paneIndex, depth) {
  const dirs = childDirs(node);
  let html = "";
  for (const dir of dirs) {
    const key = `${depth}:${dir.name}`;
    const open = !!state.panes[paneIndex].openFolders[key];
    html += `<div class="doc-tree-dir" style="padding-left:${depth ? 0.75 : 0}rem">
      <button type="button" class="docs-nav-folder" data-pane="${paneIndex}" data-folder="${escapeHtml(key)}">${open ? "▾" : "▸"} ${escapeHtml(dir.name)}</button>
      ${open ? renderTree(dir, paneIndex, depth + 1) : ""}
    </div>`;
  }
  for (const file of node.files || []) {
    const p = state.panes[paneIndex];
    html += `<button type="button" class="docs-nav-item${file.path === p.activePath ? " active" : ""}" data-pane="${paneIndex}" data-path="${escapeHtml(file.path)}" style="padding-left:${0.5 + depth * 0.75}rem">${icon("file-text")} ${escapeHtml(file.name || file.path)}</button>`;
  }
  return html;
}

function repoStatsLabel(id) {
  const st = state.repoStats[id] || {};
  const docs = st.docs != null ? st.docs : "…";
  const tasks = st.tasks != null ? st.tasks : "…";
  return `${docs} docs · ${tasks} tasks`;
}

function repoCards(paneIndex) {
  const cards = state.repos
    .map(
      (r) => `<button type="button" class="repo-card" data-pane="${paneIndex}" data-id="${escapeHtml(r.id)}">
            <strong>${escapeHtml(r.name || r.id)}</strong>
            <span>${escapeHtml(r.path || "")}</span>
            <span class="repo-card-meta">${escapeHtml(repoStatsLabel(r.id))}</span>
          </button>`
    )
    .join("");
  const add = state.addingRepo
    ? `<div class="repo-card add-repo">
          <strong>Add repository</strong>
          <form class="repo-add-form">
            <input name="name" placeholder="Name" required />
            <input name="path" placeholder="Filesystem path" required />
            <input name="id" placeholder="ID (optional)" />
            <input name="remote_url" placeholder="Remote URL (optional)" />
            <button type="submit">Save</button>
            <button type="button" class="repo-add-cancel">Cancel</button>
          </form>
          ${state.addRepoError ? `<p class="empty">${escapeHtml(state.addRepoError)}</p>` : ""}
        </div>`
    : `<button type="button" class="repo-card add-repo repo-add-open">
          <strong>+ Add repository</strong>
          <span>Register a local git repo</span>
        </button>`;
  return `<section class="repo-selector">
        <h1>Select a repository</h1>
        <div class="repo-grid">${cards}${add}</div>
        ${state.repos.length ? "" : '<p class="empty">No repositories configured.</p>'}
      </section>`;
}

function renderDocs(pane, i) {
  const tree = buildDocTree(pane.docs);
  const nav = renderTree(tree, i, 0);
  const html = pane.editing
    ? `<textarea class="docs-editor-textarea" data-pane="${i}">${escapeHtml(pane.draft)}</textarea>`
    : markdownToHtml(pane.content, pane.activePath);
  const fab = pane.activePath
    ? `<button type="button" class="docs-edit-fab" data-pane="${i}">${pane.editing ? icon("save") + " Save" : icon("file-text") + " Edit"}</button>`
    : "";
  const opts = (pane.versions || [])
    .map((v) => {
      const date = String(v.timestamp || "").slice(0, 10);
      const name = (v.message || v.hash || "commit").split("\n")[0];
      return `<option value="${escapeHtml(v.hash)}" ${pane.version === v.hash ? "selected" : ""}>${escapeHtml(`${name} (${date || "—"})`)}</option>`;
    })
    .join("");
  const bar = pane.activePath
    ? `<header class="docs-main-bar"><span class="docs-main-path">${escapeHtml(pane.activePath)}</span>
        <select class="docs-version" data-pane="${i}">
          <option value="" ${pane.version ? "" : "selected"}>HEAD</option>
          ${opts}
        </select></header>`
    : "";
  return `<div class="docs-view">
        <aside class="docs-nav">${nav}</aside>
        <div class="docs-main-wrap">${bar}<article class="docs-main">${html}</article></div>
        ${fab}
      </div>`;
}

function renderTasks(pane, i) {
  const { columns, map } = tasksByColumn(pane.tasks);
  let workspace = `<div class="kanban">${columns
    .map(
      (col) => `<section class="kanban-col" data-col="${escapeHtml(col)}">
          <h2>${escapeHtml(col)}</h2>
          ${(map[col] || [])
            .map((t) => {
              const tip = (t.blocker_ids || [])
                .map((id) => {
                  const b = pane.tasks.find((x) => x.card_id === id);
                  return b ? b.title || id : id;
                })
                .join(", ");
              const badge = t.is_blocked
                ? `<button type="button" class="blocked-indicator" title="${escapeHtml(tip || "Blocked")}" data-pane="${i}" data-blocked="${escapeHtml(t.card_id)}">⛔</button>`
                : "";
              return `<article class="task-card" data-pane="${i}" data-id="${escapeHtml(t.card_id)}">
                <header><span class="task-id">${escapeHtml(t.card_id)}</span>
                <span class="task-status">${escapeHtml(t.status || "")}</span></header>
                <p class="task-title">${escapeHtml(t.title || "")}</p>
                <p class="task-scope">${escapeHtml(truncateScope(t.scope))}</p>
                ${badge}
              </article>`;
            })
            .join("")}
        </section>`
    )
    .join("")}</div>`;
  if (pane.blockerTask) {
    const ids = pane.blockerTask.blocker_ids || [];
    const blockers = ids.map((id) => pane.tasks.find((x) => x.card_id === id) || { card_id: id, title: id });
    workspace += `<div class="blocker-modal" role="dialog"><h2>Blocked by</h2><ul>${blockers
      .map((b) => `<li><strong>${escapeHtml(b.card_id)}</strong> — ${escapeHtml(b.title || "")}</li>`)
      .join("")}</ul><button type="button" class="blocker-close" data-pane="${i}">Close</button></div>`;
  }
  if (pane.activeTask) {
    workspace += `<div class="task-editor-modal" role="dialog">
        <h2>${escapeHtml(pane.activeTask.card_id)}</h2>
        <textarea class="task-json" data-pane="${i}">${escapeHtml(pane.taskDraft)}</textarea>
        <footer>
          ${pane.taskDirty ? `<button type="button" class="task-save" data-pane="${i}">Save</button>` : ""}
          ${pane.unpublished ? `<button type="button" class="task-publish" data-pane="${i}">Publish</button>` : ""}
          <button type="button" class="task-close" data-pane="${i}">Close</button>
        </footer>
      </div>`;
  }
  return workspace;
}

function renderPane(pane, i) {
  if (!pane.repo || pane.picking) return repoCards(i);
  const name = escapeHtml(pane.repo.name || pane.repo.id);
  const publishBtn = pane.unpublished
    ? `<button type="button" class="publish-btn" data-pane="${i}" ${state.publishing ? "disabled" : ""}>${icon("upload")} ${state.publishing ? "Publishing…" : "Publish"}</button>`
    : "";
  const refresh =
    pane.behind > 0
      ? `<button type="button" class="refresh-icon${pane.syncing ? " spinning" : ""}" data-pane="${i}" ${pane.syncing ? "disabled" : ""} title="Remote changes available">↻</button>`
      : "";
  const bar = `<header class="pane-bar">
      ${refresh}
      <button type="button" class="pane-bar-repo" data-pane="${i}">${name}</button>
      <nav class="pane-bar-tabs">
        <button type="button" data-pane="${i}" data-tab="docs" class="${pane.tab === "docs" ? "active" : ""}">${icon("file-text")} Docs</button>
        <button type="button" data-pane="${i}" data-tab="tasks" class="${pane.tab === "tasks" ? "active" : ""}">${icon("list-todo")} Tasks</button>
      </nav>
      ${publishBtn}
    </header>`;
  const body = pane.tab === "docs" ? renderDocs(pane, i) : renderTasks(pane, i);
  return `${bar}${body}`;
}

async function render() {
  if (!app) return;
  const n = paneCount();
  let workspace = "";
  if (n === 1) {
    workspace = renderPane(state.panes[0], 0);
  } else {
    const orient = splitOrientation(window.innerWidth || 800, window.innerHeight || 600);
    workspace = `<div class="split-view ${orient}">
      <div class="pane" data-pane="0">${renderPane(state.panes[0], 0)}</div>
      <div class="pane" data-pane="1">${renderPane(state.panes[1], 1)}</div>
    </div>`;
  }
  const modal = state.publishError
    ? `<div class="publish-modal" role="alertdialog"><p>${escapeHtml(state.publishError)}</p><button type="button" class="publish-dismiss">Close</button></div>`
    : "";
  app.innerHTML = `<header class="top-bar">
        <div class="top-bar-left">Mesosphere</div>
        <nav class="top-bar-right">
          <button type="button" class="split-toggle">${icon("columns-two")} Split View</button>
        </nav>
      </header>${workspace}${modal}`;

  bindEvents();
}

function paneFromEl(el) {
  const i = Number(el.dataset.pane || 0);
  return state.panes[i];
}

function bindEvents() {
  const addOpen = app.querySelector(".repo-add-open");
  if (addOpen) {
    addOpen.addEventListener("click", () => {
      state.addingRepo = true;
      state.addRepoError = "";
      render();
    });
  }
  const addCancel = app.querySelector(".repo-add-cancel");
  if (addCancel) {
    addCancel.addEventListener("click", () => {
      state.addingRepo = false;
      state.addRepoError = "";
      render();
    });
  }
  const addForm = app.querySelector(".repo-add-form");
  if (addForm) {
    addForm.addEventListener("submit", async (ev) => {
      ev.preventDefault();
      const fd = new FormData(addForm);
      await addRepo({
        name: String(fd.get("name") || ""),
        path: String(fd.get("path") || ""),
        id: String(fd.get("id") || ""),
        remote_url: String(fd.get("remote_url") || ""),
      });
    });
  }
  app.querySelectorAll(".repo-card[data-id]").forEach((btn) => {
    btn.addEventListener("click", async () => {
      const i = Number(btn.dataset.pane || 0);
      const pane = state.panes[i];
      pane.repo = state.repos.find((r) => r.id === btn.dataset.id) || null;
      pane.picking = false;
      pane.docs = [];
      pane.activePath = "";
      pane.content = "";
      if (pane.repo) {
        await loadDocs(pane);
        await loadTasks(pane);
        await refreshStatus(pane);
        startStatusPoll();
      }
      render();
    });
  });
  app.querySelectorAll(".pane-bar-repo").forEach((btn) => {
    btn.addEventListener("click", () => {
      paneFromEl(btn).picking = true;
      render();
    });
  });
  app.querySelectorAll(".pane-bar [data-tab]").forEach((btn) => {
    btn.addEventListener("click", async () => {
      const pane = paneFromEl(btn);
      pane.tab = btn.dataset.tab;
      if (pane.tab === "tasks") await loadTasks(pane);
      render();
    });
  });
  app.querySelectorAll(".docs-nav-folder").forEach((btn) => {
    btn.addEventListener("click", () => {
      const pane = paneFromEl(btn);
      const key = btn.dataset.folder;
      pane.openFolders[key] = !pane.openFolders[key];
      render();
    });
  });
  app.querySelectorAll(".docs-nav-item").forEach((btn) => {
    btn.addEventListener("click", async () => {
      await selectDoc(paneFromEl(btn), btn.dataset.path);
      render();
    });
  });
  app.querySelectorAll(".docs-version").forEach((sel) => {
    sel.addEventListener("change", async () => {
      const pane = paneFromEl(sel);
      pane.version = sel.value;
      await loadDocContent(pane);
      render();
    });
  });
  app.querySelectorAll(".docs-edit-fab").forEach((fab) => {
    fab.addEventListener("click", async () => {
      const pane = paneFromEl(fab);
      if (pane.editing) await saveDoc(pane);
      else {
        pane.editing = true;
        pane.draft = pane.content;
      }
      render();
    });
  });
  app.querySelectorAll(".docs-editor-textarea").forEach((ta) => {
    ta.addEventListener("input", () => {
      paneFromEl(ta).draft = ta.value;
    });
  });
  app.querySelectorAll(".publish-btn").forEach((pub) => {
    pub.addEventListener("click", () => publishRepo(paneFromEl(pub)));
  });
  app.querySelectorAll(".blocked-indicator").forEach((el) => {
    el.addEventListener("click", (ev) => {
      ev.stopPropagation();
      const pane = paneFromEl(el);
      pane.blockerTask = pane.tasks.find((x) => x.card_id === el.dataset.blocked) || null;
      render();
    });
  });
  app.querySelectorAll(".blocker-close").forEach((el) => {
    el.addEventListener("click", () => {
      paneFromEl(el).blockerTask = null;
      render();
    });
  });
  app.querySelectorAll(".task-card").forEach((el) => {
    el.addEventListener("click", () => {
      const pane = paneFromEl(el);
      const t = pane.tasks.find((x) => x.card_id === el.dataset.id);
      if (!t) return;
      pane.activeTask = t;
      pane.taskDraft = taskEditorJSON(t);
      pane.taskDirty = false;
      render();
    });
  });
  app.querySelectorAll(".task-json").forEach((taskJson) => {
    taskJson.addEventListener("input", () => {
      const pane = paneFromEl(taskJson);
      pane.taskDraft = taskJson.value;
      pane.taskDirty = true;
      if (!app.querySelector(".task-save")) render();
    });
  });
  app.querySelectorAll(".task-save").forEach((el) => el.addEventListener("click", () => saveTask(paneFromEl(el))));
  app.querySelectorAll(".task-publish").forEach((el) => el.addEventListener("click", () => publishRepo(paneFromEl(el))));
  app.querySelectorAll(".task-close").forEach((el) => {
    el.addEventListener("click", () => {
      paneFromEl(el).activeTask = null;
      render();
    });
  });
  app.querySelectorAll(".refresh-icon").forEach((el) => {
    el.addEventListener("click", () => syncRemote(paneFromEl(el)));
  });
  const splitBtn = app.querySelector(".split-toggle");
  if (splitBtn) {
    splitBtn.addEventListener("click", () => {
      state.split = !state.split;
      render();
    });
  }
  const dismiss = app.querySelector(".publish-dismiss");
  if (dismiss) {
    dismiss.addEventListener("click", () => {
      state.publishError = "";
      render();
    });
  }
}

async function loadTasks(pane) {
  if (!pane.repo) return;
  const res = await fetch(`/api/v1/repos/${pane.repo.id}/tasks`);
  if (!res.ok) return;
  pane.tasks = await res.json();
}

async function loadDocs(pane) {
  const res = await fetch(`/api/v1/repos/${pane.repo.id}/docs`);
  if (!res.ok) return;
  pane.docs = await res.json();
}

async function selectDoc(pane, path) {
  pane.activePath = path;
  pane.version = "";
  pane.editing = false;
  const q = new URLSearchParams({ path });
  const vres = await fetch(`/api/v1/repos/${pane.repo.id}/docs/versions?${q}`);
  pane.versions = vres.ok ? await vres.json() : [];
  await loadDocContent(pane);
}

async function loadDocContent(pane) {
  const q = new URLSearchParams({ path: pane.activePath });
  if (pane.version) q.set("version", pane.version);
  const res = await fetch(`/api/v1/repos/${pane.repo.id}/docs/detail?${q}`);
  if (!res.ok) return;
  const body = await res.json();
  pane.content = body.content || "";
  pane.draft = pane.content;
  pane.editing = false;
}

async function saveDoc(pane) {
  const res = await fetch(`/api/v1/repos/${pane.repo.id}/docs/save`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ path: pane.activePath, content: pane.draft }),
  });
  if (!res.ok) return;
  pane.content = pane.draft;
  pane.editing = false;
  pane.unpublished = true;
}

async function publishRepo(pane) {
  state.publishing = true;
  state.publishError = "";
  await render();
  const res = await fetch(`/api/v1/repos/${pane.repo.id}/publish`, { method: "POST" });
  state.publishing = false;
  if (res.status === 409) {
    const body = await res.json().catch(() => ({}));
    state.publishError = body.error || "Git conflict detected. Please resolve manually in repo.";
  } else if (res.ok) {
    pane.unpublished = false;
  } else {
    state.publishError = "Publish failed";
  }
  await render();
}

async function refreshStatus(pane) {
  if (!pane.repo) return;
  const res = await fetch(`/api/v1/repos/${pane.repo.id}/status`);
  if (!res.ok) return;
  const st = await res.json();
  pane.unpublished = !!st.ahead;
  pane.behind = st.behind || 0;
}

function startStatusPoll() {
  if (state.pollTimer) clearInterval(state.pollTimer);
  state.pollTimer = setInterval(async () => {
    for (let i = 0; i < paneCount(); i++) {
      if (state.panes[i].repo) await refreshStatus(state.panes[i]);
    }
    render();
  }, 15000);
}

async function syncRemote(pane) {
  if (!pane.repo || pane.syncing) return;
  pane.syncing = true;
  await render();
  const res = await fetch(`/api/v1/repos/${pane.repo.id}/sync`, { method: "POST" });
  pane.syncing = false;
  if (res.ok) {
    pane.behind = 0;
    await loadDocs(pane);
    await loadTasks(pane);
    if (pane.activePath) await loadDocContent(pane);
  }
  await refreshStatus(pane);
  await render();
}

async function saveTask(pane) {
  if (!pane.repo || !pane.activeTask) return;
  const res = await fetch(`/api/v1/repos/${pane.repo.id}/tasks/${pane.activeTask.card_id}/save`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ raw_json: pane.taskDraft }),
  });
  if (!res.ok) return;
  pane.taskDirty = false;
  pane.unpublished = true;
  await loadTasks(pane);
  await render();
}

function taskEditorJSON(t) {
  if (!t) return "{}";
  const extra = t.extra_fields;
  if (extra && typeof extra === "object" && Object.keys(extra).length) {
    return JSON.stringify(extra, null, 2);
  }
  return JSON.stringify(
    {
      card_id: t.card_id,
      title: t.title,
      scope: t.scope,
      validation_gate: t.validation_gate,
      is_blocked: t.is_blocked,
      blocker_ids: t.blocker_ids || [],
      status: t.status,
    },
    null,
    2,
  );
}

async function addRepo(body) {
  state.addRepoError = "";
  const res = await fetch("/api/v1/repos", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    state.addRepoError = err.error || "Could not add repository";
    await render();
    return;
  }
  state.addingRepo = false;
  await load();
}

async function loadRepoStats() {
  await Promise.all(
    (state.repos || []).map(async (r) => {
      const [docsRes, tasksRes] = await Promise.all([
        fetch(`/api/v1/repos/${r.id}/docs`),
        fetch(`/api/v1/repos/${r.id}/tasks`),
      ]);
      const docs = docsRes.ok ? await docsRes.json() : [];
      const tasks = tasksRes.ok ? await tasksRes.json() : [];
      state.repoStats[r.id] = { docs: docs.length || 0, tasks: tasks.length || 0 };
    }),
  );
}

async function load() {
  try {
    const res = await fetch("/api/v1/repos");
    if (res.ok) state.repos = await res.json();
  } catch (_) {
    state.repos = [];
  }
  await loadRepoStats();
  await render();
}

window.__mesosphereUI = { state, render, load, markdownToHtml, saveDoc, publishRepo, saveTask, syncRemote };
load();
