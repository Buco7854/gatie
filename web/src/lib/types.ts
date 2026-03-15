export interface Member {
  id: string
  username: string
  display_name?: string
  role: 'ADMIN' | 'MEMBER'
  created_at: string
  updated_at: string
}

export interface MembersPage {
  items: Member[]
  total: number
  page: number
  per_page: number
}
