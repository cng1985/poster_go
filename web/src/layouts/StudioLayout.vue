<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSession } from '../stores/session'

const session = useSession()
const router = useRouter()

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
      <div class="mark"><i></i> 海报工坊</div>
      <router-link to="/" active-class="" exact-active-class="router-link-active">工作台</router-link>
      <router-link to="/templates">海报模块</router-link>
      <router-link to="/posters">我的海报</router-link>
      <router-link to="/assets">素材库</router-link>
      <div class="spacer"></div>
      <div class="who">{{ session.account?.display_name || session.account?.username }}</div>
      <button class="btn ghost" @click="logout">退出</button>
    </aside>
    <main class="main">
      <router-view />
    </main>
  </div>
</template>
