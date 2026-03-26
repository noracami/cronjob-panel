<!-- frontend/app/pages/clusters/[id]/[ns]/[name].vue -->
<template>
  <div>
    <div class="flex items-center gap-2 mb-4">
      <UButton
        :to="`/clusters/${clusterId}/cronjobs`"
        variant="ghost"
        icon="i-heroicons-arrow-left"
      />
      <h2 class="text-xl font-semibold">{{ ns }} / {{ name }}</h2>
    </div>

    <UCard class="mb-4" v-if="cronjob">
      <div class="grid grid-cols-2 gap-2 text-sm">
        <div><strong>排程：</strong>{{ cronjob.spec?.schedule }}</div>
        <div>
          <strong>狀態：</strong>
          <UBadge :color="cronjob.spec?.suspend ? 'amber' : 'emerald'">
            {{ cronjob.spec?.suspend ? '已暫停' : '執行中' }}
          </UBadge>
        </div>
        <div><strong>上次執行：</strong>{{ cronjob.status?.lastScheduleTime || '無' }}</div>
      </div>
    </UCard>

    <h3 class="text-lg font-semibold mb-2">執行歷史</h3>
    <JobHistoryTable :jobs="jobs" @view-log="viewLog" />

    <PodLogViewer v-model:open="showLog" :log="podLog" />
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const { apiFetch } = useApi()

const clusterId = route.params.id as string
const ns = route.params.ns as string
const name = route.params.name as string

const cronjob = ref<any>(null)
const jobs = ref<any[]>([])
const showLog = ref(false)
const podLog = ref('')

async function refresh() {
  const [cj, jobList] = await Promise.all([
    apiFetch(`/clusters/${clusterId}/namespaces/${ns}/cronjobs/${name}`),
    apiFetch(`/clusters/${clusterId}/namespaces/${ns}/cronjobs/${name}/jobs`),
  ])
  cronjob.value = cj
  jobs.value = jobList || []
}

async function viewLog(job: any) {
  showLog.value = true
  podLog.value = ''
  const podName = job.metadata.name
  try {
    const result = await apiFetch<{ log: string }>(
      `/clusters/${clusterId}/namespaces/${ns}/pods/${podName}/log`
    )
    podLog.value = result.log
  } catch {
    podLog.value = '無法取得 log'
  }
}

onMounted(refresh)
</script>
