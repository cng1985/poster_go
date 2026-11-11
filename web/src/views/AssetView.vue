<script setup>
import { onMounted, ref } from 'vue'
import { assetURL, post, upload } from '../api/http'

const items = ref([])
const error = ref('')
const pending = ref(false)

async function load() {
  const page = await post('/api/asset/search', { page: 1, page_size: 40 })
  const withURL = []
  for (const item of page.data) {
    try {
      item.url = await assetURL(item.id)
    } catch {
      item.url = ''
    }
    withURL.push(item)
  }
  items.value = withURL
}

async function onFile(event) {
  const file = event.target.files?.[0]
  event.target.value = ''
  if (!file) return
  error.value = ''
  pending.value = true
  try {
    await upload(file)
    await load()
  } catch (err) {
    error.value = err.message
  } finally {
    pending.value = false
  }
}

async function remove(item) {
  error.value = ''
  try {
    await post('/api/asset/delete', { id: String(item.id) })
    if (item.url) URL.revokeObjectURL(item.url)
    await load()
  } catch (err) {
    error.value = err.message
  }
}

onMounted(load)
</script>

<template>
  <div class="page-head">
    <div>
      <h1>素材库</h1>
      <p class="lede">支持 JPEG、PNG、GIF、WEBP，单张不超过 8MB。图片只属于你自己。</p>
    </div>
  </div>
  <label class="drop">
    <input type="file" accept="image/png,image/jpeg,image/gif,image/webp" hidden @change="onFile" />
    {{ pending ? '正在上传…' : '点击上传图片' }}
  </label>
  <p class="form-error">{{ error }}</p>
  <div class="grid">
    <article v-for="item in items" :key="item.id" class="card">
      <img v-if="item.url" class="thumb" :src="item.url" :alt="item.name" />
      <div class="meta">
        <h3>{{ item.name }}</h3>
        <p>{{ item.width }} × {{ item.height }}</p>
        <button class="btn ghost" @click="remove(item)">删除</button>
      </div>
    </article>
  </div>
</template>
