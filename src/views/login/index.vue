<script setup lang="ts">
import { NButton, NCard, NForm, NFormItem, NInput, NSelect, useMessage } from 'naive-ui'
import { ref } from 'vue'
import { login } from '@/api'
import { useAppStore, useAuthStore } from '@/store'
import { SvgIcon } from '@/components/common'
import { router } from '@/router'
import { t } from '@/locales'
import UsagiFooter from '@/components/usagi/UsagiFooter.vue'
import { languageOptions } from '@/utils/defaultData'
import type { Language } from '@/store/modules/app/helper'

// const userStore = useUserStore()
const authStore = useAuthStore()
const appStore = useAppStore()
const ms = useMessage()
const loading = ref(false)
const languageValue = ref<Language>(appStore.language)

// const isShowCaptcha = ref<boolean>(false)
// const isShowRegister = ref<boolean>(false)

const form = ref<Login.LoginReqest>({
  username: '',
  password: '',
})

const loginPost = async () => {
  loading.value = true
  try {
    const res = await login<Login.LoginResponse>(form.value)
    if (res.code === 0) {
      authStore.setToken(res.data.token)
      authStore.setUserInfo(res.data)

      setTimeout(() => {
        ms.success(`Hi ${res.data.name},${t('login.welcomeMessage')}`)
        loading.value = false
        router.push({ path: '/' })
      }, 500)
    }
    else {
      loading.value = false
      // captchaRef.value.refresh()
    }
  }
  catch (error) {
    loading.value = false
    // 请检查网络或者服务器错误
    console.log(error)
  }
}

function handleSubmit() {
  // 点击登录按钮触发
  loginPost()
}

function handleChangeLanuage(value: Language) {
  languageValue.value = value
  appStore.setLanguage(value)
}
</script>

<template>
  <div class="usagi-page">
    <main class="login-container usagi-content">
      <NCard class="login-card" style="border-radius: var(--usagi-radius-lg);">
        <div class="mb-5 flex items-center justify-end">
          <div class="mr-2">
            <SvgIcon icon="ion-language" style="width: 20;height: 20;" />
          </div>
          <div class="min-w-[100px]">
            <NSelect v-model:value="languageValue" size="small" :options="languageOptions" @update-value="handleChangeLanuage" />
          </div>
        </div>

        <div class="login-title  ">
          <h1>
            {{ $t('common.appName') }}
          </h1>
        </div>
        <NForm :model="form" label-width="100px" @keydown.enter="handleSubmit">
          <NFormItem>
            <NInput v-model:value="form.username" :placeholder="$t('login.usernamePlaceholder')">
              <template #prefix>
                <SvgIcon icon="ph:user-bold" />
              </template>
            </NInput>
          </NFormItem>

          <NFormItem>
            <NInput v-model:value="form.password" type="password" :placeholder="$t('login.passwordPlaceholder')">
              <template #prefix>
                <SvgIcon icon="mdi:password-outline" />
              </template>
            </NInput>
          </NFormItem>

          <!-- <NFormItem v-if="isShowCaptcha">
          <div class="w-[120px] h-[34px] mr-[20px] rounded border flex cursor-pointer">
            <Captcha ref="captchaRef" src="/api/captcha/getImage" />
          </div>
          <NInput v-model:value="form.vcode" type="text" placeholder="请输入图像验证码" />
        </NFormItem> -->
          <NFormItem style="margin-top: 10px">
            <NButton type="primary" block :loading="loading" @click="handleSubmit">
              {{ $t('login.loginButton') }}
            </NButton>
          </NFormItem>

          <!-- <div class="flex justify-end">
          <NButton v-if="isShowRegister" quaternary type="info" class="flex" @click="$router.push({ path: '/register' })">
            注册
          </NButton>
          <NButton quaternary type="info" class="flex" @click="$router.push({ path: '/resetPassword' })">
            忘记密码?
          </NButton>
        </div> -->

          <div class="flex justify-center text-slate-300">
            Powered By <a href="https://github.com/hslr-s/sun-panel" target="_blank" class="ml-[5px] text-slate-500">Sun-Panel</a>
          </div>
        </NForm>
      </NCard>
    </main>
    <UsagiFooter />
  </div>
</template>

<style scoped>
.login-container { padding: 32px 16px; display: flex; justify-content: center; align-items: center; }
.login-card { width: min(100%, 440px); border: 1px solid var(--usagi-brown-border); box-shadow: var(--usagi-shadow-card); }
.login-title { text-align: center; margin: 12px 0 24px; }
.login-title h1 { font-size: 28px; font-weight: 600; color: var(--usagi-text-main); }
</style>
