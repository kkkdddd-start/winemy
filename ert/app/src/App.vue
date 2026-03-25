<template>
  <div id="app" class="ert-app">
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useShortcuts } from './shortcuts/useShortcuts'

onMounted(() => {
  useShortcuts()
})
</script>

<style>
.ert-app {
  width: 100%;
  height: 100vh;
  overflow: hidden;
}
</style>
