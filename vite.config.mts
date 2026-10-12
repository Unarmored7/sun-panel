import path from 'path'
import { readFileSync, writeFileSync } from 'node:fs'
import { brotliCompressSync, constants, gzipSync } from 'node:zlib'
import type { PluginOption } from 'vite'
import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'

function setupPlugins(env: ImportMetaEnv): PluginOption[] {
  return [
    vue(),
    {
      name: 'precompress-static-assets',
      writeBundle(options, bundle) {
        for (const name of Object.keys(bundle)) {
          if (!/\.(?:js|css|svg)$/.test(name))
            continue
          const file = path.join(options.dir || 'dist', name)
          const bytes = readFileSync(file)
          if (bytes.length < 1024)
            continue
          writeFileSync(`${file}.gz`, gzipSync(bytes, { level: 9 }))
          writeFileSync(`${file}.br`, brotliCompressSync(bytes, { params: { [constants.BROTLI_PARAM_QUALITY]: 9 } }))
        }
      },
    },
    {
      // The home page is the landing route: preload its chunks with the entry
      // instead of waiting for the router to request them.
      name: 'preload-home-route',
      apply: 'build',
      transformIndexHtml: {
        order: 'post',
        handler(html, ctx) {
          const bundle = ctx.bundle
          if (!bundle)
            return html
          const home = Object.values(bundle).find(chunk => chunk.type === 'chunk' && chunk.facadeModuleId?.endsWith('/src/views/home/index.vue'))
          if (!home || home.type !== 'chunk')
            return html
          const files = new Set<string>()
          const css = new Set<string>()
          const visit = (name: string) => {
            const chunk = bundle[name]
            if (!chunk || chunk.type !== 'chunk' || files.has(name) || chunk.isEntry)
              return
            files.add(name)
            chunk.viteMetadata?.importedCss.forEach(file => css.add(file))
            chunk.imports.forEach(visit)
          }
          visit(home.fileName)
          const tags = [...files].filter(file => !html.includes(file)).map(file => `<link rel="modulepreload" crossorigin href="/${file}">`)
            .concat([...css].filter(file => !html.includes(file)).map(file => `<link rel="stylesheet" crossorigin href="/${file}">`))
          return html.replace('</head>', `  ${tags.join('\n  ')}\n</head>`)
        },
      },
    },
    env.VITE_GLOB_APP_PWA === 'true' && VitePWA({
      injectRegister: 'auto',
      manifest: {
        name: 'Sun-Panel',
        short_name: 'Sun-Panel',
        icons: [
          { src: 'pwa-192x192.png', sizes: '192x192', type: 'image/png' },
          { src: 'pwa-512x512.png', sizes: '512x512', type: 'image/png' },
        ],
      },
    }),
  ]
}

export default defineConfig((env) => {
  const viteEnv = loadEnv(env.mode, process.cwd()) as unknown as ImportMetaEnv

  return {
    resolve: {
      alias: {
        '@': path.resolve(process.cwd(), 'src'),
      },
    },
    plugins: setupPlugins(viteEnv),
    server: {
      host: '0.0.0.0',
      port: 1002,
      open: false,
      proxy: {
        '/api': {
          target: viteEnv.VITE_APP_API_BASE_URL,
          changeOrigin: true, // 允许跨域
          rewrite: path => path.replace('/api/', '/api/'),
        },
        '/icon-cache': {
          target: viteEnv.VITE_APP_API_BASE_URL,
          changeOrigin: true,
        },
        '/uploads': {
          target: viteEnv.VITE_APP_API_BASE_URL,
          changeOrigin: true, // 允许跨域
          rewrite: path => path.replace('/uploads/', '/uploads/'),
        },
      },
    },
    build: {
      outDir: process.env.SUN_PANEL_WEB_OUT_DIR || 'dist',
      emptyOutDir: true,
      reportCompressedSize: false,
      sourcemap: false,
      commonjsOptions: {
        ignoreTryCatch: false,
      },

    },
  }
})
