export const STATUS_TODO = "TO DO";
export const STATUS_DONE = "DONE";

export function deriveColumns(tasks) {
  const statuses = [...new Set((tasks || []).map((t) => t.status || STATUS_TODO))];
  const binary = statuses.every((s) => s === STATUS_TODO || s === STATUS_DONE);
  if (binary) return [STATUS_TODO, STATUS_DONE];
  return statuses.filter((s) => s !== "BLOCKED");
}

export function columnForTask(task) {
  if (task.is_blocked) return STATUS_TODO;
  return task.status || STATUS_TODO;
}

export function truncateScope(scope, max = 80) {
  const s = String(scope || "");
  if (s.length <= max) return s;
  return s.slice(0, max - 1) + "…";
}

export function tasksByColumn(tasks) {
  const cols = deriveColumns(tasks);
  const map = Object.fromEntries(cols.map((c) => [c, []]));
  for (const t of tasks || []) {
    const col = columnForTask(t);
    if (!map[col]) map[col] = [];
    map[col].push(t);
  }
  return { columns: Object.keys(map), map };
}
