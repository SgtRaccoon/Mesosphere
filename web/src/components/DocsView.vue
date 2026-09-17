<template>
  <div class="docs-view">
    <aside class="docs-nav">
      <button
        v-for="doc in docs"
        :key="doc.path"
        type="button"
        class="docs-nav-item"
        :class="{ active: doc.path === activePath }"
        @click="select(doc)"
      >
        {{ doc.path }}
      </button>
    </aside>
    <article class="docs-main" v-html="html"></article>
  </div>
</template>

<script>
import { markdownToHtml } from "../assets/markdown.js";

export default {
  name: "DocsView",
  props: {
    repoId: { type: String, required: true },
    docs: { type: Array, default: () => [] },
  },
  data() {
    return { activePath: "", content: "" };
  },
  computed: {
    html() {
      return markdownToHtml(this.content, this.activePath);
    },
  },
  methods: {
    async select(doc) {
      this.activePath = doc.path;
      const q = new URLSearchParams({ path: doc.path });
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
.docs-nav-item {
  display: block;
  width: 100%;
  text-align: left;
  background: transparent;
  color: inherit;
  border: 0;
  padding: 0.4rem 0.5rem;
  cursor: pointer;
}
.docs-nav-item.active {
  font-weight: 700;
}
.docs-main {
  padding: 1rem 1.5rem;
}
</style>
