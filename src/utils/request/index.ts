import { apiRespErrMsg, message } from './apiMessage'
import { t } from '@/locales'
import { useAppStore, useAuthStore } from '@/store'
import { router } from '@/router'

let loginMessageShow = false
export interface HttpOption {
  url: string
  data?: any
  method?: string
  headers?: any
  signal?: AbortSignal
  beforeRequest?: () => void
  afterRequest?: () => void
}

export interface Response<T = any> {
  data: T
  // message: string | null
  // status: string
  msg: string
  code: number
}

const baseURL = import.meta.env.VITE_GLOB_API_URL || ''

// A small fetch client: the panel only exchanges JSON with its own server.
async function request(url: string, method: string, params: Record<string, any>, headers: Record<string, string>, signal?: AbortSignal) {
  let target = baseURL + url
  const init: RequestInit = { method, headers: { ...headers }, signal }
  if (method === 'GET') {
    const query = new URLSearchParams(Object.entries(params).filter(([, v]) => v !== undefined && v !== null).map(([k, v]) => [k, String(v)])).toString()
    if (query)
      target += (target.includes('?') ? '&' : '?') + query
  }
  else {
    (init.headers as Record<string, string>)['Content-Type'] = 'application/json'
    init.body = JSON.stringify(params)
  }
  const response = await fetch(target, init)
  if (response.status !== 200)
    throw new Error(response.status.toString())
  return { data: await response.json() }
}

function http<T = any>(
  { url, data, method, headers, signal, beforeRequest, afterRequest }: HttpOption,
) {
  const authStore = useAuthStore()
  const appStore = useAppStore()
  const successHandler = (res: { data: Response<T> }) => {
    if (res.data.code === 0)
      return res.data

    if (res.data.code === 1001) {
      // 避免重复弹窗
      if (loginMessageShow === false) {
        loginMessageShow = true
        message.warning(t('api.loginExpires'), {
        // message.warning('登录过期', {
          onLeave() {
            loginMessageShow = false
          },
        })
      }

      router.push({ path: '/login' })
      authStore.removeToken()
      return res.data
    }

    if (res.data.code === 1000) {
      router.push({ path: '/login' })
      authStore.removeToken()
      return res.data
    }

    if (res.data.code === 1005) {
      message.warning(res.data.msg)
      return res.data
    }

    if (res.data.code === -1) {
      // message.warning(res.data.msg)
      // router.push({ path: '/login' })
      // authStore.removeToken()
      return res.data
    }

    if (!apiRespErrMsg(res.data))
      return Promise.reject(res.data)
    else
      return res.data
  }

  const failHandler = (error: Response<Error>) => {
    afterRequest?.()
    message.error(t('common.networkError'), {
      duration: 50000,
      closable: true,
    })
    throw new Error(error?.msg || 'Error')
  }

  beforeRequest?.()

  method = method || 'GET'

  const params = Object.assign(typeof data === 'function' ? data() : data ?? {}, {})
  if (!headers)
    headers = {}

  if (authStore.token)
    headers.Authorization = `Bearer ${authStore.token}`
  if (method !== 'GET') {
    if (authStore.token)
      headers.token = authStore.token
    if (appStore.language)
      headers.lang = appStore.language
  }
  return request(url, method, params, headers, signal).then(successHandler, failHandler)
}

// Requests started at boot, before the code of the page that needs them has loaded.
const prefetched = new Map<string, { promise: Promise<Response<any>>; at: number }>()
const requestKey = (url: string, data: any) => `${url}\0${JSON.stringify(data ?? {})}`

export function prefetch(url: string, data?: any) {
  const promise = http({ url, data, method: 'POST' })
  // The page that consumes it handles failures; an unused prefetch stays quiet.
  promise.catch(() => {})
  prefetched.set(requestKey(url, data), { promise, at: Date.now() })
}

export function get<T = any>(
  { url, data, method = 'GET', signal, beforeRequest, afterRequest }: HttpOption,
): Promise<Response<T>> {
  return http<T>({
    url,
    method,
    data,
    signal,
    beforeRequest,
    afterRequest,
  })
}

export function post<T = any>(
  { url, data, method = 'POST', headers, signal, beforeRequest, afterRequest }: HttpOption,
): Promise<Response<T>> {
  const key = requestKey(url, data)
  const early = prefetched.get(key)
  if (early) {
    prefetched.delete(key)
    if (Date.now() - early.at < 15000 && !headers && !signal)
      return early.promise
  }
  return http<T>({
    url,
    method,
    data,
    headers,
    signal,
    beforeRequest,
    afterRequest,
  })
}

export default post
