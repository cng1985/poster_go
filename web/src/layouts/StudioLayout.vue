<script setup>
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSession } from '../stores/session'

const session = useSession()
const router = useRouter()
const name = computed(() => session.account?.display_name || session.account?.username || '访客')
const initial = computed(() => name.value.slice(0, 1))

onMounted(() => {
  session.refresh().catch(() => {})
})

async function logout() {
  await session.logout()
  router.push('/login')
}
</script>

<template>
  <div class="shell">
    <aside class="nav">
      <div class="brand">
        <span class="logo" aria-hidden="true"></span>
        <div>
          <strong>海报工坊</strong>
          <small>Studio</small>
        </div>
      </div>
      <router-link to="/" active-class="" exact-active-class="router-link-active">
        <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 10.5 12 4l8 6.5V20a1 1 0 0 1-1 1h-5v-6H10v6H5a1 1 0 0 1-1-1z" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round"/></svg>
        工作台
      </router-link>
      <router-link to="/templates">
        <svg viewBox="0 0 24 24" aria-hidden="true"><rect x="4" y="4" width="7" height="7" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.6"/><rect x="13" y="4" width="7" height="7" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.6"/><rect x="4" y="13" width="7" height="7" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.6"/><rect x="13" y="13" width="7" height="7" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.6"/></svg>
        海报模块
      </router-link>
      <router-link to="/posters">
        <svg viewBox="0 0 24 24" aria-hidden="true"><rect x="5" y="3.5" width="14" height="17" rx="2" fill="none" stroke="currentColor" stroke-width="1.6"/><path d="M8 8h8M8 12h8M8 16h5" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/></svg>
        我的海报
      </router-link>
      <router-link to="/assets">
        <svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3.5" y="5" width="17" height="14" rx="2" fill="none" stroke="currentColor" stroke-width="1.6"/><circle cx="9" cy="10" r="1.4" fill="currentColor"/><path d="m7 16 3.2-3.2a1 1 0 0 1 1.4 0L17 17" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/></svg>
        素材库
      </router-link>
      <div class="spacer"></div>
      <div class="who">
        <span class="avatar">{{ initial }}</span>
        <span>{{ name }}</span>
      </div>
      <button class="btn ghost block" @click="logout">退出登录</button>
    </aside>
    <main class="main">
      <router-view />
    </main>
  </div>
</template>
