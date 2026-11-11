export const presets = {
  story: { width: 1080, height: 1920, label: '故事 1080×1920' },
  poster: { width: 1080, height: 1440, label: '海报 1080×1440' },
  square: { width: 1080, height: 1080, label: '方图 1080×1080' },
  cover: { width: 1920, height: 1080, label: '横版 1920×1080' },
}

export const styles = [
  { id: 'editorial', label: '编辑感' },
  { id: 'sale', label: '促销' },
  { id: 'festive', label: '节日' },
  { id: 'minimal', label: '留白' },
  { id: 'bold', label: '强对比' },
]

export function uid() {
  return 'l' + Math.random().toString(36).slice(2, 10)
}

export function clone(value) {
  return JSON.parse(JSON.stringify(value))
}

export function blankDocument(preset = 'poster') {
  const size = presets[preset] || presets.poster
  return {
    width: size.width,
    height: size.height,
    background: '#1c1915',
    layers: [
      textLayer('标题写在这里', 80, size.height * 0.28, size.width - 160, 180, 84, '#f3ecdf'),
      textLayer('补一句短说明', 80, size.height * 0.48, size.width - 160, 100, 32, '#d9cbb8'),
    ],
  }
}

export function textLayer(text, x, y, w, h, fontSize, color) {
  return {
    id: uid(),
    type: 'text',
    name: '文字',
    x, y, w, h,
    rotation: 0,
    opacity: 1,
    text,
    font_size: fontSize,
    font_weight: 700,
    font_family: 'serif',
    color,
    align: 'left',
  }
}

export function shapeLayer(type) {
  return {
    id: uid(),
    type,
    name: type === 'ellipse' ? '圆形' : '矩形',
    x: 120,
    y: 120,
    w: 360,
    h: type === 'ellipse' ? 360 : 180,
    rotation: 0,
    opacity: 1,
    fill: '#e4572e',
    radius: type === 'rect' ? 16 : 0,
  }
}

export function imageLayer(asset) {
  const width = Math.min(640, asset.width || 480)
  const height = Math.max(80, Math.round(width * ((asset.height || 1) / (asset.width || 1))))
  return {
    id: uid(),
    type: 'image',
    name: String(asset.name || '图片').slice(0, 40),
    x: 140,
    y: 180,
    w: width,
    h: height,
    rotation: 0,
    opacity: 1,
    asset_id: String(asset.id),
  }
}

export const samplePoster = {
  width: 1080,
  height: 1440,
  background: '#14342c',
  layers: [
    { id: 'orb', type: 'ellipse', name: 'orb', x: 640, y: -40, w: 460, h: 460, opacity: 1, fill: '#1f6a52' },
    { id: 'band', type: 'rect', name: 'band', x: 0, y: 1120, w: 1080, h: 320, opacity: 1, fill: '#e4572e', radius: 0 },
    { id: 'kicker', type: 'text', name: 'kicker', x: 80, y: 180, w: 700, h: 70, opacity: 1, text: '海报工坊', font_size: 32, font_weight: 600, font_family: 'sans', color: '#d4a656', align: 'left' },
    { id: 'title', type: 'text', name: 'title', x: 80, y: 280, w: 880, h: 280, opacity: 1, text: '把一句话\n排成一张海报', font_size: 92, font_weight: 700, font_family: 'serif', color: '#f3ecdf', align: 'left' },
    { id: 'sub', type: 'text', name: 'sub', x: 80, y: 620, w: 760, h: 120, opacity: 1, text: '模板、上传、生成，然后在编辑器里改到能发出去。', font_size: 32, font_weight: 500, font_family: 'sans', color: '#d7efe4', align: 'left' },
  ],
}
