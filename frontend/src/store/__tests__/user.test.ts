import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useUserStore } from '../user'

describe('User Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('initializes with empty user and loggedIn false', () => {
    const store = useUserStore()
    expect(store.user.name).toBe('')
    expect(store.user.alias).toBe('')
    expect(store.loggedIn).toBe(false)
  })

  it('sets user and updates loggedIn status', () => {
    const store = useUserStore()
    const user = {
      name: 'Test User',
      alias: 'testuser',
      external_ids: ['123'],
    }

    store.setUser(user)

    expect(store.user.name).toBe('Test User')
    expect(store.user.alias).toBe('testuser')
    expect(store.loggedIn).toBe(true)
  })

  it('sets loggedIn to false when user has no external_ids', () => {
    const store = useUserStore()
    const user = {
      name: 'Test User',
      alias: 'testuser',
      external_ids: [],
    }

    store.setUser(user)

    expect(store.loggedIn).toBe(false)
  })

  it('clears user and resets loggedIn', () => {
    const store = useUserStore()
    store.setUser({
      name: 'Test User',
      alias: 'testuser',
      external_ids: ['123'],
    })

    store.clearUser()

    expect(store.user.name).toBe('')
    expect(store.user.alias).toBe('')
    expect(store.loggedIn).toBe(false)
  })

  it('isAuthenticated getter returns loggedIn status', () => {
    const store = useUserStore()
    expect(store.isAuthenticated).toBe(false)

    store.setUser({
      name: 'Test User',
      alias: 'testuser',
      external_ids: ['123'],
    })

    expect(store.isAuthenticated).toBe(true)
  })

  it('currentUser getter returns user', () => {
    const store = useUserStore()
    const user = {
      name: 'Test User',
      alias: 'testuser',
      external_ids: ['123'],
    }

    store.setUser(user)

    expect(store.currentUser).toEqual(user)
  })

  it('isSuperAdmin returns true when user has superadmin role', () => {
    const store = useUserStore()
    store.setUser({
      name: 'Super Admin',
      alias: 'admin',
      external_ids: ['123'],
      roles: ['superadmin'],
    })

    expect(store.isSuperAdmin).toBe(true)
  })

  it('isSuperAdmin returns false when user does not have superadmin role', () => {
    const store = useUserStore()
    store.setUser({
      name: 'Regular User',
      alias: 'user',
      external_ids: ['123'],
      roles: ['user'],
    })

    expect(store.isSuperAdmin).toBe(false)
  })

  it('isSuperAdmin returns false when user has no roles', () => {
    const store = useUserStore()
    store.setUser({
      name: 'Regular User',
      alias: 'user',
      external_ids: ['123'],
    })

    expect(store.isSuperAdmin).toBe(false)
  })

  it('hasRole returns true when user has the role', () => {
    const store = useUserStore()
    store.setUser({
      name: 'User',
      alias: 'user',
      external_ids: ['123'],
      roles: ['moderator', 'user'],
    })

    expect(store.hasRole('moderator')).toBe(true)
    expect(store.hasRole('user')).toBe(true)
  })

  it('hasRole returns false when user does not have the role', () => {
    const store = useUserStore()
    store.setUser({
      name: 'User',
      alias: 'user',
      external_ids: ['123'],
      roles: ['user'],
    })

    expect(store.hasRole('moderator')).toBe(false)
  })
})

