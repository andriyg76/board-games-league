import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useLeagueStore } from '../league'
import LeagueApi from '@/api/LeagueApi'

// Mock the API
vi.mock('@/api/LeagueApi', () => ({
  default: {
    listLeagues: vi.fn(),
    getLeague: vi.fn(),
    getLeagueMembers: vi.fn(),
    getLeagueStandings: vi.fn(),
    createLeague: vi.fn(),
  },
}))

describe('League Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('initializes with empty state', () => {
    const store = useLeagueStore()
    expect(store.leagues).toEqual([])
    expect(store.currentLeague).toBeNull()
    expect(store.currentLeagueMembers).toEqual([])
    expect(store.currentLeagueStandings).toEqual([])
    expect(store.loading).toBe(false)
    expect(store.error).toBeNull()
  })

  it('loads leagues successfully', async () => {
    const store = useLeagueStore()
    const mockLeagues = [
      { code: 'league1', name: 'League 1', status: 'active' },
      { code: 'league2', name: 'League 2', status: 'active' },
    ]

    vi.mocked(LeagueApi.listLeagues).mockResolvedValue(mockLeagues as any)

    await store.loadLeagues()

    expect(store.leagues).toEqual(mockLeagues)
    expect(store.loading).toBe(false)
    expect(store.error).toBeNull()
  })

  it('handles error when loading leagues fails', async () => {
    const store = useLeagueStore()
    const error = new Error('Failed to load leagues')

    vi.mocked(LeagueApi.listLeagues).mockRejectedValue(error)

    await expect(store.loadLeagues()).rejects.toThrow('Failed to load leagues')
    expect(store.error).toBe('Failed to load leagues')
    expect(store.loading).toBe(false)
  })

  it('creates a new league', async () => {
    const store = useLeagueStore()
    const newLeague = { code: 'new-league', name: 'New League', status: 'active' }

    vi.mocked(LeagueApi.createLeague).mockResolvedValue(newLeague as any)

    const result = await store.createLeague('New League')

    expect(result).toEqual(newLeague)
    expect(store.leagues).toContainEqual(newLeague)
    expect(store.loading).toBe(false)
  })

  it('sets current league and loads related data', async () => {
    const store = useLeagueStore()
    const league = { code: 'league1', name: 'League 1', status: 'active' }
    const members = [{ user_id: 'user1', alias: 'Player 1', status: 'active' }]
    const standings = [{ user_id: 'user1', total_points: 100 }]

    vi.mocked(LeagueApi.getLeague).mockResolvedValue(league as any)
    vi.mocked(LeagueApi.getLeagueMembers).mockResolvedValue(members as any)
    vi.mocked(LeagueApi.getLeagueStandings).mockResolvedValue(standings as any)

    await store.setCurrentLeague('league1')

    expect(store.currentLeague).toEqual(league)
    expect(store.currentLeagueMembers).toEqual(members)
    expect(store.currentLeagueStandings).toEqual(standings)
    expect(localStorage.getItem('currentLeagueCode')).toBe('league1')
  })

  it('clears current league', () => {
    const store = useLeagueStore()
    store.currentLeague = { code: 'league1', name: 'League 1' } as any
    store.currentLeagueMembers = [{ user_id: 'user1' }] as any
    store.currentLeagueStandings = [{ user_id: 'user1' }] as any
    localStorage.setItem('currentLeagueCode', 'league1')

    store.clearCurrentLeague()

    expect(store.currentLeague).toBeNull()
    expect(store.currentLeagueMembers).toEqual([])
    expect(store.currentLeagueStandings).toEqual([])
    expect(localStorage.getItem('currentLeagueCode')).toBeNull()
  })

  it('currentLeagueCode getter returns code from state', () => {
    const store = useLeagueStore()
    store.currentLeague = { code: 'league1', name: 'League 1' } as any

    expect(store.currentLeagueCode).toBe('league1')
  })

  it('currentLeagueCode getter falls back to localStorage', () => {
    const store = useLeagueStore()
    store.currentLeague = null
    localStorage.setItem('currentLeagueCode', 'saved-league')

    expect(store.currentLeagueCode).toBe('saved-league')
  })

  it('activeLeagues getter filters only active leagues', () => {
    const store = useLeagueStore()
    store.leagues = [
      { code: 'league1', name: 'League 1', status: 'active' },
      { code: 'league2', name: 'League 2', status: 'archived' },
      { code: 'league3', name: 'League 3', status: 'active' },
    ] as any

    const activeLeagues = store.activeLeagues

    expect(activeLeagues).toHaveLength(2)
    expect(activeLeagues.every(league => league.status === 'active')).toBe(true)
  })

  it('archivedLeagues getter filters only archived leagues', () => {
    const store = useLeagueStore()
    store.leagues = [
      { code: 'league1', name: 'League 1', status: 'active' },
      { code: 'league2', name: 'League 2', status: 'archived' },
      { code: 'league3', name: 'League 3', status: 'archived' },
    ] as any

    const archivedLeagues = store.archivedLeagues

    expect(archivedLeagues).toHaveLength(2)
    expect(archivedLeagues.every(league => league.status === 'archived')).toBe(true)
  })

  it('getLeagueByCode returns league with matching code', () => {
    const store = useLeagueStore()
    store.leagues = [
      { code: 'league1', name: 'League 1', status: 'active' },
      { code: 'league2', name: 'League 2', status: 'active' },
    ] as any

    const league = store.getLeagueByCode('league1')

    expect(league).toEqual({ code: 'league1', name: 'League 1', status: 'active' })
  })

  it('getLeagueByCode returns undefined for non-existent code', () => {
    const store = useLeagueStore()
    store.leagues = [
      { code: 'league1', name: 'League 1', status: 'active' },
    ] as any

    const league = store.getLeagueByCode('non-existent')

    expect(league).toBeUndefined()
  })

  it('getTopPlayers returns top N players sorted by points', () => {
    const store = useLeagueStore()
    store.currentLeagueStandings = [
      { user_id: 'user1', total_points: 50 },
      { user_id: 'user2', total_points: 100 },
      { user_id: 'user3', total_points: 75 },
      { user_id: 'user4', total_points: 200 },
    ] as any

    const topPlayers = store.getTopPlayers(2)

    expect(topPlayers).toHaveLength(2)
    expect(topPlayers[0].total_points).toBe(200)
    expect(topPlayers[1].total_points).toBe(100)
  })
})

