<template>
  <header class="pane-bar" v-if="repo">
    <button type="button" class="pane-bar-repo" @click="$emit('open-repos')">{{ repo.name || repo.id }}</button>
    <nav class="pane-bar-tabs">
      <button type="button" :class="{ active: tab === 'docs' }" @click="$emit('update:tab', 'docs')">
        <Icon name="file-text" /> Docs
      </button>
      <button type="button" :class="{ active: tab === 'tasks' }" @click="$emit('update:tab', 'tasks')">
        <Icon name="list-todo" /> Tasks
      </button>
    </nav>
  </header>
</template>

<script>
import Icon from "./Icon.vue";

export default {
  name: "PaneBar",
  components: { Icon },
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
  transition: color 0.12s ease;
}
.pane-bar-repo:hover {
  color: #7aa2f7;
}
.pane-bar-tabs {
  display: flex;
  gap: 0.15rem;
  align-items: center;
  background: #1a222c;
  border-radius: 8px;
  padding: 0.15rem;
}
.pane-bar-tabs button {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background: #1a222c;
  color: inherit;
  border: 0;
  border-radius: 6px;
  padding: 0.3rem 0.65rem;
  cursor: pointer;
  transition: color 0.12s ease, background 0.12s ease;
}
.pane-bar-tabs button:hover:not(.active) {
  color: #7aa2f7;
}
.pane-bar-tabs button.active {
  font-weight: 700;
  background: #0f1419;
  color: #e7ecf1;
}
</style>
