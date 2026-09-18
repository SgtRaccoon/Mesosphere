<template>
  <div class="docs-view">
    <aside class="docs-nav">
      <button type="button" class="docs-new" @click="newDoc">
        <Icon name="file-plus-corner" /> New Document
      </button>
      <DocTreeNode :node="tree" :active-path="activePath" :depth="0" @select="select" />
    </aside>
    <div class="docs-main-wrap">
      <header class="docs-main-bar" v-if="activePath">
        <span class="docs-main-path">{{ activePath }}</span>
        <label class="docs-version-wrap" v-if="versions.length > 1">
          <select class="docs-version" v-model="version" @change="loadContent">
            <option value="">HEAD</option>
            <option v-for="v in versions" :key="v.hash" :value="v.hash">
              {{ versionLabel(v) }}
            </option>
          </select>
          <Icon name="chevron-down" />
        </label>
      </header>
      <article class="docs-main" v-html="html"></article>
    </div>
  </div>
</template>

<script>
import { markdownToHtml } from "../assets/markdown.js";
import { buildDocTree } from "../assets/docs-tree.js";
import DocTreeNode from "./DocTreeNode.vue";
import Icon from "./Icon.vue";

export default {
  name: "DocsView",
  components: { DocTreeNode, Icon },
  props: {
    repoId: { type: String, required: true },
    docs: { type: Array, default: () => [] },
  },
  data() {
    return { activePath: "", content: "", versions: [], version: "" };
  },
  computed: {
    html() {
      return markdownToHtml(this.content, this.activePath);
    },
    tree() {
      return buildDocTree(this.docs);
    },
  },
  methods: {
    versionLabel(v) {
      const date = (v.timestamp || "").slice(0, 10);
      const name = (v.message || v.hash || "").split("\n")[0];
      return `${name || "commit"} (${date || "—"})`;
    },
    newDoc() {
      const path = window.prompt("New document path", "docs/untitled.md");
      if (!path) return;
      this.activePath = path.trim();
      this.version = "";
      this.versions = [];
      this.content = "# Untitled\n\n";
    },
    async select(doc) {
      this.activePath = doc.path;
      this.version = "";
      await this.loadVersions();
      await this.loadContent();
    },
    async loadVersions() {
      const q = new URLSearchParams({ path: this.activePath });
      const res = await fetch(`/api/v1/repos/${this.repoId}/docs/versions?${q}`);
      if (!res.ok) {
        this.versions = [];
        return;
      }
      this.versions = await res.json();
    },
    async loadContent() {
      const q = new URLSearchParams({ path: this.activePath });
      if (this.version) q.set("version", this.version);
      const res = await fetch(`/api/v1/repos/${this.repoId}/docs/detail?${q}`);
      if (!res.ok) return;
      const body = await res.json();
      this.content = body.content || "";
    },
  },
};
</script>

<style scoped>
.docs-view {
  display: grid;
  grid-template-columns: 240px 1fr;
  min-height: 0;
  height: 100%;
}
.docs-nav {
  border-right: 1px solid #2a3644;
  padding: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.docs-new {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  width: 100%;
  box-sizing: border-box;
  background: #0f1419;
  color: inherit;
  border: 0;
  border-radius: 6px;
  padding: 0.45rem 0.65rem;
  cursor: pointer;
  transition: background 0.12s ease, color 0.12s ease;
}
.docs-new:hover {
  background: #1a222c;
  color: #7aa2f7;
}
.docs-main-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 1.5rem;
  border-bottom: 1px solid #2a3644;
}
.docs-version-wrap {
  margin-left: auto;
  position: relative;
  display: inline-flex;
  align-items: center;
}
.docs-version-wrap .icon {
  position: absolute;
  right: 0.35rem;
  pointer-events: none;
}
.docs-version {
  text-align: right;
  text-align-last: right;
  color: inherit;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 6px;
  padding: 0.25rem 1.5rem 0.25rem 0.4rem;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
  transition: border-color 0.15s ease, background 0.15s ease;
}
.docs-version:hover,
.docs-version:focus {
  background: #1a222c;
  border-color: #2a3644;
  outline: none;
}
.docs-main {
  padding: 1rem 1.5rem;
}
</style>
