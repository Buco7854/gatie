import { apiFetch } from './client'
import type { Role, Permission } from '@/lib/types'

// --- Roles ---

export function listRoles(): Promise<Role[]> {
  return apiFetch<Role[]>('/roles')
}

export function createRole(data: { id: string; description: string }): Promise<Role> {
  return apiFetch<Role>('/roles', { method: 'POST', body: JSON.stringify(data) })
}

export function updateRole(id: string, data: { description: string }): Promise<Role> {
  return apiFetch<Role>(`/roles/${id}`, { method: 'PATCH', body: JSON.stringify(data) })
}

export function deleteRole(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/roles/${id}`, { method: 'DELETE' })
}

export function setRolePermissions(roleId: string, permissions: string[]): Promise<Role> {
  return apiFetch<Role>(`/roles/${roleId}/permissions`, {
    method: 'PUT',
    body: JSON.stringify({ permissions }),
  })
}

// --- Permissions ---

export function listPermissions(): Promise<Permission[]> {
  return apiFetch<Permission[]>('/permissions')
}

export function createPermission(data: { id: string; description: string }): Promise<Permission> {
  return apiFetch<Permission>('/permissions', { method: 'POST', body: JSON.stringify(data) })
}

export function updatePermission(id: string, data: { description: string }): Promise<Permission> {
  return apiFetch<Permission>(`/permissions/${id}`, { method: 'PATCH', body: JSON.stringify(data) })
}

export function deletePermission(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/permissions/${id}`, { method: 'DELETE' })
}
