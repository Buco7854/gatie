const BASE_URL = ""

type ApiError = {
  status: number
  title: string
  detail: string
}

class ApiClient {
  private accessToken: string | null = null

  setAccessToken(token: string | null) {
    this.accessToken = token
  }

  getAccessToken() {
    return this.accessToken
  }

  async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string>),
    }

    if (this.accessToken) {
      headers["Authorization"] = `Bearer ${this.accessToken}`
    }

    const response = await fetch(`${BASE_URL}${path}`, {
      ...options,
      headers,
    })

    if (!response.ok) {
      const error: ApiError = await response.json().catch(() => ({
        status: response.status,
        title: response.statusText,
        detail: "An error occurred",
      }))
      throw error
    }

    if (response.status === 204) {
      return undefined as T
    }

    return response.json()
  }

  setup(username: string, password: string) {
    return this.request<AuthResponse>("/setup", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    })
  }

  login(username: string, password: string) {
    return this.request<AuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    })
  }

  refresh(refreshToken: string) {
    return this.request<AuthResponse>("/auth/refresh", {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
  }

  logout(refreshToken: string) {
    return this.request<void>("/auth/logout", {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
  }

  healthCheck() {
    return this.request<HealthResponse>("/health")
  }
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  member_id: string
  role: string
  username: string
}

export interface HealthResponse {
  status: string
  services: Record<string, string>
}

export const api = new ApiClient()
