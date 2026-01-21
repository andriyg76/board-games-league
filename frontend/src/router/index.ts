import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'

const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/HomeView.vue')
  },
  {
    path: '/ui/user',
    name: 'User',
    component: () => import('../views/UserView.vue')
  },
  {
    path: '/ui/auth-callback', // constant at backend/auth/auth.go
    name: 'AuthCallback',
    component: () => import('../components/AuthCallback.vue')
  },
  {
    path: '/ui/admin/create-user', // constant at backend/auth/auth.go
    name: 'CreateUser',
    component: () => import('../components/CreateUser.vue'),
  },
  {
    path: '/ui/admin/game-types',
    name: 'GameTypes',
    component: () => import('@/gametypes/ListGameTypes.vue'),
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
    component: () => import('../views/LeagueList.vue')
  },
  {
    path: '/ui/leagues/:code',
    name: 'LeagueDetails',
    component: () => import('../views/LeagueDetails.vue'),
    props: true
  },
  {
    path: '/ui/leagues/join/:token',
    name: 'AcceptInvitation',
    component: () => import('../views/AcceptInvitation.vue'),
    props: true
  },
  {
    path: '/ui/game-rounds/new',
    name: 'NewGameRound',
    component: () => import('@/gametypes/GameRoundWizard.vue'),
    alias: '/ui/startgame'
  },
  {
    path: '/ui/game-rounds/:id',
    name: 'EditGameRound',
    component: () => import('@/gametypes/GameRoundWizard.vue'),
    props: true
  },
  {
    path: '/ui/game-rounds/:id/edit',
    name: 'EditCompletedGameRound',
    component: () => import('@/gametypes/GameroundEdit.vue'),
    props: true
  },
  {
    path: '/ui/wizard/:code',
    name: 'WizardGame',
    component: () => import('@/wizard/WizardGamePlay.vue'),
    props: true
  },
  {
    path: '/m',
    component: () => import('@/mobile/layouts/MobileRoot.vue'),
    meta: { layout: 'mobile' },
    children: [
      {
        path: '',
        name: 'MobileEntry',
        component: () => import('@/mobile/views/MobileEntry.vue')
      },
      {
        path: 'login',
        name: 'MobileLogin',
        component: () => import('@/mobile/views/MobileLogin.vue')
      },
      {
        path: 'accept-invite/:token',
        name: 'MobileAcceptInvite',
        component: () => import('@/mobile/views/MobileAcceptInvite.vue'),
        props: true
      },
      {
        path: 'league/select',
        name: 'MobileLeagueSelect',
        component: () => import('@/mobile/views/MobileLeagueSelect.vue')
      },
      {
        path: 'league',
        name: 'MobileLeagueHome',
        component: () => import('@/mobile/views/MobileLeagueHome.vue')
      },
      {
        path: 'game/start',
        name: 'MobileGameStart',
        component: () => import('@/mobile/views/MobileGameStart.vue'),
        alias: '/m/startgame'
      },
      {
        path: 'game/:code',
        name: 'MobileGameFlow',
        component: () => import('@/mobile/views/MobileGameFlow.vue'),
        props: true
      },
    ]
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

export default router
