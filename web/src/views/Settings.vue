<script setup lang="ts">
import { ref } from 'vue'
import { KeyRound, UserCog } from 'lucide-vue-next'
import AdminLayout from '../layouts/AdminLayout.vue'
import { updateAccount, updatePassword } from '../api/admin'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
const auth = useAuthStore(); const router = useRouter()
const account = ref(auth.account); const password = ref(''); const toast = ref(''); const error = ref(''); const loading = ref(false)
function forceLogout(message: string){ toast.value = message; setTimeout(()=>{ auth.logout(); router.push('/login') }, 1200) }
async function saveAccount(){ loading.value=true; error.value=''; try{ const res = await updateAccount(account.value); forceLogout(res.message) } catch(e){ error.value=(e as Error).message } finally{ loading.value=false } }
async function savePassword(){ loading.value=true; error.value=''; try{ const res = await updatePassword(password.value); forceLogout(res.message) } catch(e){ error.value=(e as Error).message } finally{ loading.value=false } }
</script>
<template>
  <AdminLayout>
    <div class="mb-8"><h1 class="text-4xl font-black">系统设置</h1><p class="text-slate-500">修改管理员账号或密码后会自动强制退出登录。</p></div>
    <div v-if="toast" class="alert-success mb-4">{{ toast }}</div><div v-if="error" class="alert-error mb-4">{{ error }}</div>
    <div class="grid gap-6 lg:grid-cols-2">
      <form class="card space-y-4 p-6" @submit.prevent="saveAccount"><div class="flex items-center gap-3"><UserCog class="text-blue-600" /><h2 class="text-xl font-bold">更改管理员账号</h2></div><label class="label">新账号</label><input v-model="account" class="input" placeholder="4-32 位管理员账号"><button class="btn-primary" :disabled="loading">保存账号并退出</button></form>
      <form class="card space-y-4 p-6" @submit.prevent="savePassword"><div class="flex items-center gap-3"><KeyRound class="text-cyan-600" /><h2 class="text-xl font-bold">更改管理员密码</h2></div><label class="label">新密码</label><input v-model="password" class="input" type="password" placeholder="请输入 8-128 位自定义密码"><button class="btn-primary" :disabled="loading">保存密码并退出</button></form>
    </div>
  </AdminLayout>
</template>
