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
        <span class="repo-card-path">{{ repo.path }}</span>
        <span class="repo-card-meta">
          <span class="repo-stat"><Icon name="file-text" /> {{ count(repo, "doc") }}</span>
          <span class="repo-stat"><Icon name="list-todo" /> {{ count(repo, "task") }}</span>
        </span>
      </button>
    </div>
    <p v-if="!repos.length" class="empty">No repositories configured.</p>
  </section>
</template>

<script>
import Icon from "./Icon.vue";

export default {
  name: "RepoSelector",
  components: { Icon },
  props: {
    repos: { type: Array, default: () => [] },
    modelValue: { type: Object, default: null },
  },
  emits: ["update:modelValue"],
  methods: {
    count(repo, kind) {
      const key = kind === "doc" ? "doc_count" : "task_count";
      return repo[key] != null ? repo[key] : "…";
    },
  },
};
</script>

<style scoped>
.repo-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1.25rem;
}
.repo-card {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  text-align: left;
  padding: 1.5rem 1.35rem;
  min-height: 9rem;
  border: 1px solid #2a3644;
  background: #1a222c;
  color: inherit;
  border-radius: 10px;
  cursor: pointer;
  min-width: 0;
  overflow: hidden;
  transition: border-color 0.15s ease, background 0.15s ease, transform 0.15s ease;
}
.repo-card > strong,
.repo-card-path {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.repo-card > strong {
  font-size: 1.1rem;
}
.repo-card:hover {
  border-color: #7aa2f7;
  background: #223044;
  transform: translateY(-2px);
}
.repo-card-meta {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  margin-top: auto;
  font-size: 0.85rem;
  opacity: 0.85;
}
.repo-stat {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
}
</style>
