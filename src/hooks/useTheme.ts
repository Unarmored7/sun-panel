import type { GlobalThemeOverrides } from 'naive-ui'
import { computed, watch } from 'vue'
import { darkTheme, useOsTheme } from 'naive-ui'
import { useAppStore } from '@/store'

export function useTheme() {
  const appStore = useAppStore()

  const OsTheme = useOsTheme()

  const isDark = computed(() => {
    if (appStore.theme === 'auto')
      return OsTheme.value === 'dark'
    else
      return appStore.theme === 'dark'
  })

  const theme = computed(() => {
    return isDark.value ? darkTheme : undefined
  })

  const themeOverrides = computed<GlobalThemeOverrides>(() => {
    if (isDark.value) {
      return {
        common: {},
      }
    }
    const tokens = getComputedStyle(document.documentElement)
    const token = (name: string, fallback: string) => tokens.getPropertyValue(name).trim() || fallback
    return {
      common: {
        primaryColor: token('--usagi-pink-dark', '#C2597F'), primaryColorHover: token('--usagi-pink-hover', '#F28BB6'), primaryColorPressed: token('--usagi-pink-deep', '#9A284E'), primaryColorSuppl: token('--usagi-pink-dark', '#C2597F'),
        bodyColor: token('--usagi-bg-base', '#FAF2E6'), cardColor: token('--usagi-bg-card', '#FFFFFF'), modalColor: token('--usagi-bg-card', '#FFFFFF'), popoverColor: token('--usagi-bg-alt', '#FFFDF9'),
        textColorBase: token('--usagi-text-main', '#3A3131'), textColor1: token('--usagi-text-main', '#3A3131'), textColor2: token('--usagi-text-main', '#3A3131'), textColor3: token('--usagi-text-muted', 'rgba(58,49,49,.65)'),
        borderColor: token('--usagi-brown-border', '#E3D8C8'), borderRadius: token('--usagi-radius-md', '6px'), fontFamily: token('--usagi-font-body', 'sans-serif'), lineHeight: '1.7',
      },
      Card: { borderRadius: token('--usagi-radius-lg', '10px') },
      Button: { borderRadiusMedium: token('--usagi-radius-md', '6px') },
    }
  })

  watch(
    () => isDark.value,
    (dark) => {
      if (dark)
        document.documentElement.classList.add('dark')
      else
        document.documentElement.classList.remove('dark')
    },
    { immediate: true },
  )

  return { theme, themeOverrides }
}
