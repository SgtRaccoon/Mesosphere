<template>
  <div class="blocker-modal" role="dialog" v-if="open">
    <h2>Blocked by</h2>
    <ul>
      <li v-for="b in blockers" :key="b.card_id">
        <strong>{{ b.card_id }}</strong> — {{ b.title }}
        <p>{{ b.scope }}</p>
      </li>
    </ul>
    <button type="button" @click="$emit('close')">Close</button>
  </div>
</template>

<script>
export default {
  name: "BlockerModal",
  props: {
    open: { type: Boolean, default: false },
    blockers: { type: Array, default: () => [] },
  },
  emits: ["close"],
};
</script>
