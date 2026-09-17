<template>
  <header class="top-bar">
    <div class="top-bar-left">Mesosphere</div>
    <nav class="top-bar-right">
      <slot />
    </nav>
  </header>
</template>

<script>
export default {
  name: "TopBar",
};
</script>

<style scoped>
.top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #2a3644;
}
.top-bar-left {
  font-weight: 600;
}
.top-bar-right {
  display: flex;
  gap: 0.5rem;
}
</style>
