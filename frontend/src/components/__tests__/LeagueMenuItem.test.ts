import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import LeagueMenuItem from '../LeagueMenuItem.vue'
import { setupTestEnv } from '@/test/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { useLeagueStore } from '@/store/league'
import { useUserStore } from '@/store/user'

// Mock the API
vi.mock('@/api/LeagueApi', () => ({
  default: {
    listLeagues: vi.fn(() => Promise.resolve([])),
  },
}))

describe('LeagueMenuItem', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    setupTestEnv()
    localStorage.clear()
  })

  it('does not render menu when there are 0 or 1 active leagues and user is not superadmin', () => {
    const leagueStore = useLeagueStore()
    const userStore = useUserStore()
    
    leagueStore.leagues = []
    userStore.setUser({ name: 'Test User', alias: 'test', external_ids: [] })

    const wrapper = mount(LeagueMenuItem)
    expect(wrapper.find('button').exists()).toBe(false)
  })

  it('renders menu when user is superadmin even with 0 leagues', () => {
    const leagueStore = useLeagueStore()
    const userStore = useUserStore()
    
    leagueStore.leagues = []
    userStore.setUser({ 
      name: 'Super Admin', 
      alias: 'admin', 
      external_ids: [],
      roles: ['superadmin']
    })

    const wrapper = mount(LeagueMenuItem)
    expect(wrapper.find('button').exists()).toBe(true)
  })

  it('renders menu when there are more than 1 active leagues', () => {
    const leagueStore = useLeagueStore()
    const userStore = useUserStore()
    
    leagueStore.leagues = [
      { code: 'league1', name: 'League 1', status: 'active' },
      { code: 'league2', name: 'League 2', status: 'active' },
    ] as any
    userStore.setUser({ name: 'Test User', alias: 'test', external_ids: [] })

    const wrapper = mount(LeagueMenuItem)
    expect(wrapper.find('button').exists()).toBe(true)
  })

  it('displays current league name when set', () => {
    const leagueStore = useLeagueStore()
    const userStore = useUserStore()
    
    leagueStore.leagues = [
      { code: 'league1', name: 'League 1', status: 'active' },
      { code: 'league2', name: 'League 2', status: 'active' },
    ] as any
    leagueStore.currentLeague = { code: 'league1', name: 'League 1', status: 'active' } as any
    userStore.setUser({ name: 'Test User', alias: 'test', external_ids: [] })

    const wrapper = mount(LeagueMenuItem)
    expect(wrapper.text()).toContain('League 1')
  })

  it('shows all leagues for superadmin including archived', () => {
    const leagueStore = useLeagueStore()
    const userStore = useUserStore()
    
    leagueStore.leagues = [
      { code: 'league1', name: 'League 1', status: 'active' },
      { code: 'league2', name: 'League 2', status: 'archived' },
    ] as any
    userStore.setUser({ 
      name: 'Super Admin', 
      alias: 'admin', 
      external_ids: [],
      roles: ['superadmin']
    })

    const wrapper = mount(LeagueMenuItem, {
      global: {
        stubs: {
          'n-dropdown': {
            template: '<div><slot /></div>',
            props: ['options', 'trigger'],
          },
          'n-button': {
            template: '<button><slot /></button>',
          },
          'n-icon': {
            template: '<span><slot /></span>',
          },
        },
      },
    })
    
    // Access the component instance to check menuOptions
    const component = wrapper.vm as any
    const menuOptions = component.menuOptions
    // Should have 2 leagues + divider + "All Leagues" = 4 items
    expect(menuOptions.length).toBe(4)
  })

  it('shows only active leagues for regular users', () => {
    const leagueStore = useLeagueStore()
    const userStore = useUserStore()
    
    leagueStore.leagues = [
      { code: 'league1', name: 'League 1', status: 'active' },
      { code: 'league2', name: 'League 2', status: 'archived' },
    ] as any
    userStore.setUser({ name: 'Test User', alias: 'test', external_ids: [] })

    const wrapper = mount(LeagueMenuItem, {
      global: {
        stubs: {
          'n-dropdown': {
            template: '<div><slot /></div>',
            props: ['options', 'trigger'],
          },
          'n-button': {
            template: '<button><slot /></button>',
          },
          'n-icon': {
            template: '<span><slot /></span>',
          },
        },
      },
    })
    
    // Access the component instance to check menuOptions
    const component = wrapper.vm as any
    const menuOptions = component.menuOptions
    // Should have 1 active league + divider + "All Leagues" = 3 items
    expect(menuOptions.length).toBe(3)
  })

  it('loads leagues on mount if leagues list is empty', async () => {
    const leagueStore = useLeagueStore()
    const userStore = useUserStore()
    
    leagueStore.leagues = []
    userStore.setUser({ 
      name: 'Super Admin', 
      alias: 'admin', 
      external_ids: [],
      roles: ['superadmin']
    })

    const loadLeaguesSpy = vi.spyOn(leagueStore, 'loadLeagues')
    
    mount(LeagueMenuItem)
    
    // Wait for onMounted to execute
    await new Promise(resolve => setTimeout(resolve, 0))
    
    expect(loadLeaguesSpy).toHaveBeenCalled()
  })
})

