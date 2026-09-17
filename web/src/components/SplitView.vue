<template>
  <div class="split-view" :class="orientation">
    <section class="pane" v-for="(pane, i) in panes" :key="i">
      <slot :pane="pane" :index="i" />
    </section>
  </div>
</template>

<script>
export function splitOrientation(width, height) {
  return width > height ? "row" : "col";
}

export default {
  name: "SplitView",
  props: {
    panes: { type: Array, default: () => [{}, {}] },
  },
  data() {
    return { orientation: "row" };
  },
  mounted() {
    this.updateOrientation();
    window.addEventListener("resize", this.updateOrientation);
  },
  beforeUnmount() {
    window.removeEventListener("resize", this.updateOrientation);
  },
  methods: {
    updateOrientation() {
      this.orientation = splitOrientation(window.innerWidth, window.innerHeight);
    },
  },
};
</script>

<style scoped>
.split-view {
  display: flex;
  min-height: 50vh;
}
.split-view.row {
  flex-direction: row;
}
.split-view.col {
  flex-direction: column;
}
.pane {
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow: auto;
}
</style>
