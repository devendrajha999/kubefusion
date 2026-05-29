const TOKEN_KEY = 'kubefusion_token'

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

export async function api<T>(url: string, init?: RequestInit): Promise<T> {
  const token = getToken()
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(init?.headers || {}),
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(url, { ...init, headers })
  if (res.status === 401) {
    clearToken()
    throw new Error('Unauthorized')
  }
  if (!res.ok) {
    const txt = await res.text()
    throw new Error(txt || `Request failed: ${res.status}`)
  }
  return res.json() as Promise<T>
}
