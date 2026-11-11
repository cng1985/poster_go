<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { assetURL, post } from '../api/http'
import { exportPNG } from '../studio/export'
import { imageLayer, shapeLayer, textLayer } from '../studio/document'

const route = useRoute()
const poster = ref(null)
const doc = ref(null)
const title = ref('')
const selected = ref('')
const status = ref('加载中')
const error = ref('')
const report = ref(null)
const assets = ref([])
const showAssets = ref(false)
const stageRef = ref(null)
const scale = ref(0.35)
const images = reactive({})
const history = []
let drag = null

const layer = computed(() => doc.value?.layers.find((item) => item.id === selected.value) || null)

function snapshot() {
  history.push(JSON.stringify({ title: title.value, doc: doc.value }))
  if (history.length > 40) history.shift()
  status.value = '未保存'
}

function undo() {
  const raw = history.pop()
  if (!raw || !doc.value) return
  const data = JSON.parse(raw)
  title.value = data.title
  doc.value = data.doc
  status.value = '未保存'
}

async function load() {
  error.value = ''
  try {
    const result = await post('/api/poster/one', { id: String(route.params.id) })
    poster.value = result.data
    doc.value = result.data.document
    if (!doc.value.layers) doc.value.layers = []
    title.value = result.data.title
    status.value = '已保存'
    report.value = null
    await hydrate()
    await nextTick()
    measure()
  } catch (err) {
    poster.value = null
    doc.value = null
    status.value = '无法打开'
    error.value = err.message || '这张海报无法打开'
  }
}

async function hydrate() {
  for (const item of doc.value.layers) {
    if (item.type !== 'image' || !item.asset_id) continue
    const id = String(item.asset_id)
    if (images[id]) continue
    const url = await assetURL(id)
    images[id] = url
    const element = new Image()
    element.src = url
    await element.decode().catch(() => {})
    images['el:' + id] = element
  }
}

function measure() {
  if (!stageRef.value || !doc.value) return
  const rect = stageRef.value.getBoundingClientRect()
  scale.value = Math.max(0.12, Math.min((rect.width - 48) / doc.value.width, (rect.height - 48) / doc.value.height, 1))
}

async function save() {
  if (!poster.value) return
  status.value = '保存中'
  error.value = ''
  try {
    const result = await post('/api/poster/update', {
      id: String(poster.value.id),
      title: title.value,
      version: poster.value.version,
      document: doc.value,
    })
    poster.value = result.data
    doc.value = result.data.document
    status.value = '已保存'
  } catch (err) {
    status.value = '保存失败'
    error.value = err.message
    if (err.code === 10004 && confirm('这张海报刚刚被更新过。重新加载会丢掉这次未保存的修改。')) {
      await load()
    }
  }
}

function addText() {
  snapshot()
  const item = textLayer('新的一行', 80, 80, Math.min(640, doc.value.width - 160), 140, 56, '#f4efe6')
  doc.value.layers.push(item)
  selected.value = item.id
}

function addShape(type) {
  snapshot()
  const item = shapeLayer(type)
  doc.value.layers.push(item)
  selected.value = item.id
}

async function toggleAssets() {
  showAssets.value = !showAssets.value
  if (showAssets.value && assets.value.length === 0) {
    const page = await post('/api/asset/search', { page: 1, page_size: 40 })
    assets.value = page.data
  }
}

async function useAsset(asset) {
  snapshot()
  const item = imageLayer(asset)
  doc.value.layers.push(item)
  selected.value = item.id
  showAssets.value = false
  await hydrate()
}

function removeSelected() {
  if (!selected.value) return
  snapshot()
  doc.value.layers = doc.value.layers.filter((item) => item.id !== selected.value)
  selected.value = ''
}

function shift(step) {
  const index = doc.value.layers.findIndex((item) => item.id === selected.value)
  const next = index + step
  if (index < 0 || next < 0 || next >= doc.value.layers.length) return
  snapshot()
  const [item] = doc.value.layers.splice(index, 1)
  doc.value.layers.splice(next, 0, item)
}

function pointerDown(event, item, mode) {
  event.stopPropagation()
  selected.value = item.id
  snapshot()
  drag = { mode, id: item.id, sx: event.clientX, sy: event.clientY, x: item.x, y: item.y, w: item.w, h: item.h }
  window.addEventListener('pointermove', pointerMove)
  window.addEventListener('pointerup', pointerUp)
}

function pointerMove(event) {
  if (!drag || !doc.value) return
  const item = doc.value.layers.find((layer) => layer.id === drag.id)
  if (!item) return
  const dx = (event.clientX - drag.sx) / scale.value
  const dy = (event.clientY - drag.sy) / scale.value
  if (drag.mode === 'move') {
    item.x = Math.round(drag.x + dx)
    item.y = Math.round(drag.y + dy)
  } else {
    item.w = Math.max(24, Math.round(drag.w + dx))
    item.h = Math.max(24, Math.round(drag.h + dy))
  }
}

function pointerUp() {
  drag = null
  window.removeEventListener('pointermove', pointerMove)
  window.removeEventListener('pointerup', pointerUp)
}

function onKey(event) {
  const typing = event.target.matches('input, textarea, select')
  if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 's') {
    event.preventDefault()
    save()
    return
  }
  if (typing) return
  if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'z') {
    event.preventDefault()
    undo()
  }
  if ((event.key === 'Delete' || event.key === 'Backspace') && selected.value) removeSelected()
  if (!layer.value) return
  const step = event.shiftKey ? 10 : 1
  if (event.key.startsWith('Arrow')) {
    event.preventDefault()
    snapshot()
    if (event.key === 'ArrowLeft') layer.value.x -= step
    if (event.key === 'ArrowRight') layer.value.x += step
    if (event.key === 'ArrowUp') layer.value.y -= step
    if (event.key === 'ArrowDown') layer.value.y += step
  }
}

async function analyze() {
  await save()
  if (status.value !== '已保存') return
  report.value = (await post('/api/analysis/create', { poster_id: String(poster.value.id) })).data
}

async function download() {
  const drawn = {}
  for (const item of doc.value.layers) {
    if (item.type === 'image' && images['el:' + item.asset_id]) drawn[String(item.asset_id)] = images['el:' + item.asset_id]
  }
  await exportPNG(doc.value, drawn, title.value)
}

function providerLabel(value) {
  if (value === 'rules') return '规则'
  if (value === 'local') return '本地排版'
  return value || ''
}

function boxStyle(item, index) {
  return {
    left: item.x * scale.value + 'px',
    top: item.y * scale.value + 'px',
    width: item.w * scale.value + 'px',
    height: item.h * scale.value + 'px',
    opacity: item.opacity ?? 1,
    transform: `rotate(${item.rotation || 0}deg)`,
    zIndex: index + 1,
  }
}

function paintStyle(item) {
  const size = (item.font_size || 32) * scale.value
  return {
    background: item.type === 'rect' || item.type === 'ellipse' ? item.fill : 'transparent',
    borderRadius: item.type === 'ellipse' ? '50%' : (item.radius || 0) * scale.value + 'px',
    color: item.color,
    fontSize: size + 'px',
    fontWeight: item.font_weight || 600,
    fontFamily: item.font_family === 'sans' ? 'sans-serif' : 'serif',
    textAlign: item.align || 'left',
    lineHeight: 1.25,
    backgroundImage: item.type === 'image' && images[String(item.asset_id)] ? `url("${images[String(item.asset_id)]}")` : 'none',
    backgroundSize: 'cover',
    backgroundPosition: 'center',
  }
}

onMounted(async () => {
  await load()
  window.addEventListener('keydown', onKey)
  window.addEventListener('resize', measure)
})
onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey)
  window.removeEventListener('resize', measure)
  pointerUp()
})
watch(() => route.params.id, () => load())
</script>

<template>
  <div v-if="!doc" class="editor">
    <header class="editor-bar">
      <router-link class="btn ghost" to="/posters">返回</router-link>
      <span class="status">{{ error || '加载中' }}</span>
    </header>
  </div>
  <div class="editor" v-else>
    <header class="editor-bar">
      <router-link class="btn ghost" to="/posters">返回</router-link>
      <input class="title" v-model="title" @change="status = '未保存'" />
      <span class="status">{{ status }}</span>
      <button class="btn ghost" @click="undo">撤销</button>
      <button class="btn ghost" @click="analyze">分析</button>
      <button class="btn ghost" @click="download">导出 PNG</button>
      <button class="btn primary" @click="save">保存</button>
    </header>
    <aside class="tools">
      <button class="tool" @click="addText">字</button>
      <button class="tool" @click="addShape('rect')">矩</button>
      <button class="tool" @click="addShape('ellipse')">圆</button>
      <button class="tool" :class="{ on: showAssets }" @click="toggleAssets">图</button>
    </aside>
    <div class="stage-wrap" ref="stageRef">
      <div
        class="sheet"
        :style="{ width: doc.width * scale + 'px', height: doc.height * scale + 'px', background: doc.background }"
        @pointerdown="selected = ''"
      >
        <div
          v-for="(item, index) in doc.layers"
          :key="item.id"
          class="layer"
          :class="{ selected: selected === item.id }"
          :style="boxStyle(item, index)"
          @pointerdown="pointerDown($event, item, 'move')"
        >
          <div class="body" :style="paintStyle(item)">{{ item.type === 'text' ? item.text : '' }}</div>
          <i v-if="selected === item.id" class="handle" data-handle @pointerdown="pointerDown($event, item, 'resize')"></i>
        </div>
      </div>
    </div>
    <aside class="inspector">
      <p v-if="error" class="form-error">{{ error }}</p>
      <div v-if="showAssets">
        <h3>放入图片</h3>
        <p v-if="assets.length === 0" class="empty">素材库还是空的，先去上传。</p>
        <button v-for="asset in assets" :key="asset.id" class="layer-row" @click="useAsset(asset)">{{ asset.name }}</button>
      </div>
      <template v-else-if="layer">
        <label>
          <span>名称</span>
          <input v-model="layer.name" @focus="snapshot" />
        </label>
        <template v-if="layer.type === 'text'">
          <label>
            <span>文字</span>
            <textarea v-model="layer.text" @focus="snapshot" />
          </label>
          <label>
            <span>字号 {{ Math.round(layer.font_size) }}</span>
            <input type="range" min="12" max="200" :value="layer.font_size" @focus="snapshot" @input="layer.font_size = Number($event.target.value)" />
          </label>
          <label>
            <span>颜色</span>
            <input type="color" v-model="layer.color" @focus="snapshot" />
          </label>
          <label>
            <span>对齐</span>
            <select v-model="layer.align" @focus="snapshot">
              <option value="left">左</option>
              <option value="center">中</option>
              <option value="right">右</option>
            </select>
          </label>
          <label>
            <span>字体</span>
            <select v-model="layer.font_family" @focus="snapshot">
              <option value="serif">衬线</option>
              <option value="sans">无衬线</option>
            </select>
          </label>
        </template>
        <template v-else-if="layer.type !== 'image'">
          <label>
            <span>填充</span>
            <input type="color" v-model="layer.fill" @focus="snapshot" />
          </label>
          <label v-if="layer.type === 'rect'">
            <span>圆角</span>
            <input type="range" min="0" max="200" :value="layer.radius || 0" @focus="snapshot" @input="layer.radius = Number($event.target.value)" />
          </label>
        </template>
        <label>
          <span>透明 {{ layer.opacity }}</span>
          <input type="range" min="0" max="1" step="0.05" :value="layer.opacity" @focus="snapshot" @input="layer.opacity = Number($event.target.value)" />
        </label>
        <div class="actions">
          <button class="btn ghost" @click="shift(-1)">下移</button>
          <button class="btn ghost" @click="shift(1)">上移</button>
          <button class="btn ghost" @click="removeSelected">删除</button>
        </div>
      </template>
      <template v-else>
        <label>
          <span>背景</span>
          <input type="color" v-model="doc.background" @focus="snapshot" />
        </label>
        <p class="empty">点选图层后改字号、颜色和位置。方向键微调，Delete 删除。</p>
      </template>
      <div v-if="report">
        <div class="score">{{ report.score }}</div>
        <p>{{ report.summary }}</p>
        <p class="empty">{{ providerLabel(report.provider) }}</p>
        <div v-for="(item, index) in report.suggestions" :key="index" class="suggest" :class="item.level">{{ item.text }}</div>
      </div>
      <h3>图层</h3>
      <button
        v-for="item in [...doc.layers].reverse()"
        :key="item.id"
        class="layer-row"
        :class="{ on: selected === item.id }"
        @click="selected = item.id"
      >
        <span>{{ item.name || item.type }}</span>
        <span>{{ item.type }}</span>
      </button>
    </aside>
  </div>
</template>
