<script setup lang="ts">
import { ref } from 'vue'
import { Bot, LayoutDashboard, Menu, Users, LogOut, Settings, ShieldCheck } from 'lucide-vue-next'
import { useAuthStore } from '../stores/auth'
const collapsed = ref(false)
const auth = useAuthStore()
</script>
<template>
  <div class="min-h-screen overflow-hidden bg-[radial-gradient(circle_at_top_left,#dbeafe,transparent_35%),linear-gradient(135deg,#f8fafc,#e0f2fe)] text-slate-900">
    <aside :class="['fixed inset-y-0 left-0 z-20 border-r border-white/20 bg-slate-950/80 text-white shadow-2xl backdrop-blur-2xl transition-all', collapsed ? 'w-20' : 'w-72']">
      <div class="flex h-20 items-center gap-3 px-6">
        <div class="grid h-11 w-11 place-items-center rounded-2xl bg-gradient-to-br from-blue-500 to-cyan-400 shadow-lg shadow-blue-500/30"><Bot /></div>
        <div v-if="!collapsed"><b class="text-lg">TelePilot</b><p class="text-xs text-slate-300">Bot 管理后台</p></div>
      </div>
      <nav class="space-y-2 px-4">
        <RouterLink class="nav-link" to="/"><LayoutDashboard /> <span v-if="!collapsed">仪表盘</span></RouterLink>
        <RouterLink class="nav-link" to="/users"><Users /> <span v-if="!collapsed">用户管理</span></RouterLink>
        <RouterLink class="nav-link" to="/bot"><Bot /> <span v-if="!collapsed">Telegram Bot 管理</span></RouterLink>
        <RouterLink class="nav-link" to="/settings"><Settings /> <span v-if="!collapsed">系统设置</span></RouterLink>
      </nav>
      <div v-if="!collapsed" class="absolute bottom-5 left-4 right-4 rounded-3xl border border-white/10 bg-white/10 p-4 text-xs text-slate-200">
        <ShieldCheck class="mb-2 h-5 w-5 text-cyan-300" />
        已启用 JWT、指纹校验、请求签名和安全响应头。
      </div>
    </aside>
    <main :class="['transition-all', collapsed ? 'pl-20' : 'pl-72']">
      <header class="sticky top-0 z-10 flex h-20 items-center justify-between border-b border-white/40 bg-white/55 px-8 shadow-sm backdrop-blur-2xl">
        <button class="btn bg-white/70" @click="collapsed=!collapsed"><Menu /></button>
        <div class="flex items-center gap-4">
          <span class="badge">管理员 {{ auth.account }}</span>
          <div class="grid h-11 w-11 place-items-center rounded-full bg-gradient-to-br from-blue-500 to-cyan-400 font-bold text-white shadow-lg">A</div>
          <button class="btn bg-white/70" @click="auth.logout(); $router.push('/login')"><LogOut /></button>
        </div>
      </header>
      <section class="p-8"><slot /></section>
    </main>
  </div>
</template>
