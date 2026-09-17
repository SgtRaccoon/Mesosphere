<template>
  <header class="pane-bar" v-if="repo">
    <button type="button" class="pane-bar-repo" @click="$emit('open-repos')">{{ repo.name || repo.id }}</button>
    <nav class="pane-bar-tabs">
      <button type="button" :class="{ active: tab === 'docs' }" @click="$emit('update:tab', 'docs')">Docs</button>
      <button type="button" :class="{ active: tab === 'tasks' }" @click="$emit('update:tab', 'tasks')">Tasks</button>
    </nav>
  </header>
</template>

<script>
export default {
  name: "PaneBar",
  props: {
    repo: { type: Object, default: null },
    tab: { type: String, default: "docs" },
  },
  emits: ["update:tab", "open-repos"],
};
</script>

<style scoped>
.pane-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0.75rem;
  border-bottom: 1px solid #2a3644;
}
.pane-bar-repo {
  font-weight: 600;
  background: transparent;
  color: inherit;
  border: 0;
  cursor: pointer;
}
.pane-bar-tabs {
  display: flex;
  gap: 0.5rem;
}
.pane-bar-tabs button.active {
  font-weight: 700;
}
</style>
