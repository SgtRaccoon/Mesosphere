import { markdownToHtml } from "./markdown.js";
import { tasksByColumn, truncateScope } from "./kanban.js";
import { splitOrientation } from "./split.js";

const app = document.getElementById("app");

const state = {
  repos: [],
  selected: null,
  tab: "docs",
  docs: [],
  activePath: "",
  content: "",
  editing: false,
  draft: "",
  unpublished: false,
  publishing: false,
  publishError: "",
  tasks: [],
  activeTask: null,
  taskDraft: "",
  taskDirty: false,
  split: false,
  blockerTask: null,
  behind: 0,
  syncing: false,
  pollTimer: null,
};

function escapeHtml(s) {
  return String(s)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

async function render() {
  if (!app) return;
  if (!state.selected) {
    const cards = state.repos
      .map(
        (r) => `<button type="button" class="repo-card" data-id="${escapeHtml(r.id)}">
            <strong>${escapeHtml(r.name || r.id)}</strong>
            <span>${escapeHtml(r.path || "")}</span>
          </button>`
      )
      .join("");
    app.innerHTML = `<section class="repo-selector">
        <h1>Select a repository</h1>
        <div class="repo-grid">${cards || ""}</div>
        ${state.repos.length ? "" : '<p class="empty">No repositories configured.</p>'}
      </section>`;
    app.querySelectorAll(".repo-card").forEach((btn) => {
      btn.addEventListener("click", async () => {
        state.selected = state.repos.find((r) => r.id === btn.dataset.id) || null;
        state.docs = [];
        state.activePath = "";
        state.content = "";
        if (state.selected) {
          await loadDocs();
          await loadTasks();
          await refreshStatus();
          startStatusPoll();
        }
        render();
      });
    });
    return;
  }
  const name = escapeHtml(state.selected.name || state.selected.id);
  let workspace = "";
  if (state.tab === "docs") {
    const nav = state.docs
      .map(
        (d) =>
          `<button type="button" class="docs-nav-item${d.path === state.activePath ? " active" : ""}" data-path="${escapeHtml(d.path)}">${escapeHtml(d.path)}</button>`
      )
      .join("");
    const html = state.editing
      ? `<textarea class="docs-editor-textarea">${escapeHtml(state.draft)}</textarea>`
      : markdownToHtml(state.content, state.activePath);
    const fab = state.activePath
      ? `<button type="button" class="docs-edit-fab">${state.editing ? "Save" : "✎"}</button>`
      : "";
    workspace = `<div class="docs-view">
        <aside class="docs-nav">${nav}</aside>
        <article class="docs-main">${html}</article>
        ${fab}
      </div>`;
  } else {
    const { columns, map } = tasksByColumn(state.tasks);
    workspace = `<div class="kanban">${columns
      .map(
        (col) => `<section class="kanban-col" data-col="${escapeHtml(col)}">
          <h2>${escapeHtml(col)}</h2>
          ${(map[col] || [])
            .map(
              (t) => {
                const tip = (t.blocker_ids || [])
                  .map((id) => {
                    const b = state.tasks.find((x) => x.card_id === id);
                    return b ? b.title || id : id;
                  })
                  .join(", ");
                const badge = t.is_blocked
                  ? `<button type="button" class="blocked-indicator" title="${escapeHtml(tip || "Blocked")}" data-blocked="${escapeHtml(t.card_id)}">⛔</button>`
                  : "";
                return `<article class="task-card" data-id="${escapeHtml(t.card_id)}">
                <header><span class="task-id">${escapeHtml(t.card_id)}</span>
                <span class="task-status">${escapeHtml(t.status || "")}</span></header>
                <p class="task-title">${escapeHtml(t.title || "")}</p>
                <p class="task-scope">${escapeHtml(truncateScope(t.scope))}</p>
                ${badge}
              </article>`;
              }
            )
            .join("")}
        </section>`
      )
      .join("")}</div>`;
    if (state.blockerTask) {
      const ids = state.blockerTask.blocker_ids || [];
      const blockers = ids.map((id) => state.tasks.find((x) => x.card_id === id) || { card_id: id, title: id });
      workspace += `<div class="blocker-modal" role="dialog"><h2>Blocked by</h2><ul>${blockers
        .map((b) => `<li><strong>${escapeHtml(b.card_id)}</strong> — ${escapeHtml(b.title || "")}</li>`)
        .join("")}</ul><button type="button" class="blocker-close">Close</button></div>`;
    }
    if (state.activeTask) {
      workspace += `<div class="task-editor-modal" role="dialog">
        <h2>${escapeHtml(state.activeTask.card_id)}</h2>
        <textarea class="task-json">${escapeHtml(state.taskDraft)}</textarea>
        <footer>
          ${state.taskDirty ? `<button type="button" class="task-save">Save</button>` : ""}
          ${state.unpublished ? `<button type="button" class="task-publish">Publish</button>` : ""}
          <button type="button" class="task-close">Close</button>
        </footer>
      </div>`;
    }
  }
  if (state.split) {
    const orient = splitOrientation(window.innerWidth || 800, window.innerHeight || 600);
    workspace = `<div class="split-view ${orient}"><div class="pane" data-pane="0">${workspace}</div><div class="pane" data-pane="1">${workspace}</div></div>`;
  }
  const publishBtn = state.unpublished
    ? `<button type="button" class="publish-btn" ${state.publishing ? "disabled" : ""}>${state.publishing ? "Publishing…" : "Publish"}</button>`
    : "";
  const modal = state.publishError
    ? `<div class="publish-modal" role="alertdialog"><p>${escapeHtml(state.publishError)}</p><button type="button" class="publish-dismiss">Close</button></div>`
    : "";
  const refresh =
    state.behind > 0
      ? `<button type="button" class="refresh-icon${state.syncing ? " spinning" : ""}" ${state.syncing ? "disabled" : ""} title="Remote changes available">↻</button>`
      : "";
  app.innerHTML = `<header class="top-bar">
        ${refresh}
        <div class="top-bar-left">${name}</div>
        <nav class="top-bar-right">
          <button type="button" data-tab="docs" class="${state.tab === "docs" ? "active" : ""}">Docs</button>
          <button type="button" data-tab="tasks" class="${state.tab === "tasks" ? "active" : ""}">Tasks</button>
          <button type="button" class="split-toggle">Split View</button>
          ${publishBtn}
        </nav>
      </header>${workspace}${modal}`;
  app.querySelectorAll(".top-bar [data-tab]").forEach((btn) => {
    btn.addEventListener("click", async () => {
      state.tab = btn.dataset.tab;
      if (state.tab === "tasks") await loadTasks();
      render();
    });
  });
  app.querySelectorAll(".docs-nav-item").forEach((btn) => {
    btn.addEventListener("click", async () => {
      await selectDoc(btn.dataset.path);
      render();
    });
  });
  const fab = app.querySelector(".docs-edit-fab");
  if (fab) {
    fab.addEventListener("click", async () => {
      if (state.editing) await saveDoc();
      else {
        state.editing = true;
        state.draft = state.content;
      }
      render();
    });
  }
  const ta = app.querySelector(".docs-editor-textarea");
  if (ta) {
    ta.addEventListener("input", () => {
      state.draft = ta.value;
    });
  }
  const pub = app.querySelector(".publish-btn");
  if (pub) pub.addEventListener("click", () => publishRepo());
  app.querySelectorAll(".blocked-indicator").forEach((el) => {
    el.addEventListener("click", (ev) => {
      ev.stopPropagation();
      state.blockerTask = state.tasks.find((x) => x.card_id === el.dataset.blocked) || null;
      render();
    });
  });
  const blockerClose = app.querySelector(".blocker-close");
  if (blockerClose) {
    blockerClose.addEventListener("click", () => {
      state.blockerTask = null;
      render();
    });
  }
  app.querySelectorAll(".task-card").forEach((el) => {
    el.addEventListener("click", () => {
      const t = state.tasks.find((x) => x.card_id === el.dataset.id);
      if (!t) return;
      state.activeTask = t;
      try {
        state.taskDraft = JSON.stringify(JSON.parse(t.raw_content || "{}"), null, 2);
      } catch {
        state.taskDraft = t.raw_content || "{}";
      }
      state.taskDirty = false;
      render();
    });
  });
  const taskJson = app.querySelector(".task-json");
  if (taskJson) {
    taskJson.addEventListener("input", () => {
      state.taskDraft = taskJson.value;
      state.taskDirty = true;
      const saveBtn = app.querySelector(".task-save");
      if (!saveBtn) render();
    });
  }
  const taskSave = app.querySelector(".task-save");
  if (taskSave) taskSave.addEventListener("click", () => saveTask());
  const taskPub = app.querySelector(".task-publish");
  if (taskPub) taskPub.addEventListener("click", () => publishRepo());
  const taskClose = app.querySelector(".task-close");
  if (taskClose) {
    taskClose.addEventListener("click", () => {
      state.activeTask = null;
      render();
    });
  }
  const refreshBtn = app.querySelector(".refresh-icon");
  if (refreshBtn) {
    refreshBtn.addEventListener("click", () => syncRemote());
  }
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

async function loadTasks() {
  if (!state.selected) return;
  const res = await fetch(`/api/v1/repos/${state.selected.id}/tasks`);
  if (!res.ok) return;
  state.tasks = await res.json();
}

async function loadDocs() {
  const res = await fetch(`/api/v1/repos/${state.selected.id}/docs`);
  if (!res.ok) return;
  state.docs = await res.json();
}

async function selectDoc(path) {
  state.activePath = path;
  const q = new URLSearchParams({ path });
  const res = await fetch(`/api/v1/repos/${state.selected.id}/docs/detail?${q}`);
  if (!res.ok) return;
  const body = await res.json();
  state.content = body.content || "";
  state.editing = false;
  state.draft = state.content;
}

async function saveDoc() {
  const res = await fetch(`/api/v1/repos/${state.selected.id}/docs/save`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ path: state.activePath, content: state.draft }),
  });
  if (!res.ok) return;
  state.content = state.draft;
  state.editing = false;
  state.unpublished = true;
}

async function publishRepo() {
  state.publishing = true;
  state.publishError = "";
  await render();
  const res = await fetch(`/api/v1/repos/${state.selected.id}/publish`, { method: "POST" });
  state.publishing = false;
  if (res.status === 409) {
    const body = await res.json().catch(() => ({}));
    state.publishError = body.error || "Git conflict detected. Please resolve manually in repo.";
  } else if (res.ok) {
    state.unpublished = false;
  } else {
    state.publishError = "Publish failed";
  }
  await render();
}

async function refreshStatus() {
  if (!state.selected) return;
  const res = await fetch(`/api/v1/repos/${state.selected.id}/status`);
  if (!res.ok) return;
  const st = await res.json();
  state.unpublished = !!st.ahead;
  state.behind = st.behind || 0;
}

function startStatusPoll() {
  if (state.pollTimer) clearInterval(state.pollTimer);
  state.pollTimer = setInterval(async () => {
    if (!state.selected) return;
    await refreshStatus();
    render();
  }, 15000);
}

async function syncRemote() {
  if (!state.selected || state.syncing) return;
  state.syncing = true;
  await render();
  const res = await fetch(`/api/v1/repos/${state.selected.id}/sync`, { method: "POST" });
  state.syncing = false;
  if (res.ok) {
    state.behind = 0;
    await loadDocs();
    await loadTasks();
    if (state.activePath) await selectDoc(state.activePath);
  }
  await refreshStatus();
  await render();
}

async function saveTask() {
  if (!state.selected || !state.activeTask) return;
  const res = await fetch(`/api/v1/repos/${state.selected.id}/tasks/${state.activeTask.card_id}/save`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ raw_json: state.taskDraft }),
  });
  if (!res.ok) return;
  state.taskDirty = false;
  state.unpublished = true;
  await loadTasks();
  await render();
}

async function load() {
  try {
    const res = await fetch("/api/v1/repos");
    if (res.ok) state.repos = await res.json();
  } catch (_) {
    state.repos = [];
  }
  await render();
}

window.__mesosphereUI = { state, render, load, markdownToHtml, saveDoc, publishRepo, saveTask, syncRemote };
load();
