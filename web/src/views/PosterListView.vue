<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { post } from '../api/http'
import PosterPreview from '../components/PosterPreview.vue'

const router = useRouter()
const items = ref([])
const error = ref('')

async function load() {
  const page = await post('/api/poster/search', { page: 1, page_size: 40, sort: 'updated_at', order: 'desc' })
  const details = await Promise.all(page.data.map((item) => post('/api/poster/one', { id: String(item.id) })))
  items.value = details.map((item) => item.data)
}

async function duplicate(item) {
  error.value = ''
  try {
    const created = await post('/api/poster/duplicate', { id: String(item.id) })
    router.push('/editor/' + created.data.id)
  } catch (err) {
    error.value = err.message
  }
}

async function remove(item) {
  if (!confirm('删除这张海报？此操作不能撤销。')) return
  await post('/api/poster/delete', { id: String(item.id) })
  await load()
}

onMounted(load)
</script>

<template>
  <div class="page-head">
    <div>
      <h1>我的海报</h1>
      <p class="lede">只显示你自己的作品。</p>
    </div>
  </div>
  <p class="form-error">{{ error }}</p>
  <div v-if="items.length === 0" class="empty">这里还是空的。</div>
  <div class="grid">
    <article v-for="item in items" :key="item.id" class="card">
      <router-link :to="'/editor/' + item.id">
        <PosterPreview :document="item.document" :max-width="240" />
      </router-link>
      <div class="meta">
        <h3>{{ item.title }}</h3>
        <p>版本 {{ item.version }}</p>
        <div class="actions">
          <button class="btn ghost" @click="duplicate(item)">复制</button>
          <button class="btn ghost" @click="remove(item)">删除</button>
        </div>
      </div>
    </article>
  </div>
</template>
