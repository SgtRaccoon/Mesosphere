<template>
  <div id="app-root">
    <TopBar>
      <button type="button" class="split-toggle" @click="split = !split">
        <Icon name="columns-two" /> Split View
      </button>
    </TopBar>
    <SplitView v-if="split" :panes="panes">
      <template #default="{ pane, index }">
        <PaneWorkspace :pane="pane" :repos="repos" @update="updatePane(index, $event)" />
      </template>
    </SplitView>
    <PaneWorkspace v-else :pane="panes[0]" :repos="repos" @update="updatePane(0, $event)" />
  </div>
</template>

<script>
import TopBar from "./components/TopBar.vue";
import SplitView from "./components/SplitView.vue";
import PaneWorkspace from "./components/PaneWorkspace.vue";
import Icon from "./components/Icon.vue";

function emptyPane() {
  return { repo: null, tab: "docs", picking: true };
}

export default {
  name: "App",
  components: { TopBar, SplitView, PaneWorkspace, Icon },
  data() {
    return {
      repos: [],
      split: false,
      panes: [emptyPane(), emptyPane()],
    };
  },
  methods: {
    updatePane(index, patch) {
      this.panes[index] = { ...this.panes[index], ...patch };
    },
  },
};
</script>
