<script setup>
import { computed } from 'vue'

const props = defineProps({
  document: { type: Object, required: true },
  maxWidth: { type: Number, default: 280 },
})

const scale = computed(() => props.maxWidth / (props.document.width || 1))
const box = computed(() => ({
  width: props.maxWidth + 'px',
  height: (props.document.height || 1) * scale.value + 'px',
}))
const sheet = computed(() => ({
  width: props.document.width + 'px',
  height: props.document.height + 'px',
  background: props.document.background,
  transform: `scale(${scale.value})`,
  transformOrigin: 'top left',
}))

function styleOf(layer) {
  return {
    left: layer.x + 'px',
    top: layer.y + 'px',
    width: layer.w + 'px',
    height: layer.h + 'px',
    opacity: layer.opacity ?? 1,
    transform: `rotate(${layer.rotation || 0}deg)`,
    background: layer.type === 'image' ? 'rgba(255,255,255,.18)' : layer.fill || 'transparent',
    borderRadius: layer.type === 'ellipse' ? '50%' : (layer.radius || 0) + 'px',
    color: layer.color,
    fontSize: (layer.font_size || 32) + 'px',
    fontWeight: layer.font_weight || 600,
    fontFamily: layer.font_family === 'sans' ? 'sans-serif' : 'serif',
    textAlign: layer.align || 'left',
    lineHeight: 1.25,
    whiteSpace: 'pre-wrap',
  }
}
</script>

<template>
  <div class="preview" :style="box">
    <div class="sheet" :style="sheet">
      <div v-for="layer in document.layers" :key="layer.id" class="layer" :style="styleOf(layer)">
        <div v-if="layer.type === 'text'" class="body">{{ layer.text }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.preview { position: relative; overflow: hidden; border-radius: 14px; }
.sheet { position: absolute; left: 0; top: 0; }
.layer { position: absolute; overflow: hidden; }
.body { width: 100%; height: 100%; }
</style>
