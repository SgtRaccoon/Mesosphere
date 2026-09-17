<template>
  <div class="task-editor-modal" role="dialog" v-if="open">
    <h2>{{ task && task.card_id }}</h2>
    <textarea class="task-json" :value="draft" @input="onInput"></textarea>
    <footer>
      <button type="button" v-if="dirty" @click="$emit('save', draft)">Save</button>
      <button type="button" v-if="unpublished" @click="$emit('publish')">Publish</button>
      <button type="button" @click="$emit('close')">Close</button>
    </footer>
  </div>
</template>

<script>
export default {
  name: "TaskEditorModal",
  props: {
    open: { type: Boolean, default: false },
    task: { type: Object, default: null },
    unpublished: { type: Boolean, default: false },
  },
  data() {
    return { draft: "", dirty: false };
  },
  watch: {
    task: {
      immediate: true,
      handler(t) {
        this.draft = t ? JSON.stringify(JSON.parse(t.raw_content || "{}"), null, 2) : "";
        this.dirty = false;
      },
    },
  },
  methods: {
    onInput(e) {
      this.draft = e.target.value;
      this.dirty = true;
    },
  },
  emits: ["save", "publish", "close"],
};
</script>
