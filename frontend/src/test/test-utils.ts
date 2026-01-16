import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import { config } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'

// Create a minimal i18n instance for tests
const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  messages: {
    en: {
      leagues: {
        menu: 'Leagues',
        noActiveLeagues: 'No active leagues',
        archived: 'Archived',
      },
      nav: {
        leagues: 'Leagues',
      },
    },
  },
})

// Create a minimal router for tests
const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'Home', component: { template: '<div>Home</div>' } },
    { path: '/ui/leagues', name: 'Leagues', component: { template: '<div>Leagues</div>' } },
    { path: '/ui/leagues/:code', name: 'LeagueDetails', component: { template: '<div>League</div>' } },
  ],
})

// Global test utilities
export function setupTestEnv() {
  // Setup Pinia
  setActivePinia(createPinia())

  // Setup global plugins
  config.global.plugins = [i18n, router]
}

export { i18n, router }

