// frontend/composables/useAuth.ts
export function useAuth() {
  const { apiFetch } = useApi()

  function login() {
    window.location.href = '/api/auth/discord'
  }

  async function logout() {
    await apiFetch('/auth/logout', { method: 'POST' })
    navigateTo('/login')
  }

  return { login, logout }
}
