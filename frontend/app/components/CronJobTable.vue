<!-- frontend/app/components/CronJobTable.vue -->
<template>
  <UTable :columns="columns" :rows="cronjobs">
    <template #status-data="{ row }">
      <UBadge :color="row.spec?.suspend ? 'amber' : 'emerald'">
        {{ row.spec?.suspend ? '已暫停' : '執行中' }}
      </UBadge>
    </template>
    <template #actions-data="{ row }">
      <div class="flex gap-2">
        <UButton
          :to="`/clusters/${clusterId}/${row.metadata.namespace}/${row.metadata.name}`"
          size="xs"
          variant="soft"
          label="詳情"
        />
        <UButton
          size="xs"
          variant="soft"
          :icon="row.spec?.suspend ? 'i-heroicons-play' : 'i-heroicons-pause'"
          @click="$emit('toggleSuspend', row)"
        />
        <UButton
          size="xs"
          variant="soft"
          icon="i-heroicons-play-circle"
          label="觸發"
          @click="$emit('trigger', row)"
        />
      </div>
    </template>
  </UTable>
</template>

<script setup lang="ts">
defineProps<{
  cronjobs: any[]
  clusterId: string
}>()

defineEmits<{
  toggleSuspend: [row: any]
  trigger: [row: any]
}>()

const columns = [
  { key: 'metadata.name', label: '名稱' },
  { key: 'metadata.namespace', label: 'Namespace' },
  { key: 'spec.schedule', label: '排程' },
  { key: 'status.lastScheduleTime', label: '上次執行' },
  { key: 'status', label: '狀態' },
  { key: 'actions', label: '操作' },
]
</script>
