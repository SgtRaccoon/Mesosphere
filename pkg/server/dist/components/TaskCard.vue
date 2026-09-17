<template>
  <article class="task-card" :data-id="task.card_id">
    <header>
      <span class="task-id">{{ task.card_id }}</span>
      <span class="task-status">{{ task.status }}</span>
    </header>
    <p class="task-title">{{ task.title }}</p>
    <p class="task-scope">{{ truncated }}</p>
    <BlockedIndicator :task="task" :tasks="tasks" @inspect="$emit('inspect', $event)" />
  </article>
</template>

<script>
import { truncateScope } from "../assets/kanban.js";
import BlockedIndicator from "./BlockedIndicator.vue";

export default {
  name: "TaskCard",
  components: { BlockedIndicator },
  props: {
    task: { type: Object, required: true },
    tasks: { type: Array, default: () => [] },
  },
  emits: ["inspect"],
  computed: {
    truncated() {
      return truncateScope(this.task.scope);
    },
  },
};
</script>
