<template>
  <div class="docs-editor">
    <textarea v-if="editing" class="docs-editor-textarea" :value="draft" @input="$emit('update:draft', $event.target.value)"></textarea>
    <button
      type="button"
      class="docs-edit-fab"
      :title="editing ? 'Save' : 'Edit'"
      @click="editing ? $emit('save') : $emit('edit')"
    >
      <Icon :name="editing ? 'save' : 'pencil'" />
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
  position: absolute;
  z-index: 2;
  right: 1.5rem;
  bottom: 1.5rem;
  width: 3rem;
  height: 3rem;
  padding: 0;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #2a3644;
  background: #1a222c;
  color: inherit;
  cursor: pointer;
  transition: border-color 0.15s ease, background 0.15s ease, transform 0.15s ease;
}
.docs-edit-fab:hover {
  border-color: #7aa2f7;
  background: #223044;
  transform: translateY(-2px);
}
</style>
