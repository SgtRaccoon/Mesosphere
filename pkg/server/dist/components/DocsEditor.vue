<template>
  <div class="docs-editor">
    <textarea v-if="editing" class="docs-editor-textarea" :value="draft" @input="$emit('update:draft', $event.target.value)"></textarea>
    <button
      type="button"
      class="docs-edit-fab"
      :title="editing ? 'Save' : 'Edit'"
      @click="editing ? $emit('save') : $emit('edit')"
    >
      {{ editing ? "Save" : "✎" }}
    </button>
  </div>
</template>

<script>
export default {
  name: "DocsEditor",
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
