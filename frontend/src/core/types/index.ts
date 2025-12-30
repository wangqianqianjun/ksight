// Core application types
export type Locale = 'en' | 'zh'
export type Theme = 'light' | 'dark'

// Export plugin system types
export * from './plugin'
export * from './layout'
export * from './views'
export * from './k8s'

// Legacy store interface (to be refactored)
export interface MainStore {
  count: number
  theme: Theme
  locale: Locale
  increment: () => void
  toggleTheme: () => void
  setLocale: (newLocale: Locale) => void
}

// I18n message structure
export interface LocaleMessages {
  hello: string
  welcome: string
  language: string
  theme: string
  [key: string]: string
}

// Note: Window global type declarations are defined in @/lib/k8s-sdk.ts
// to avoid duplicate declarations
