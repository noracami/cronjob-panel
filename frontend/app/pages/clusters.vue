<template>
  <div>
    <div class="flex justify-between items-center mb-4">
      <h2 class="text-xl font-semibold">叢集管理</h2>
      <UButton icon="i-heroicons-plus" label="新增叢集" @click="showForm = true" />
    </div>

    <UTable :columns="columns" :rows="clusters">
      <template #actions-data="{ row }">
        <div class="flex gap-2">
          <UButton
            :to="`/clusters/${row.id}/cronjobs`"
            size="xs"
            variant="soft"
            label="CronJobs"
          />
          <UButton
            size="xs"
            color="red"
            variant="soft"
            icon="i-heroicons-trash"
            @click="remove(row.id)"
          />
        </div>
      </template>
    </UTable>

    <ClusterForm v-model:open="showForm" @created="refresh" />
  </div>
</template>

<script setup lang="ts">
const { apiFetch } = useApi()
const showForm = ref(false)
const clusters = ref<any[]>([])

const columns = [
  { key: 'name', label: '名稱' },
  { key: 'endpoint', label: 'Endpoint' },
  { key: 'auth_type', label: '認證方式' },
  { key: 'actions', label: '操作' },
]

async function refresh() {
  clusters.value = await apiFetch('/clusters')
}

async function remove(id: string) {
  await apiFetch(`/clusters/${id}`, { method: 'DELETE' })
  refresh()
}

onMounted(refresh)
</script>
