import { createRouter, createWebHistory } from 'vue-router'
import Login from '../views/Login.vue'
import Dashboard from '../views/Dashboard.vue'
import Users from '../views/Users.vue'
import BotManage from '../views/BotManage.vue'
const router = createRouter({ history: createWebHistory(), routes: [{ path: '/login', component: Login }, { path: '/', component: Dashboard }, { path: '/users', component: Users }, { path: '/bot', component: BotManage }] })
router.beforeEach((to) => { if (to.path !== '/login' && !localStorage.getItem('token')) return '/login' })
export default router
