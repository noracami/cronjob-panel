<!-- frontend/app/components/PodLogViewer.vue -->
<template>
  <UModal v-model:open="open">
    <template #header>
      <h3 class="text-lg font-semibold">Pod Log</h3>
    </template>
    <div class="p-4">
      <pre class="bg-gray-900 text-green-400 p-4 rounded text-sm overflow-auto max-h-96 font-mono">{{ log || '載入中...' }}</pre>
    </div>
  </UModal>
</template>

<script setup lang="ts">
const open = defineModel<boolean>('open', { default: false })
defineProps<{ log: string }>()
</script>
