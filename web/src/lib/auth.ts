import { api, type AuthResponse } from "./api"

type AuthState = {
  isAuthenticated: boolean
  memberId: string | null
  username: string | null
  role: string | null
  refreshToken: string | null
}

const REFRESH_TOKEN_KEY = "gatie_refresh_token"

let state: AuthState = {
  isAuthenticated: false,
  memberId: null,
  username: null,
  role: null,
  refreshToken: localStorage.getItem(REFRESH_TOKEN_KEY),
}

let listeners: Array<() => void> = []

function notify() {
  listeners.forEach((fn) => fn())
}

export const auth = {
  getState(): AuthState {
    return state
  },

  subscribe(fn: () => void) {
    listeners.push(fn)
    return () => {
      listeners = listeners.filter((l) => l !== fn)
    }
  },

  handleAuthResponse(res: AuthResponse) {
    api.setAccessToken(res.access_token)
    localStorage.setItem(REFRESH_TOKEN_KEY, res.refresh_token)
    state = {
      isAuthenticated: true,
      memberId: res.member_id,
      username: res.username,
      role: res.role,
      refreshToken: res.refresh_token,
    }
    notify()
  },

  async tryRefresh(): Promise<boolean> {
    const refreshToken = state.refreshToken || localStorage.getItem(REFRESH_TOKEN_KEY)
    if (!refreshToken) return false

    try {
      const res = await api.refresh(refreshToken)
      this.handleAuthResponse(res)
      return true
    } catch {
      this.clear()
      return false
    }
  },

  clear() {
    api.setAccessToken(null)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    state = {
      isAuthenticated: false,
      memberId: null,
      username: null,
      role: null,
      refreshToken: null,
    }
    notify()
  },

  async logout() {
    const refreshToken = state.refreshToken
    if (refreshToken) {
      try {
        await api.logout(refreshToken)
      } catch {
        // best effort
      }
    }
    this.clear()
  },
}
