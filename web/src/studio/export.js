export async function exportPNG(doc, images, title) {
  const canvas = document.createElement('canvas')
  canvas.width = doc.width
  canvas.height = doc.height
  const ctx = canvas.getContext('2d')
  ctx.fillStyle = doc.background || '#111111'
  ctx.fillRect(0, 0, doc.width, doc.height)
  for (const layer of doc.layers) {
    ctx.save()
    ctx.globalAlpha = layer.opacity ?? 1
    const cx = layer.x + layer.w / 2
    const cy = layer.y + layer.h / 2
    ctx.translate(cx, cy)
    ctx.rotate(((layer.rotation || 0) * Math.PI) / 180)
    ctx.translate(-layer.w / 2, -layer.h / 2)
    if (layer.type === 'rect') {
      roundRect(ctx, 0, 0, layer.w, layer.h, layer.radius || 0)
      ctx.fillStyle = layer.fill || '#e4572e'
      ctx.fill()
    } else if (layer.type === 'ellipse') {
      ctx.beginPath()
      ctx.ellipse(layer.w / 2, layer.h / 2, layer.w / 2, layer.h / 2, 0, 0, Math.PI * 2)
      ctx.fillStyle = layer.fill || '#e4572e'
      ctx.fill()
    } else if (layer.type === 'image') {
      const image = images[String(layer.asset_id)]
      if (image) ctx.drawImage(image, 0, 0, layer.w, layer.h)
    } else if (layer.type === 'text') {
      drawText(ctx, layer)
    }
    ctx.restore()
  }
  const blob = await new Promise((resolve) => canvas.toBlob(resolve, 'image/png'))
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = (title || '海报') + '.png'
  link.click()
  URL.revokeObjectURL(url)
}

function roundRect(ctx, x, y, w, h, radius) {
  const r = Math.min(radius, w / 2, h / 2)
  ctx.beginPath()
  ctx.moveTo(x + r, y)
  ctx.arcTo(x + w, y, x + w, y + h, r)
  ctx.arcTo(x + w, y + h, x, y + h, r)
  ctx.arcTo(x, y + h, x, y, r)
  ctx.arcTo(x, y, x + w, y, r)
  ctx.closePath()
}

function drawText(ctx, layer) {
  const family = layer.font_family === 'sans' ? 'sans-serif' : 'serif'
  const weight = layer.font_weight || 600
  const size = layer.font_size || 48
  ctx.fillStyle = layer.color || '#ffffff'
  ctx.font = `${weight} ${size}px ${family}`
  ctx.textBaseline = 'top'
  const lines = wrap(ctx, layer.text || '', layer.w)
  const lineHeight = size * 1.25
  lines.forEach((line, index) => {
    let x = 0
    if (layer.align === 'center') x = (layer.w - ctx.measureText(line).width) / 2
    if (layer.align === 'right') x = layer.w - ctx.measureText(line).width
    ctx.fillText(line, x, index * lineHeight)
  })
}

function wrap(ctx, text, maxWidth) {
  const lines = []
  let line = ''
  for (const ch of String(text)) {
    if (ch === '\n') {
      lines.push(line)
      line = ''
      continue
    }
    const next = line + ch
    if (line && ctx.measureText(next).width > maxWidth) {
      lines.push(line)
      line = ch
    } else {
      line = next
    }
  }
  lines.push(line)
  return lines
}
