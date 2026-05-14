declare global {
  interface Window { __APP_CONFIG__?: { adminRoutePrefix?: string } }
}

const encoder = new TextEncoder()
const API_PREFIX = window.__APP_CONFIG__?.adminRoutePrefix || import.meta.env.VITE_API_PREFIX || '/api/admin'
export type ApiResult<T> = T & { message?: string }

async function fingerprint() {
  const raw = [navigator.userAgent, screen.width, screen.height, Intl.DateTimeFormat().resolvedOptions().timeZone].join('|')
  const digest = await crypto.subtle.digest('SHA-256', encoder.encode(raw))
  return toBase64(new Uint8Array(digest))
}

function toBase64(bytes: Uint8Array) {
  let binary = ''
  bytes.forEach((b) => { binary += String.fromCharCode(b) })
  return btoa(binary)
}

function pemToArrayBuffer(pem: string) {
  const b64 = pem.replace(/-----BEGIN PUBLIC KEY-----|-----END PUBLIC KEY-----|\s/g, '')
  const binary = atob(b64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
  return bytes.buffer
}

export async function encryptForLogin(publicKeyPem: string, nonce: string, payload: Record<string, string>) {
  const key = await crypto.subtle.importKey('spki', pemToArrayBuffer(publicKeyPem), { name: 'RSA-OAEP', hash: 'SHA-256' }, false, ['encrypt'])
  const ciphertext = await crypto.subtle.encrypt({ name: 'RSA-OAEP', label: encoder.encode(nonce) }, key, encoder.encode(JSON.stringify({ ...payload, nonce })))
  return toBase64(new Uint8Array(ciphertext))
}

async function signRequest(token: string, method: string, path: string, timestamp: string, fp: string, body: string) {
  const key = await crypto.subtle.importKey('raw', encoder.encode(token), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign'])
  const data = [method, `${API_PREFIX}${path}`, timestamp, fp, body].join('\n')
  const digest = await crypto.subtle.sign('HMAC', key, encoder.encode(data))
  return toBase64(new Uint8Array(digest))
}

export async function request<T>(path: string, options: RequestInit = {}): Promise<ApiResult<T>> {
  const token = localStorage.getItem('token') || ''
  const method = (options.method || 'GET').toUpperCase()
  const body = typeof options.body === 'string' ? options.body : options.body ? JSON.stringify(options.body) : ''
  const fp = await fingerprint()
  const headers: Record<string, string> = { 'Content-Type': 'application/json', 'X-Device-Fingerprint': fp, ...(options.headers as Record<string,string> || {}) }
  if (token) headers.Authorization = `Bearer ${token}`
  if (token && body && method !== 'GET') {
    const timestamp = Math.floor(Date.now() / 1000).toString()
    headers['X-Request-Timestamp'] = timestamp
    headers['X-Request-Signature'] = await signRequest(token, method, path, timestamp, fp, body)
  }
  const res = await fetch(`${API_PREFIX}${path}`, { ...options, method, body: body || undefined, headers, credentials: 'same-origin' })
  const json = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(json.message || `HTTP ${res.status}`)
  return json
}
export { API_PREFIX, fingerprint }
