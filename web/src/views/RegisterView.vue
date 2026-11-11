<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useSession } from '../stores/session'

const router = useRouter()
const session = useSession()
const form = ref({ username: '', password: '', display_name: '' })
const error = ref('')
const pending = ref(false)

async function submit() {
  error.value = ''
  pending.value = true
  try {
    await session.register(form.value)
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
      <div class="gate-copy">
        <h1>开一个自己的版面。</h1>
        <p>用户名以字母开头。密码至少 8 位，并且同时包含字母和数字。</p>
      </div>
    </section>
    <section class="gate-form">
      <form class="panel" @submit.prevent="submit">
        <h2>注册</h2>
        <p class="hint">注册后会直接登录。</p>
        <label>
          <span>用户名</span>
          <input v-model="form.username" autocomplete="username" required />
        </label>
        <label>
          <span>显示名称</span>
          <input v-model="form.display_name" autocomplete="nickname" />
        </label>
        <label>
          <span>密码</span>
          <input v-model="form.password" type="password" autocomplete="new-password" required />
        </label>
        <p class="form-error">{{ error }}</p>
        <button class="btn primary block" :disabled="pending">创建账号</button>
        <p class="switch-link">已有账号？ <router-link to="/login">登录</router-link></p>
      </form>
    </section>
  </div>
</template>
