import { apiFetch } from './client'
import type { Gate, GateWithToken, GatesPage } from '@/lib/types'

export function listGates(page: number, perPage: number): Promise<GatesPage> {
  return apiFetch<GatesPage>(`/gates?page=${page}&per_page=${perPage}`)
}

export function createGate(data: { name: string; status_ttl_seconds: number }): Promise<GateWithToken> {
  return apiFetch<GateWithToken>('/gates', { method: 'POST', body: JSON.stringify(data) })
}

export function updateGate(id: string, data: { name: string; status_ttl_seconds: number }): Promise<Gate> {
  return apiFetch<Gate>(`/gates/${id}`, { method: 'PATCH', body: JSON.stringify(data) })
}

export function deleteGate(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/gates/${id}`, { method: 'DELETE' })
}

export function regenerateGateToken(id: string): Promise<GateWithToken> {
  return apiFetch<GateWithToken>(`/gates/${id}/token`, { method: 'POST' })
}
