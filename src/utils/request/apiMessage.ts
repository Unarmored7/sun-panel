import { createDiscreteApi } from 'naive-ui'
import { computed } from 'vue'
import type { Response } from './index'
import { t } from '@/locales'
import { useTheme } from '@/hooks/useTheme'

// Messages outside the app tree use the same Usagi palette as the page.
const { themeOverrides } = useTheme()
export const { message } = createDiscreteApi(['message'], { configProviderProps: computed(() => ({ themeOverrides: themeOverrides.value })) })

export function apiRespErrMsg(res: Response): boolean {
  const apiErrorCodeName = `apiErrorCode.${res.code}`
  const getI18nValue = t(apiErrorCodeName)
  if (apiErrorCodeName === getI18nValue) {
    return false
  }
  else {
    message.error(t(`apiErrorCode.${res.code}`))
    return true
  }
}
