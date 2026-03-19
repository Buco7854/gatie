import { getAccessToken, setAuth, clearAuth } from '@/lib/auth'
import type { AuthTokens } from '@/lib/types'

export interface ApiErrorDetail {
  location?: string
  message?: string
  value?: unknown
}

interface ApiErrorBody {
  status?: number
  detail?: string
  errors?: ApiErrorDetail[]
}

export class ApiError extends Error {
  status: number
  detail: string
  errors: ApiErrorDetail[]

  constructor(status: number, body: ApiErrorBody | null) {
    const detail = body?.detail ?? `HTTP ${status}`
    super(detail)
    this.name = 'ApiError'
    this.status = status
    this.detail = detail
    this.errors = body?.errors ?? []
  }

  hasFieldError(location: string): boolean {
    return this.errors.some((e) => e.location === location)
  }
}

let refreshPromise: Promise<boolean> | null = null

async function refreshTokens(): Promise<boolean> {
  if (refreshPromise) return refreshPromise

  refreshPromise = (async () => {
    try {
      const res = await fetch('/api/auth/refresh', {
        method: 'POST',
        credentials: 'include',
      })
      if (!res.ok) return false
      const data = (await res.json()) as AuthTokens
      setAuth(data)
      return true
    } catch {
      return false
    } finally {
      refreshPromise = null
    }
  })()

  return refreshPromise
}

async function throwApiError(res: Response): never {
  const body = (await res.json().catch(() => null)) as ApiErrorBody | null
  throw new ApiError(res.status, body)
}

export async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getAccessToken()
  const headers = new Headers(options.headers as HeadersInit | undefined)
  if (options.body) headers.set('Content-Type', 'application/json')
  if (token) headers.set('Authorization', `Bearer ${token}`)

  const res = await fetch(`/api${path}`, { ...options, credentials: 'include', headers })

  if (res.status === 401) {
    const refreshed = await refreshTokens()
    if (refreshed) {
      const newToken = getAccessToken()
      if (newToken) headers.set('Authorization', `Bearer ${newToken}`)
      const retry = await fetch(`/api${path}`, { ...options, credentials: 'include', headers })
      if (!retry.ok) await throwApiError(retry)
      if (retry.status === 204) return undefined as T
      return retry.json() as Promise<T>
    }
    clearAuth()
    window.location.href = '/login'
    throw new ApiError(401, null)
  }

  if (!res.ok) await throwApiError(res)

  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

export async function silentRefresh(): Promise<boolean> {
  return refreshTokens()
}
