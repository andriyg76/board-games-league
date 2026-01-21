/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<object, object, unknown>
  export default component
}

interface ImportMetaEnv {
  readonly VITE_BASE_URL: string
  readonly VITE_API_BASE_URL?: string
  readonly VITE_PUBLIC_WEB_BASE_URL?: string
  readonly VITE_ELECTRON?: string
  // add more env variables here as needed
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
