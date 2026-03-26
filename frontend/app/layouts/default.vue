<template>
  <div class="min-h-screen flex">
    <aside class="w-64 border-r p-4 flex flex-col">
      <h1 class="text-lg font-bold mb-4">CronJob Panel</h1>
      <nav class="flex flex-col gap-1">
        <UButton
          to="/clusters"
          variant="ghost"
          icon="i-heroicons-server-stack"
          label="叢集管理"
        />
      </nav>
      <div class="mt-auto">
        <UButton
          variant="ghost"
          icon="i-heroicons-arrow-right-on-rectangle"
          label="登出"
          @click="logout"
        />
      </div>
    </aside>
    <main class="flex-1 p-6">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
const { logout } = useAuth()
</script>
