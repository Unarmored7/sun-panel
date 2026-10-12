import { ss } from '@/utils/storage'

const LOCAL_NAME = 'appSetting'

export type Language = 'zh-CN' | 'en-US'

export interface AppState {
  siderCollapsed: boolean
  language: Language
  languageVersion?: number
}

// The site is Chinese; English stays available as an explicit choice in settings.
const LANGUAGE_VERSION = 1

export function defaultSetting(): AppState {
  return { siderCollapsed: false, language: 'zh-CN', languageVersion: LANGUAGE_VERSION }
}

export function getLocalSetting(): AppState {
  const localSetting: Partial<AppState> | undefined = ss.get(LOCAL_NAME)
  // Older versions stored the browser language automatically; reset that once.
  if (localSetting?.languageVersion !== LANGUAGE_VERSION)
    return defaultSetting()
  return { ...defaultSetting(), siderCollapsed: localSetting.siderCollapsed ?? false, language: localSetting.language ?? 'zh-CN' }
}

export function setLocalSetting(setting: AppState): void {
  ss.set(LOCAL_NAME, setting)
}

export function removeLocalState() {
  ss.remove(LOCAL_NAME)
}
