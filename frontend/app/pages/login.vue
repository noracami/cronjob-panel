<!-- frontend/pages/login.vue -->
<template>
  <div class="flex items-center justify-center min-h-screen">
    <UCard class="w-80">
      <template #header>
        <h2 class="text-lg font-semibold text-center">CronJob Panel</h2>
      </template>
      <UButton
        icon="i-simple-icons-discord"
        label="使用 Discord 登入"
        color="primary"
        block
        size="lg"
        @click="login"
      />
    </UCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })
const { login } = useAuth()
</script>
