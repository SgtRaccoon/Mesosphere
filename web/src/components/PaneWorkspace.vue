<template>
  <div class="pane-workspace">
    <PaneBar v-if="pane.repo && !pane.picking" :repo="pane.repo" :tab="pane.tab" @update:tab="$emit('update', { tab: $event })" @open-repos="$emit('update', { picking: true })" />
    <RepoSelector v-if="!pane.repo || pane.picking" :repos="repos" :model-value="null" @update:modelValue="onSelect" />
    <DocsView v-else-if="pane.tab === 'docs'" :repo-id="pane.repo.id" :docs="docs" />
    <KanbanBoard v-else :tasks="tasks" />
  </div>
</template>

<script>
import PaneBar from "./PaneBar.vue";
import RepoSelector from "./RepoSelector.vue";
import DocsView from "./DocsView.vue";
import KanbanBoard from "./KanbanBoard.vue";

export default {
  name: "PaneWorkspace",
  components: { PaneBar, RepoSelector, DocsView, KanbanBoard },
  props: {
    pane: { type: Object, required: true },
    repos: { type: Array, default: () => [] },
  },
  emits: ["update"],
  data() {
    return { docs: [], tasks: [] };
  },
  methods: {
    onSelect(repo) {
      this.$emit("update", { repo, picking: false, tab: this.pane.tab || "docs" });
    },
  },
};
</script>
