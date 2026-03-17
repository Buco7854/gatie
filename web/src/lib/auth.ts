interface AuthState {
  accessToken: string | null
  memberId: string | null
  username: string | null
  role: string | null
}

let state: AuthState = {
  accessToken: null,
  memberId: null,
  username: null,
  role: null,
}

const listeners = new Set<() => void>()

function emitChange() {
  // Create a new object so useSyncExternalStore detects the change
  state = { ...state }
  for (const listener of listeners) {
    listener()
  }
}

export function subscribe(listener: () => void): () => void {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

export function getSnapshot(): AuthState {
  return state
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
  emitChange()
}

export function clearAuth() {
  state.accessToken = null
  state.memberId = null
  state.username = null
  state.role = null
  emitChange()
}

export function getAccessToken(): string | null {
  return state.accessToken
}

export function isAuthenticated(): boolean {
  return state.accessToken !== null
}
