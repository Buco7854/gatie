export interface AuthTokens {
  access_token: string
  refresh_token: string
  member_id: string
  role: string
  username: string
}

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

export interface Gate {
  id: string
  name: string
  status_ttl_seconds: number
  created_at: string
  updated_at: string
}

export interface GateWithToken extends Gate {
  token: string
}

export interface GatesPage {
  items: Gate[]
  total: number
  page: number
  per_page: number
}
