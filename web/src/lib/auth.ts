interface AuthState {
  accessToken: string | null
  memberId: string | null
  username: string | null
  role: string | null
}

const state: AuthState = {
  accessToken: null,
  memberId: null,
  username: null,
  role: null,
}

export function setAuth(data: {
  access_token: string
  member_id: string
  username: string
  role: string
}) {
  state.accessToken = data.access_token
  state.memberId = data.member_id
  state.username = data.username
  state.role = data.role
}

export function clearAuth() {
  state.accessToken = null
  state.memberId = null
  state.username = null
  state.role = null
}

export function getAccessToken(): string | null {
  return state.accessToken
}

export function getAuthState(): Readonly<AuthState> {
  return state
}

export function isAuthenticated(): boolean {
  return state.accessToken !== null
}
