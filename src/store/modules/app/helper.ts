import { ss } from '@/utils/storage'

const LOCAL_NAME = 'appSetting'
const THEME_PREFERENCE_VERSION = 1

export type Theme = 'light' | 'dark' | 'auto'

export type Language = 'zh-CN' | 'en-US'

export interface AppState {
  siderCollapsed: boolean
  theme: Theme
  language: Language
  themePreferenceVersion?: number
}

export function defaultSetting(): AppState {
  const lan = (navigator.language).toLowerCase()
  let language: Language = 'en-US'
  if (lan.includes('zh'))
    language = 'zh-CN'

  return { siderCollapsed: false, theme: 'light', language, themePreferenceVersion: THEME_PREFERENCE_VERSION }
}

export function getLocalSetting(): AppState {
  const localSetting: AppState | undefined = ss.get(LOCAL_NAME)
  const setting = { ...defaultSetting(), ...localSetting }
  // Legacy dark/auto settings predate the Usagi palette and must not change
  // the site's appearance on upgrade. Later explicit choices remain respected.
  if (localSetting?.themePreferenceVersion !== THEME_PREFERENCE_VERSION) {
    setting.theme = 'light'
    setting.themePreferenceVersion = THEME_PREFERENCE_VERSION
    ss.set(LOCAL_NAME, setting)
  }
  return setting
}

export function setLocalSetting(setting: AppState): void {
  ss.set(LOCAL_NAME, setting)
}

export function removeLocalState() {
  ss.remove(LOCAL_NAME)
}
