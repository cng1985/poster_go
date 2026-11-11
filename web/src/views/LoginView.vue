<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import PosterPreview from '../components/PosterPreview.vue'
import { samplePoster } from '../studio/document'
import { useSession } from '../stores/session'

const router = useRouter()
const session = useSession()
const form = ref({ username: '', password: '' })
const error = ref('')
const pending = ref(false)

async function submit() {
  error.value = ''
  pending.value = true
  try {
    await session.login(form.value)
    router.push('/')
  } catch (err) {
    error.value = err.message
  } finally {
    pending.value = false
  }
}
</script>

<template>
  <div class="gate">
    <section class="gate-stage">
      <div class="mark"><i></i> 海报工坊</div>
      <div class="showcase">
        <PosterPreview :document="samplePoster" :max-width="280" />
      </div>
      <div class="gate-copy">
        <p class="kicker">设计工作室</p>
        <h1>先排一版，<br />再交给模型。</h1>
        <p>从现成模块开始，上传自己的图片，或用一句话生成海报。编辑器里可以改字、改色、导出。</p>
      </div>
    </section>
    <section class="gate-form">
      <form class="panel" @submit.prevent="submit">
        <h2>登录</h2>
        <p class="hint">继续你的海报。</p>
        <label>
          <span>用户名</span>
          <input v-model="form.username" autocomplete="username" required />
        </label>
        <label>
          <span>密码</span>
          <input v-model="form.password" type="password" autocomplete="current-password" required />
        </label>
        <p class="form-error">{{ error }}</p>
        <button class="btn primary block" :disabled="pending">进入工坊</button>
        <p class="switch-link">还没有账号？ <router-link to="/register">注册</router-link></p>
      </form>
    </section>
  </div>
</template>
