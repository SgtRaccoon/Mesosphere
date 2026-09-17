<template>
  <div class="publish-wrap">
    <button
      v-if="visible"
      type="button"
      class="publish-btn"
      :disabled="publishing"
      @click="$emit('publish')"
    >
      <span v-if="publishing" class="spinner" aria-hidden="true"></span>
      <Icon name="upload" /> {{ publishing ? "Publishing…" : "Publish" }}
    </button>
    <div v-if="error" class="publish-modal" role="alertdialog">
      <p>{{ error }}</p>
      <button type="button" @click="$emit('dismiss')">Close</button>
    </div>
  </div>
</template>

<script>
import Icon from "./Icon.vue";

export default {
  name: "PublishButton",
  components: { Icon },
  props: {
    visible: { type: Boolean, default: false },
    publishing: { type: Boolean, default: false },
    error: { type: String, default: "" },
  },
  emits: ["publish", "dismiss"],
};
</script>
