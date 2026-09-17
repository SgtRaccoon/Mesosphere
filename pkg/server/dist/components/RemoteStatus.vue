<template>
  <button
    v-if="behind > 0"
    type="button"
    class="refresh-icon"
    :class="{ spinning: syncing }"
    :disabled="syncing"
    title="Remote changes available"
    @click="$emit('refresh')"
  >
    ↻
  </button>
</template>

<script>
export default {
  name: "RemoteStatus",
  props: {
    behind: { type: Number, default: 0 },
    syncing: { type: Boolean, default: false },
  },
  emits: ["refresh"],
};
</script>

<style scoped>
.refresh-icon {
  position: relative;
  margin-right: auto;
}
.refresh-icon.spinning {
  animation: spin 0.8s linear infinite;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
