<template>
  <div id="app-root">
    <TopBar v-if="selected" :repo="selected" :tab="tab" @update:tab="tab = $event">
      <button type="button" class="split-toggle" @click="split = !split">Split View</button>
    </TopBar>
    <SplitView v-if="split && selected" :panes="panes">
      <template #default="{ pane }">
        <DocsView v-if="pane.tab === 'docs'" :repo-id="pane.repoId" :docs="docs" />
        <KanbanBoard v-else :tasks="tasks" />
      </template>
    </SplitView>
    <DocsView v-else-if="selected && tab === 'docs'" :repo-id="selected.id" :docs="docs" />
    <KanbanBoard v-else-if="selected" :tasks="tasks" />
    <RepoSelector v-else :repos="repos" v-model="selected" />
  </div>
</template>

<script>
import TopBar from "./components/TopBar.vue";
import RepoSelector from "./components/RepoSelector.vue";
import SplitView from "./components/SplitView.vue";
import DocsView from "./components/DocsView.vue";
import KanbanBoard from "./components/KanbanBoard.vue";

export default {
  name: "App",
  components: { TopBar, RepoSelector, SplitView, DocsView, KanbanBoard },
  data() {
    return {
      repos: [],
      selected: null,
      tab: "docs",
      split: false,
      docs: [],
      tasks: [],
    };
  },
  computed: {
    panes() {
      const id = this.selected && this.selected.id;
      return [
        { repoId: id, tab: this.tab },
        { repoId: id, tab: this.tab === "docs" ? "tasks" : "docs" },
      ];
    },
  },
};
</script>
