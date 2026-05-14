const API_PREFIX = '/api/admin'
export type ApiResult<T> = T & { message?: string }
async function fingerprint() { return btoa([navigator.userAgent, screen.width, screen.height, Intl.DateTimeFormat().resolvedOptions().timeZone].join('|')) }
async function sign(payload: string) {
  const enc = new TextEncoder()
  const secret = import.meta.env.VITE_API_HMAC_SECRET || 'change-me-jwt-secret'
  const key = await crypto.subtle.importKey('raw', enc.encode(secret), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign'])
  const digest = await crypto.subtle.sign('HMAC', key, enc.encode(payload))
  return btoa(String.fromCharCode(...new Uint8Array(digest)))
}
export async function request<T>(path: string, options: RequestInit = {}): Promise<ApiResult<T>> {
  const token = localStorage.getItem('token') || ''
  const body = typeof options.body === 'string' ? options.body : options.body ? JSON.stringify(options.body) : ''
  const headers: Record<string, string> = { 'Content-Type': 'application/json', 'X-Device-Fingerprint': await fingerprint(), ...(options.headers as Record<string,string> || {}) }
  if (token) headers.Authorization = `Bearer ${token}`
  if (body) headers['X-Payload-Signature'] = await sign(body)
  const res = await fetch(`${API_PREFIX}${path}`, { ...options, body: body || undefined, headers })
  const json = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(json.message || `HTTP ${res.status}`)
  return json
}
export { fingerprint }
