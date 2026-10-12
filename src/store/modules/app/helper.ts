import { ss } from '@/utils/storage'

const LOCAL_NAME = 'appSetting'

export type Language = 'zh-CN' | 'en-US'

export interface AppState {
  siderCollapsed: boolean
  language: Language
}

export function defaultSetting(): AppState {
  const lan = (navigator.language).toLowerCase()
  let language: Language = 'en-US'
  if (lan.includes('zh'))
    language = 'zh-CN'

  return { siderCollapsed: false, language }
}

export function getLocalSetting(): AppState {
  const localSetting: Partial<AppState> | undefined = ss.get(LOCAL_NAME)
  return { ...defaultSetting(), siderCollapsed: localSetting?.siderCollapsed ?? false, language: localSetting?.language ?? defaultSetting().language }
}

export function setLocalSetting(setting: AppState): void {
  ss.set(LOCAL_NAME, setting)
}

export function removeLocalState() {
  ss.remove(LOCAL_NAME)
}
