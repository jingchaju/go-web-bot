<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Bot } from 'lucide-vue-next'
import { useAuthStore } from '../stores/auth'
const account = ref(''); const password = ref(''); const error = ref(''); const loading = ref(false); const router = useRouter(); const auth = useAuthStore()
async function submit() { loading.value = true; error.value = ''; try { await auth.login(account.value, password.value); router.push('/') } catch (e) { error.value = (e as Error).message } finally { loading.value = false } }
</script>
<template><div class="grid min-h-screen place-items-center bg-gradient-to-br from-slate-950 via-blue-950 to-slate-900 p-6"><form class="card w-full max-w-md p-8" @submit.prevent="submit"><div class="mb-8 flex items-center gap-3"><div class="grid h-12 w-12 place-items-center rounded-2xl bg-brand text-white"><Bot /></div><div><h1 class="text-2xl font-black">TelePilot Console</h1><p class="text-slate-500">企业级 Telegram Bot 管理系统</p></div></div><label class="label">管理员账号</label><input v-model="account" class="input mb-4" maxlength="6" placeholder="首次启动生成的6位账号"><label class="label">管理员密码</label><input v-model="password" class="input mb-4" type="password" placeholder="64位安全密码"><p v-if="error" class="mb-4 rounded-2xl bg-red-50 p-3 text-red-600">{{ error }}</p><button class="btn-primary w-full" :disabled="loading">{{ loading ? '校验中...' : '安全登录' }}</button></form></div></template>
