<template>
  <div class="doc-tree" :style="{ paddingLeft: depth ? '0.75rem' : '0' }">
    <div v-for="dir in dirs" :key="dir.name" class="doc-tree-dir">
      <button type="button" class="docs-nav-folder" @click="toggle(dir.name)">
        {{ open[dir.name] ? "▾" : "▸" }} {{ dir.name }}
      </button>
      <DocTreeNode v-if="open[dir.name]" :node="dir" :active-path="activePath" :depth="depth + 1" @select="$emit('select', $event)" />
    </div>
    <button
      v-for="file in node.files"
      :key="file.path"
      type="button"
      class="docs-nav-item"
      :class="{ active: file.path === activePath }"
      @click="$emit('select', file)"
    >
      {{ file.name || file.path }}
    </button>
  </div>
</template>

<script>
import { childDirs } from "../assets/docs-tree.js";

export default {
  name: "DocTreeNode",
  props: {
    node: { type: Object, required: true },
    activePath: { type: String, default: "" },
    depth: { type: Number, default: 0 },
  },
  emits: ["select"],
  data() {
    return { open: {} };
  },
  computed: {
    dirs() {
      return childDirs(this.node);
    },
  },
  methods: {
    toggle(name) {
      this.open = { ...this.open, [name]: !this.open[name] };
    },
  },
};
</script>
