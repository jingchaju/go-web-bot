<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Bot, LockKeyhole, ShieldCheck } from 'lucide-vue-next'
import { useAuthStore } from '../stores/auth'
const account = ref(''); const password = ref(''); const error = ref(''); const loading = ref(false); const router = useRouter(); const auth = useAuthStore()
async function submit() { loading.value = true; error.value = ''; try { await auth.login(account.value, password.value); router.push('/') } catch (e) { error.value = (e as Error).message } finally { loading.value = false } }
</script>
<template><div class="grid min-h-screen place-items-center bg-[radial-gradient(circle_at_top,#38bdf8,transparent_25%),linear-gradient(135deg,#020617,#0f172a_45%,#172554)] p-6"><form class="card w-full max-w-md p-8" @submit.prevent="submit"><div class="mb-8 flex items-center gap-3"><div class="grid h-12 w-12 place-items-center rounded-2xl bg-gradient-to-br from-blue-600 to-cyan-400 text-white shadow-lg"><Bot /></div><div><h1 class="text-2xl font-black">TelePilot Console</h1><p class="text-slate-500">企业级 Telegram Bot 管理系统</p></div></div><div class="mb-5 rounded-2xl bg-blue-50 p-4 text-sm text-blue-700"><ShieldCheck class="mr-2 inline h-4 w-4" />登录密码会使用一次性 RSA-OAEP 挑战加密后提交。</div><label class="label">管理员账号</label><input v-model="account" class="input mb-4" placeholder="管理员账号"><label class="label">管理员密码</label><input v-model="password" class="input mb-4" type="password" placeholder="安全密码"><p v-if="error" class="alert-error mb-4">{{ error }}</p><button class="btn-primary w-full" :disabled="loading"><LockKeyhole class="h-4 w-4" />{{ loading ? '校验中...' : '安全登录' }}</button></form></div></template>
