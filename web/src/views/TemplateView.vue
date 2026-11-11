<script setup>
import { onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { post } from '../api/http'
import PosterPreview from '../components/PosterPreview.vue'

const router = useRouter()
const categories = ['全部', '饮品', '促销', '招聘', '活动', '节日', '发布', '餐饮', '教育']
const category = ref('全部')
const items = ref([])

async function load() {
  const body = { page: 1, page_size: 40 }
  if (category.value !== '全部') body.category = category.value
  const page = await post('/api/template/search', body)
  const details = await Promise.all(page.data.map((item) => post('/api/template/one', { id: String(item.id) })))
  items.value = details.map((item) => item.data)
}

async function useTemplate(item) {
  const created = await post('/api/poster/create', { title: item.name, document: item.document })
  router.push('/editor/' + created.data.id)
}

onMounted(load)
watch(category, load)
</script>

<template>
  <div class="page-head">
    <div>
      <h1>海报模块</h1>
      <p class="lede">这些是系统准备的版式。选中后会复制成你自己的海报，改动不会写回模块。</p>
    </div>
  </div>
  <div class="chips">
    <button v-for="item in categories" :key="item" class="chip" :class="{ on: category === item }" @click="category = item">{{ item }}</button>
  </div>
  <div class="grid">
    <button v-for="item in items" :key="item.id" class="card" @click="useTemplate(item)">
      <PosterPreview :document="item.document" :max-width="240" />
      <div class="meta">
        <h3>{{ item.name }}</h3>
        <p>{{ item.category }} · {{ item.summary }}</p>
      </div>
    </button>
  </div>
</template>
