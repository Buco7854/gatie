import { apiFetch } from './client'
import type { Member, MembersPage } from '@/lib/types'

export function listMembers(page: number, perPage: number): Promise<MembersPage> {
  return apiFetch<MembersPage>(`/members?page=${page}&per_page=${perPage}`)
}

export interface CreateMemberData {
  username: string
  display_name?: string
  password: string
  role: 'ADMIN' | 'MEMBER'
}

export function createMember(data: CreateMemberData): Promise<Member> {
  return apiFetch<Member>('/members', { method: 'POST', body: JSON.stringify(data) })
}

export interface UpdateMemberData {
  username: string
  display_name?: string | null
  role: 'ADMIN' | 'MEMBER'
}

export function updateMember(id: string, data: UpdateMemberData): Promise<Member> {
  return apiFetch<Member>(`/members/${id}`, { method: 'PATCH', body: JSON.stringify(data) })
}

export function deleteMember(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/members/${id}`, { method: 'DELETE' })
}
