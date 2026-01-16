import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import LanguageSwitcher from '../LanguageSwitcher.vue'
import { setupTestEnv } from '@/test/test-utils'
import { createPinia, setActivePinia } from 'pinia'

describe('LanguageSwitcher', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    setupTestEnv()
    // Clear localStorage before each test
    localStorage.clear()
  })

  it('renders the language switcher button', () => {
    const wrapper = mount(LanguageSwitcher)
    expect(wrapper.find('button').exists()).toBe(true)
  })

  it('displays the current locale flag', () => {
    const wrapper = mount(LanguageSwitcher)
    const img = wrapper.find('img')
    expect(img.exists()).toBe(true)
    expect(img.attributes('src')).toBe('/flags/en.svg')
  })

  it('loads locale from localStorage on mount', () => {
    localStorage.setItem('locale', 'uk')
    const wrapper = mount(LanguageSwitcher)
    const img = wrapper.find('img')
    expect(img.attributes('src')).toBe('/flags/uk.svg')
  })

  it('changes locale when dropdown option is selected', async () => {
    const wrapper = mount(LanguageSwitcher, {
      global: {
        stubs: {
          'n-dropdown': {
            template: '<div><slot /></div>',
            props: ['options', 'trigger'],
            emits: ['select'],
          },
          'n-button': {
            template: '<button><slot /></button>',
          },
        },
      },
    })
    
    // Get the component instance to call changeLocale directly
    const component = wrapper.vm as any
    component.changeLocale('et')
    await wrapper.vm.$nextTick()

    // Check that locale was saved to localStorage
    expect(localStorage.getItem('locale')).toBe('et')
    
    // Check that the flag image updated
    const img = wrapper.find('img')
    expect(img.attributes('src')).toBe('/flags/et.svg')
  })

  it('has all three locale options available', () => {
    const wrapper = mount(LanguageSwitcher, {
      global: {
        stubs: {
          'n-dropdown': {
            template: '<div><slot /></div>',
            props: ['options', 'trigger'],
          },
          'n-button': {
            template: '<button><slot /></button>',
          },
        },
      },
    })
    
    // Access the component instance to check localeOptions
    const component = wrapper.vm as any
    const options = component.localeOptions
    expect(options).toHaveLength(3)
    expect(options.map((opt: any) => opt.key)).toEqual(['en', 'uk', 'et'])
  })
})

