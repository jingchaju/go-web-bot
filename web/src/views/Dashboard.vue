<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AdminLayout from '../layouts/AdminLayout.vue'
import { dashboard } from '../api/admin'
const data = ref<any>({ users: 0, bot: { running: false, processed: 0 } }); onMounted(async()=> data.value = await dashboard())
</script>
<template><AdminLayout><div class="mb-8"><h1 class="text-3xl font-black">仪表盘</h1><p class="text-slate-500">实时观察服务、用户与 Bot 处理状态。</p></div><div class="grid gap-6 md:grid-cols-3"><div class="card p-6"><p class="text-slate-500">总用户</p><b class="text-4xl">{{ data.users }}</b></div><div class="card p-6"><p class="text-slate-500">Bot 状态</p><b class="text-4xl" :class="data.bot.running?'text-emerald-600':'text-slate-400'">{{ data.bot.running ? '运行中' : '未启动' }}</b></div><div class="card p-6"><p class="text-slate-500">已处理更新</p><b class="text-4xl">{{ data.bot.processed }}</b></div></div><div class="card mt-6 p-6"><h2 class="mb-4 text-xl font-bold">数据分析</h2><div class="h-56 rounded-3xl bg-gradient-to-r from-blue-50 to-cyan-50 p-6 text-slate-500">轻量趋势面板：接入真实事件后可扩展为图表。</div></div></AdminLayout></template>
