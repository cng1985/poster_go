<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { post } from '../api/http'
import PosterPreview from '../components/PosterPreview.vue'
import { blankDocument, presets, styles } from '../studio/document'

const router = useRouter()
const recent = ref([])
const open = ref(false)
const pending = ref(false)
const error = ref('')
const form = ref({ prompt: '春日咖啡新品，暖色，留白多一点', style: 'editorial', size_preset: 'poster', title: '' })

onMounted(async () => {
  const page = await post('/api/poster/search', { page: 1, page_size: 4, sort: 'updated_at', order: 'desc' })
  const details = await Promise.all(page.data.map((item) => post('/api/poster/one', { id: String(item.id) })))
  recent.value = details.map((item) => item.data)
})

async function createBlank() {
  const created = await post('/api/poster/create', { title: '未命名海报', document: blankDocument('poster') })
  router.push('/editor/' + created.data.id)
}

async function generate() {
  error.value = ''
  pending.value = true
  try {
    const result = await post('/api/generation/create', { ...form.value, save_poster: true })
    if (!result.data.poster_id) throw new Error(result.data.message || '没有生成作品')
    router.push('/editor/' + result.data.poster_id)
  } catch (err) {
    error.value = err.message
  } finally {
    pending.value = false
  }
}
</script>

<template>
  <section class="hero" :class="{ solo: !recent[0] }">
    <div>
      <p class="kicker">工作台</p>
      <h1>今天做哪一张？</h1>
      <p class="lede">选一个模块，上传图片，或让模型先排一版。没有配置模型密钥时，工坊会用本地排版接上，不会空着手。</p>
      <div class="actions">
        <button class="btn primary" @click="open = true">用一句话生成</button>
        <button class="btn ghost" @click="createBlank">空白海报</button>
        <router-link class="btn ghost" to="/templates">看海报模块</router-link>
      </div>
    </div>
    <div class="hero-card" v-if="recent[0]">
      <p class="kicker">最近一张</p>
      <PosterPreview :document="recent[0].document" :max-width="320" />
      <p class="hero-caption">{{ recent[0].title }}</p>
    </div>
  </section>
  <h2>最近的海报</h2>
  <div v-if="recent.length === 0" class="empty">还没有海报。可以从模块或一句话开始。</div>
  <div class="grid">
    <router-link v-for="item in recent" :key="item.id" class="card" :to="'/editor/' + item.id">
      <PosterPreview :document="item.document" :max-width="240" />
      <div class="meta">
        <h3>{{ item.title }}</h3>
        <p>{{ item.width }} × {{ item.height }}</p>
      </div>
    </router-link>
  </div>
  <div v-if="open" class="modal-back" @click.self="open = false">
    <form class="modal" @submit.prevent="generate">
      <h2>生成海报</h2>
      <label>
        <span>想要什么样的海报</span>
        <textarea v-model="form.prompt" required />
      </label>
      <div class="row">
        <label>
          <span>风格</span>
          <select v-model="form.style">
            <option v-for="item in styles" :key="item.id" :value="item.id">{{ item.label }}</option>
          </select>
        </label>
        <label>
          <span>尺寸</span>
          <select v-model="form.size_preset">
            <option v-for="(item, key) in presets" :key="key" :value="key">{{ item.label }}</option>
          </select>
        </label>
      </div>
      <label>
        <span>标题，可留空</span>
        <input v-model="form.title" />
      </label>
      <p class="form-error">{{ error }}</p>
      <div class="actions">
        <button class="btn primary" :disabled="pending">生成并打开</button>
        <button class="btn ghost" type="button" @click="open = false">取消</button>
      </div>
    </form>
  </div>
</template>
