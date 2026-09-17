<template>
  <header class="top-bar" v-if="repo">
    <div class="top-bar-left">{{ repo.name || repo.id }}</div>
    <nav class="top-bar-right">
      <button type="button" :class="{ active: tab === 'docs' }" @click="$emit('update:tab', 'docs')">Docs</button>
      <button type="button" :class="{ active: tab === 'tasks' }" @click="$emit('update:tab', 'tasks')">Tasks</button>
    </nav>
  </header>
</template>

<script>
export default {
  name: "TopBar",
  props: {
    repo: { type: Object, default: null },
    tab: { type: String, default: "docs" },
  },
  emits: ["update:tab"],
};
</script>

<style scoped>
.top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #2a3644;
}
.top-bar-left {
  font-weight: 600;
}
.top-bar-right {
  display: flex;
  gap: 0.5rem;
}
.top-bar-right button.active {
  font-weight: 700;
}
</style>
