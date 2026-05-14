import { createRouter, createWebHistory } from 'vue-router'
import Login from '../views/Login.vue'
import Dashboard from '../views/Dashboard.vue'
import Users from '../views/Users.vue'
import BotManage from '../views/BotManage.vue'
import Settings from '../views/Settings.vue'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: Login, meta: { public: true } },
    { path: '/', component: Dashboard },
    { path: '/users', component: Users },
    { path: '/bot', component: BotManage },
    { path: '/settings', component: Settings },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.meta.public) return auth.token ? '/' : true
  if (!auth.token) return '/login'
  try {
    if (!auth.account) await auth.hydrate()
    return true
  } catch {
    auth.logout()
    return '/login'
  }
})
export default router
