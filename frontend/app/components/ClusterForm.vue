<template>
  <UModal v-model:open="open">
    <template #header>
      <h3 class="text-lg font-semibold">新增叢集</h3>
    </template>
    <UForm :state="form" @submit="submit" class="p-4 space-y-4">
      <UFormField label="名稱" required>
        <UInput v-model="form.name" placeholder="prod-gke" />
      </UFormField>
      <UFormField label="Endpoint" required>
        <UInput v-model="form.endpoint" placeholder="https://k8s.example.com" />
      </UFormField>
      <UFormField label="認證方式" required>
        <USelect v-model="form.auth_type" :items="['token', 'kubeconfig']" />
      </UFormField>
      <UFormField label="認證資料" required>
        <UTextarea v-model="form.auth_data" rows="4" placeholder="Bearer token 或 kubeconfig 內容" />
      </UFormField>
      <UButton type="submit" label="新增" :loading="loading" />
    </UForm>
  </UModal>
</template>

<script setup lang="ts">
const open = defineModel<boolean>('open', { default: false })
const emit = defineEmits<{ created: [] }>()
const { apiFetch } = useApi()
const loading = ref(false)

const form = reactive({
  name: '',
  endpoint: '',
  auth_type: 'token',
  auth_data: '',
})

async function submit() {
  loading.value = true
  try {
    await apiFetch('/clusters', {
      method: 'POST',
      body: JSON.stringify(form),
    })
    emit('created')
    open.value = false
    Object.assign(form, { name: '', endpoint: '', auth_type: 'token', auth_data: '' })
  } finally {
    loading.value = false
  }
}
</script>
