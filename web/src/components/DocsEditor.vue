<template>
  <div class="docs-editor">
    <textarea v-if="editing" class="docs-editor-textarea" :value="draft" @input="$emit('update:draft', $event.target.value)"></textarea>
    <button
      type="button"
      class="docs-edit-fab"
      :title="editing ? 'Save' : 'Edit'"
      @click="editing ? $emit('save') : $emit('edit')"
    >
      <Icon :name="editing ? 'save' : 'file-text'" /> {{ editing ? "Save" : "Edit" }}
    </button>
  </div>
</template>

<script>
import Icon from "./Icon.vue";

export default {
  name: "DocsEditor",
  components: { Icon },
  props: {
    editing: { type: Boolean, default: false },
    draft: { type: String, default: "" },
  },
  emits: ["edit", "save", "update:draft"],
};
</script>

<style scoped>
.docs-edit-fab {
  position: fixed;
  right: 1.5rem;
  bottom: 1.5rem;
}
</style>
