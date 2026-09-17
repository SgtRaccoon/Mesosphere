<template>
  <div class="kanban">
    <section v-for="col in columns" :key="col" class="kanban-col">
      <h2>{{ col }}</h2>
      <TaskCard v-for="task in grouped[col]" :key="task.card_id" :task="task" />
    </section>
  </div>
</template>

<script>
import TaskCard from "./TaskCard.vue";
import { tasksByColumn } from "../assets/kanban.js";

export default {
  name: "KanbanBoard",
  components: { TaskCard },
  props: {
    tasks: { type: Array, default: () => [] },
  },
  computed: {
    grouped() {
      return tasksByColumn(this.tasks).map;
    },
    columns() {
      return tasksByColumn(this.tasks).columns;
    },
  },
};
</script>

<style scoped>
.kanban {
  display: flex;
  gap: 1rem;
  padding: 1rem;
  overflow-x: auto;
}
.kanban-col {
  min-width: 220px;
  background: #151b22;
  border-radius: 8px;
  padding: 0.75rem;
}
</style>
