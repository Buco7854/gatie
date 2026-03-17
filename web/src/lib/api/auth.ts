import { apiFetch } from './client'
import type { AuthTokens } from '@/lib/types'

export function login(username: string, password: string): Promise<AuthTokens> {
  return apiFetch<AuthTokens>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
}

export function logout(): Promise<undefined> {
  return apiFetch<undefined>('/auth/logout', { method: 'POST' })
}

export function getSetupStatus(): Promise<{ needs_setup: boolean }> {
  return apiFetch<{ needs_setup: boolean }>('/setup/status')
}

export function setup(username: string, password: string): Promise<AuthTokens> {
  return apiFetch<AuthTokens>('/setup', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
}
