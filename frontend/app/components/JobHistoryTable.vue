<!-- frontend/app/components/JobHistoryTable.vue -->
<template>
  <UTable :columns="columns" :rows="jobs">
    <template #status-data="{ row }">
      <UBadge :color="jobStatusColor(row)">
        {{ jobStatusText(row) }}
      </UBadge>
    </template>
    <template #actions-data="{ row }">
      <UButton
        size="xs"
        variant="soft"
        label="查看 Log"
        @click="$emit('viewLog', row)"
      />
    </template>
  </UTable>
</template>

<script setup lang="ts">
defineProps<{ jobs: any[] }>()
defineEmits<{ viewLog: [row: any] }>()

const columns = [
  { key: 'metadata.name', label: 'Job 名稱' },
  { key: 'status.startTime', label: '開始時間' },
  { key: 'status.completionTime', label: '完成時間' },
  { key: 'status', label: '狀態' },
  { key: 'actions', label: '操作' },
]

function jobStatusText(job: any): string {
  if (job.status?.succeeded) return '成功'
  if (job.status?.failed) return '失敗'
  if (job.status?.active) return '執行中'
  return '未知'
}

function jobStatusColor(job: any): string {
  if (job.status?.succeeded) return 'emerald'
  if (job.status?.failed) return 'red'
  if (job.status?.active) return 'sky'
  return 'gray'
}
</script>
