<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { CheckCircle2, Play, Save, Square, WandSparkles } from 'lucide-vue-next'
import AdminLayout from '../layouts/AdminLayout.vue'
import { getBotConfig, saveBotConfig, startBot, stopBot, type BotConfig } from '../api/admin'
const form = ref<BotConfig>({ bot_token:'', webhook_port:'8080', webhook_path:'/telegram/webhook', webhook_secret:'', commands_json:'[{"command":"start","description":"开始使用"}]', keyboard_json:'[]', welcome_html:'<b>欢迎使用 Telegram Bot</b>', bot_running:false })
const toast = ref('')
const error = ref('')
const loading = ref(false)
async function load(){ form.value = await getBotConfig() }
function notify(message: string){ toast.value = message; error.value = ''; window.setTimeout(()=> toast.value='', 3500) }
function fail(e: unknown){ error.value = (e as Error).message; toast.value = '' }
async function save(){ loading.value=true; try{ form.value = await saveBotConfig(form.value); await load(); notify('配置已保存并刷新') } catch(e){ fail(e) } finally{ loading.value=false } }
async function toggle(){ loading.value=true; try{ form.value.bot_running ? await stopBot() : await startBot(); await load(); notify(form.value.bot_running ? 'Bot 已成功启动，Webhook 已注册' : 'Bot 已停止') } catch(e){ fail(e) } finally{ loading.value=false } }
function fillExample(){ form.value.commands_json = JSON.stringify([{command:'start',description:'开始使用'},{command:'help',description:'查看帮助'},{command:'contact',description:'联系客服'}], null, 2); form.value.keyboard_json = JSON.stringify([{text:'官网',url:'https://example.com'},{text:'联系客服',url:'https://t.me/your_support'},{text:'帮助',data:'help'},{text:'购买套餐',data:'buy_plan'},{text:'查询余额',data:'check_balance'},{text:'邀请好友',data:'invite'}], null, 2) }
onMounted(load)
</script>
<template>
  <AdminLayout>
    <div class="mb-8 flex flex-wrap items-center justify-between gap-4">
      <div><h1 class="text-4xl font-black tracking-tight">Telegram Bot 管理</h1><p class="text-slate-500">保存 token、webhook、指令、键盘与欢迎语后再启动 Bot。</p></div>
      <span class="badge text-sm" :class="form.bot_running ? 'bg-emerald-50 text-emerald-700 ring-emerald-100' : ''"><CheckCircle2 class="mr-1 inline h-4 w-4" />{{ form.bot_running ? '实时运行' : '等待启动' }}</span>
    </div>
    <div v-if="toast" class="alert-success mb-4">{{ toast }}</div>
    <div v-if="error" class="alert-error mb-4">{{ error }}</div>
    <div class="grid gap-6 lg:grid-cols-3">
      <form class="card space-y-5 p-6 lg:col-span-2" @submit.prevent="save">
        <div class="flex justify-end"><button type="button" class="btn bg-white/80 text-blue-700" @click="fillExample"><WandSparkles class="h-4 w-4" />填入示例 JSON</button></div>
        <div class="grid gap-4 md:grid-cols-2"><div><label class="label">Bot Token</label><input v-model="form.bot_token" class="input" placeholder="123456:ABC"></div><div><label class="label">Webhook SECRET</label><input v-model="form.webhook_secret" class="input" placeholder="高强度密钥"></div><div><label class="label">Webhook 端口</label><input v-model="form.webhook_port" class="input"></div><div><label class="label">Webhook 路径或完整 HTTPS URL</label><input v-model="form.webhook_path" class="input" placeholder="/telegram 或 https://域名/telegram"><p class="mt-1 text-xs text-slate-500">填写相对路径时，请在后端 .env 设置 PUBLIC_BASE_URL=https://你的域名。</p></div></div>
        <label class="label">Commands 指令 JSON</label><textarea v-model="form.commands_json" class="input h-36 font-mono text-sm"></textarea>
        <label class="label">内联回复键盘 JSON（三个一组）</label><textarea v-model="form.keyboard_json" class="input h-44 font-mono text-sm"></textarea>
        <label class="label">/start 欢迎消息 HTML</label><textarea v-model="form.welcome_html" class="input h-40"></textarea>
        <div class="flex flex-wrap gap-3"><button class="btn-primary" :disabled="loading"><Save class="h-4 w-4" />保存配置</button><button type="button" class="btn bg-slate-900 text-white" @click="toggle" :disabled="loading"><Square v-if="form.bot_running" class="h-4 w-4" /><Play v-else class="h-4 w-4" />{{ form.bot_running ? '停止 Bot' : '确认启动 Bot' }}</button></div>
      </form>
      <aside class="card p-6"><h2 class="mb-4 text-xl font-bold">实时监控</h2><div class="space-y-4"><div class="rounded-3xl bg-white/60 p-4"><p class="text-slate-500">Webhook</p><b class="break-all">{{ form.webhook_path }}</b></div><div class="rounded-3xl bg-white/60 p-4"><p class="text-slate-500">运行状态</p><b :class="form.bot_running?'text-emerald-600':'text-slate-400'">{{ form.bot_running ? '在线' : '离线' }}</b></div><div class="rounded-3xl bg-blue-50/80 p-4 text-blue-700">启动时会自动调用 Telegram setWebhook，并将 Commands 与内联键盘注入 Bot。</div></div></aside>
    </div>
  </AdminLayout>
</template>
