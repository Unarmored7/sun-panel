<script setup lang="ts">
import { NConfigProvider } from 'naive-ui'
import NaiveProvider from '@/components/common/NaiveProvider/index.vue'
import { useTheme } from '@/hooks/useTheme'
import { useLanguage } from '@/hooks/useLanguage'

const { themeOverrides } = useTheme()
const { language } = useLanguage()
</script>

<template>
  <NConfigProvider
    class="usagi-app"
    :theme-overrides="themeOverrides"
    :locale="language"
  >
    <NaiveProvider>
      <RouterView />
    </NaiveProvider>
  </NConfigProvider>
</template>
