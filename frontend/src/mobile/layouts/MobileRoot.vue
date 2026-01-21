<template>
  <div class="mobile-root">
    <router-view />
  </div>
</template>

<script setup lang="ts"></script>

<style scoped>
.mobile-root {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
</style>
