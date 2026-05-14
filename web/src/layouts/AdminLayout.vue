<script setup lang="ts">
import { ref } from 'vue'
import { Bot, LayoutDashboard, Menu, Users, LogOut } from 'lucide-vue-next'
import { useAuthStore } from '../stores/auth'
const collapsed = ref(false); const auth = useAuthStore()
</script>
<template>
  <div class="min-h-screen bg-slate-100">
    <aside :class="['fixed inset-y-0 left-0 z-20 bg-ink text-white transition-all', collapsed ? 'w-20' : 'w-72']">
      <div class="flex h-20 items-center gap-3 px-6"><div class="grid h-11 w-11 place-items-center rounded-2xl bg-brand"><Bot /></div><div v-if="!collapsed"><b>TelePilot</b><p class="text-xs text-slate-300">Bot 管理后台</p></div></div>
      <nav class="space-y-2 px-4"><RouterLink class="flex items-center gap-3 rounded-2xl px-4 py-3 hover:bg-white/10" to="/"><LayoutDashboard /> <span v-if="!collapsed">仪表盘</span></RouterLink><RouterLink class="flex items-center gap-3 rounded-2xl px-4 py-3 hover:bg-white/10" to="/users"><Users /> <span v-if="!collapsed">用户管理</span></RouterLink><RouterLink class="flex items-center gap-3 rounded-2xl px-4 py-3 hover:bg-white/10" to="/bot"><Bot /> <span v-if="!collapsed">Telegram Bot管理</span></RouterLink></nav>
    </aside>
    <main :class="['transition-all', collapsed ? 'pl-20' : 'pl-72']">
      <header class="sticky top-0 z-10 flex h-20 items-center justify-between border-b border-slate-200 bg-white/80 px-8 backdrop-blur"><button class="btn" @click="collapsed=!collapsed"><Menu /></button><div class="flex items-center gap-4"><span class="badge">管理员 {{ auth.account }}</span><div class="grid h-11 w-11 place-items-center rounded-full bg-gradient-to-br from-blue-500 to-cyan-400 font-bold text-white">A</div><button class="btn" @click="auth.logout(); $router.push('/login')"><LogOut /></button></div></header>
      <section class="p-8"><slot /></section>
    </main>
  </div>
</template>
