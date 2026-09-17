<template>
  <div class="mermaid-block" ref="root"></div>
</template>

<script>
import { renderMermaid } from "../assets/mermaid-lite.js";

export default {
  name: "MermaidRenderer",
  props: {
    code: { type: String, required: true },
  },
  mounted() {
    this.draw();
  },
  watch: {
    code() {
      this.draw();
    },
  },
  methods: {
    draw() {
      const el = this.$refs.root;
      if (!el) return;
      el.innerHTML = renderMermaid(this.code);
    },
  },
};
</script>
