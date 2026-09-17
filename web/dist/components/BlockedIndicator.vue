<template>
  <button
    v-if="task && task.is_blocked"
    type="button"
    class="blocked-indicator"
    :title="tooltip"
    @click.stop="$emit('inspect', task)"
  >
    ⛔
  </button>
</template>

<script>
export default {
  name: "BlockedIndicator",
  props: {
    task: { type: Object, required: true },
    tasks: { type: Array, default: () => [] },
  },
  emits: ["inspect"],
  computed: {
    tooltip() {
      const ids = this.task.blocker_ids || [];
      const titles = ids.map((id) => {
        const b = this.tasks.find((t) => t.card_id === id);
        return b ? b.title || id : id;
      });
      return titles.join(", ") || "Blocked";
    },
  },
};
</script>

<style scoped>
.blocked-indicator {
  background: #b42318;
  color: #fff;
  border: 0;
  border-radius: 999px;
  cursor: pointer;
}
</style>
