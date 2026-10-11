<script setup lang="ts">
import SvgIcon from '@/components/common/SvgIcon/index.vue'

const links = [
  { title: '导航', href: 'https://nav.usaginest.com', icon: 'lucide:compass', active: true },
  { title: '博客', href: 'https://blog.usaginest.com', icon: 'lucide:pen' },
  { title: '网盘', href: 'https://files.usaginest.com', icon: 'lucide:folder-open' },
  { title: '论坛', href: 'https://bbs.usaginest.com', icon: 'lucide:messages' },
  { title: '管理', href: '#/login', icon: 'lucide:gear' },
]
</script>

<template>
  <footer class="usagi-footer">
    <nav class="usagi-footer-main" aria-label="乌萨奇站点">
      <a
        v-for="link in links" :key="link.title" class="nav-link" :class="{ 'is-active': link.active }"
        :href="link.href" :aria-current="link.active ? 'page' : undefined"
        :target="link.active || link.title === '管理' ? undefined : '_blank'"
        rel="noopener noreferrer"
      >
        <SvgIcon :icon="link.icon" aria-hidden="true" />
        {{ link.title }}
      </a>
    </nav>
    <div class="usagi-footer-tagline">
      © 2024–2026 Usagi Nest · 保留所有权利
    </div>
  </footer>
</template>
