<!-- frontend/app/pages/clusters/[id]/cronjobs.vue -->
<template>
  <div>
    <div class="flex items-center gap-2 mb-4">
      <UButton to="/clusters" variant="ghost" icon="i-heroicons-arrow-left" />
      <h2 class="text-xl font-semibold">CronJobs</h2>
    </div>

    <CronJobTable
      :cronjobs="cronjobs"
      :cluster-id="clusterId"
      @toggle-suspend="toggleSuspend"
      @trigger="trigger"
    />
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const { apiFetch } = useApi()

const clusterId = route.params.id as string
const cronjobs = ref<any[]>([])

async function refresh() {
  cronjobs.value = await apiFetch(`/clusters/${clusterId}/cronjobs`)
}

async function toggleSuspend(row: any) {
  const ns = row.metadata.namespace
  const name = row.metadata.name
  await apiFetch(`/clusters/${clusterId}/namespaces/${ns}/cronjobs/${name}/suspend`, {
    method: 'PATCH',
    body: JSON.stringify({ suspend: !row.spec?.suspend }),
  })
  refresh()
}

async function trigger(row: any) {
  const ns = row.metadata.namespace
  const name = row.metadata.name
  await apiFetch(`/clusters/${clusterId}/namespaces/${ns}/cronjobs/${name}/trigger`, {
    method: 'POST',
  })
  refresh()
}

onMounted(refresh)
</script>
