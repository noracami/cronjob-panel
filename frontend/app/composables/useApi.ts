// frontend/composables/useApi.ts
export function useApi() {
  async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
    const res = await fetch(`/api${path}`, {
      credentials: 'include',
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    })

    if (res.status === 401) {
      navigateTo('/login')
      throw new Error('Unauthorized')
    }

    if (!res.ok) {
      const body = await res.json().catch(() => ({}))
      throw new Error(body.error || `API error: ${res.status}`)
    }

    return res.json()
  }

  return { apiFetch }
}
