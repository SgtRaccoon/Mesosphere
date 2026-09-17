<template>
  <div class="docs-view">
    <aside class="docs-nav">
      <DocTreeNode :node="tree" :active-path="activePath" :depth="0" @select="select" />
    </aside>
    <div class="docs-main-wrap">
      <header class="docs-main-bar" v-if="activePath">
        <span class="docs-main-path">{{ activePath }}</span>
        <select class="docs-version" v-model="version" @change="loadContent">
          <option value="">HEAD</option>
          <option v-for="v in versions" :key="v.hash" :value="v.hash">
            {{ versionLabel(v) }}
          </option>
        </select>
      </header>
      <article class="docs-main" v-html="html"></article>
    </div>
  </div>
</template>

<script>
import { markdownToHtml } from "../assets/markdown.js";
import { buildDocTree } from "../assets/docs-tree.js";
import DocTreeNode from "./DocTreeNode.vue";

export default {
  name: "DocsView",
  components: { DocTreeNode },
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
  min-height: 60vh;
}
.docs-nav {
  border-right: 1px solid #2a3644;
  padding: 0.5rem;
}
.docs-main-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 1.5rem;
  border-bottom: 1px solid #2a3644;
}
.docs-version {
  margin-left: auto;
}
.docs-main {
  padding: 1rem 1.5rem;
}
</style>
