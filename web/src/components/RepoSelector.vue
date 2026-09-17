<template>
  <section class="repo-selector" v-if="!modelValue">
    <h1>Select a repository</h1>
    <div class="repo-grid">
      <button
        v-for="repo in repos"
        :key="repo.id"
        class="repo-card"
        type="button"
        @click="$emit('update:modelValue', repo)"
      >
        <strong>{{ repo.name || repo.id }}</strong>
        <span>{{ repo.path }}</span>
        <span class="repo-card-meta">{{ statsLabel(repo) }}</span>
      </button>
    </div>
    <p v-if="!repos.length" class="empty">No repositories configured.</p>
  </section>
</template>

<script>
export default {
  name: "RepoSelector",
  props: {
    repos: { type: Array, default: () => [] },
    modelValue: { type: Object, default: null },
  },
  emits: ["update:modelValue"],
  methods: {
    statsLabel(repo) {
      const d = repo.doc_count != null ? repo.doc_count : "…";
      const t = repo.task_count != null ? repo.task_count : "…";
      return `${d} docs · ${t} tasks`;
    },
  },
};
</script>

<style scoped>
.repo-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 1rem;
}
.repo-card {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  text-align: left;
  padding: 1rem;
  border: 1px solid #2a3644;
  background: #1a222c;
  color: inherit;
  border-radius: 8px;
  cursor: pointer;
  min-width: 0;
  overflow: hidden;
  transition: border-color 0.15s ease, background 0.15s ease, transform 0.15s ease;
}
.repo-card > strong,
.repo-card > span {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.repo-card:hover {
  border-color: #7aa2f7;
  background: #223044;
  transform: translateY(-2px);
}
.repo-card-meta {
  font-size: 0.8rem;
  opacity: 0.8;
}
</style>
