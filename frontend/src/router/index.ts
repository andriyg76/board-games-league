import { createRouter, createWebHistory, RouteLocationNormalized, RouteRecordRaw } from 'vue-router'
import { i18n } from '@/i18n'

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
    component: () => import('@/gametypes/GameRoundWizard.vue')
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
        component: () => import('@/mobile/views/MobileGameStart.vue')
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

type UiMode = 'mobile' | 'desktop'

const UI_MODE_KEY = 'ui_mode'

const getStoredMode = (): UiMode | null => {
  if (typeof window === 'undefined') return null
  try {
    const value = sessionStorage.getItem(UI_MODE_KEY)
    if (value === 'mobile' || value === 'desktop') {
      return value
    }
  } catch (error) {
    console.warn('Unable to read ui_mode from sessionStorage:', error)
  }
  return null
}

const setStoredMode = (mode: UiMode) => {
  if (typeof window === 'undefined') return
  try {
    sessionStorage.setItem(UI_MODE_KEY, mode)
  } catch (error) {
    console.warn('Unable to write ui_mode to sessionStorage:', error)
  }
}

const getCurrentMode = (path: string): UiMode | null => {
  if (path.startsWith('/m')) return 'mobile'
  if (path.startsWith('/ui')) return 'desktop'
  return null
}

const isMobileDevice = () => {
  if (typeof window === 'undefined' || typeof navigator === 'undefined') return false
  const widthMatch = window.matchMedia?.('(max-width: 768px)').matches ?? false
  const ua = navigator.userAgent || ''
  const uaMatch = /Android|iPhone|iPad|iPod/i.test(ua)
  const touchMatch = navigator.maxTouchPoints > 0
  return widthMatch || uaMatch || touchMatch
}

const getPreferredMode = (): UiMode => (isMobileDevice() ? 'mobile' : 'desktop')

const getConfirmMessage = (targetMode: UiMode) => {
  const modeLabel = i18n.global.t(`ui.mode.${targetMode}`)
  return i18n.global.t('ui.switchPrompt', { mode: modeLabel })
}

const routeMappings: Record<UiMode, Record<string, (to: RouteLocationNormalized) => any>> = {
  mobile: {
    AcceptInvitation: (to) => ({ name: 'MobileAcceptInvite', params: to.params, query: to.query }),
    NewGameRound: () => ({ name: 'MobileGameStart' }),
    WizardGame: (to) => ({ name: 'MobileGameFlow', params: { code: to.params.code }, query: to.query }),
    Leagues: () => ({ name: 'MobileLeagueSelect' }),
    LeagueDetails: () => ({ name: 'MobileLeagueHome' }),
    Home: () => ({ name: 'MobileEntry' }),
    User: () => ({ name: 'MobileLogin' }),
  },
  desktop: {
    MobileAcceptInvite: (to) => ({ name: 'AcceptInvitation', params: to.params, query: to.query }),
    MobileGameStart: () => ({ name: 'NewGameRound' }),
    MobileGameFlow: (to) => ({ name: 'WizardGame', params: { code: to.params.code }, query: to.query }),
    MobileLeagueSelect: () => ({ name: 'Leagues' }),
    MobileLeagueHome: () => ({ name: 'Leagues' }),
    MobileEntry: () => ({ name: 'Home' }),
    MobileLogin: () => ({ name: 'Home' }),
  },
}

const getRedirectTarget = (targetMode: UiMode, to: RouteLocationNormalized) => {
  const name = to.name ? String(to.name) : ''
  const mapper = routeMappings[targetMode][name]
  if (mapper) {
    return mapper(to)
  }
  return targetMode === 'mobile' ? { name: 'MobileEntry' } : { name: 'Home' }
}

const shouldSkipAutoMode = (path: string) => {
  return path.startsWith('/ui/auth-callback')
}

router.beforeEach((to) => {
  if (typeof window === 'undefined') return true
  if (shouldSkipAutoMode(to.path)) return true

  const currentMode = getCurrentMode(to.path)
  const storedMode = getStoredMode()
  const preferredMode = getPreferredMode()

  if (!currentMode) {
    if (storedMode === 'mobile') {
      return getRedirectTarget('mobile', to)
    }
    if (!storedMode && preferredMode === 'mobile') {
      const confirmed = window.confirm(getConfirmMessage('mobile'))
      if (confirmed) {
        setStoredMode('mobile')
        return getRedirectTarget('mobile', to)
      }
      setStoredMode('desktop')
    } else if (!storedMode) {
      setStoredMode('desktop')
    }
    return true
  }

  if (storedMode === currentMode) {
    return true
  }

  const currentFitsDevice = currentMode === preferredMode
  if (currentFitsDevice) {
    setStoredMode(currentMode)
    return true
  }

  const confirmed = window.confirm(getConfirmMessage(preferredMode))
  if (confirmed) {
    setStoredMode(preferredMode)
    return getRedirectTarget(preferredMode, to)
  }
  setStoredMode(currentMode)
  return true
})

export default router
