# Frontend Testing

This directory contains test utilities and configuration for the frontend application.

## Test Setup

The test setup is configured in `setup.ts` and provides:
- Mocked `localStorage` for testing
- Mocked `window.location` for routing tests
- Global Vue Test Utils configuration

## Test Utilities

The `test-utils.ts` file provides helper functions:
- `setupTestEnv()` - Sets up Pinia, i18n, and router for tests
- Pre-configured i18n instance with minimal translations
- Pre-configured router with basic routes

## Running Tests

```bash
# Run tests in watch mode
npm run test

# Run tests with UI
npm run test:ui

# Run tests once
npm run test:run

# Run tests with coverage
npm run test:coverage
```

## Writing Tests

### Component Tests

Component tests should be placed in `__tests__` directories next to the components:

```
src/
  components/
    MyComponent.vue
    __tests__/
      MyComponent.test.ts
```

### Store Tests

Store tests should be placed in `__tests__` directories next to the stores:

```
src/
  store/
    myStore.ts
    __tests__/
      myStore.test.ts
```

### Example Component Test

```typescript
import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import MyComponent from '../MyComponent.vue'
import { setupTestEnv } from '@/test/test-utils'
import { createPinia, setActivePinia } from 'pinia'

describe('MyComponent', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    setupTestEnv()
  })

  it('renders correctly', () => {
    const wrapper = mount(MyComponent)
    expect(wrapper.find('button').exists()).toBe(true)
  })
})
```

### Example Store Test

```typescript
import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useMyStore } from '../myStore'

describe('My Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('initializes correctly', () => {
    const store = useMyStore()
    expect(store.value).toBe(0)
  })
})
```

## Mocking

### API Calls

Use `vi.mock()` to mock API modules:

```typescript
vi.mock('@/api/MyApi', () => ({
  default: {
    fetchData: vi.fn(() => Promise.resolve([])),
  },
}))
```

### LocalStorage

LocalStorage is automatically mocked in the test setup. Use it directly:

```typescript
localStorage.setItem('key', 'value')
expect(localStorage.getItem('key')).toBe('value')
```

## Test Coverage

Coverage reports are generated in the `coverage/` directory. Open `coverage/index.html` in a browser to view the detailed coverage report.

