import { fingerprint, request } from './http'
export interface BotConfig { bot_token: string; webhook_port: string; webhook_path: string; webhook_secret: string; commands_json: string; keyboard_json: string; welcome_html: string; bot_running: boolean }
export async function login(account: string, password: string) { return request<{token:string; account:string; expires_in:number}>('/login', { method: 'POST', body: JSON.stringify({ account, password, fingerprint: await fingerprint() }) }) }
export const me = () => request<{account:string}>('/me')
export const dashboard = () => request<{users:number; bot:{running:boolean; processed:number; started_at:string}; server_time:string}>('/dashboard')
export const getBotConfig = () => request<BotConfig>('/bot/config')
export const saveBotConfig = (payload: BotConfig) => request<BotConfig>('/bot/config', { method: 'POST', body: JSON.stringify(payload) })
export const startBot = () => request('/bot/start', { method: 'POST', body: JSON.stringify({}) })
export const stopBot = () => request('/bot/stop', { method: 'POST', body: JSON.stringify({}) })
