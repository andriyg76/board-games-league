import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import GameTypesListGameTypes from "@/gametypes/ListGameTypes.vue";

const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    name: 'home',
    component: HomeView
  },
  {
    path: '/ui/user',
    name: 'user',
    component: () => import('../views/UserView.vue')
  },
  {
    path: '/ui/auth-callback', // constant at backend/auth/auth.go
    name: 'auth-callback',
    component: () => import('../components/AuthCallback.vue')
  },
  {
    path: '/ui/admin/create-user', // constant at backend/auth/auth.go
    name: 'CreateUser',
    component: () => import('../components/CreateUser.vue'),
  },
  {
    path: '/ui/admin/game-types',
    name: 'CreateUser',
    component: GameTypesListGameTypes,
  },
  {
    path: '/ui/admin/diagnostics',
    name: 'Diagnostics',
    component: () => import('../views/DiagnosticsView.vue'),
  },
  {
    path: '/ui/game-rounds',
    name: 'GameRounds',
    component: () => import('@/gametypes/GameroundsList.vue')
  },
  {
    path: '/ui/leagues',
    name: 'Leagues',
    component: () => import('@/gametypes/ListLeagues.vue')
  },
  {
    path: '/ui/game-rounds/new',
    name: 'NewGameRound',
    component: () => import('@/gametypes/GameroundEdit.vue')
  },
  {
    path: '/ui/game-rounds/:id',
    name: 'EditGameRound',
    component: () => import('@/gametypes/GameroundEdit.vue'),
    props: true
  },
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes
})

export default router
