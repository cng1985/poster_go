import { FONT_STACKS } from './fonts.js';

const cache = new Map();

function loadImage(src) {
  if (!src) return Promise.reject(new Error('empty'));
  if (cache.has(src)) return cache.get(src);
  const pending = new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve(img);
    img.onerror = () => {
      cache.delete(src);
      reject(new Error(src));
    };
    img.src = src;
  });
  cache.set(src, pending);
  return pending;
}

function coverRect(sw, sh, dw, dh, fit) {
  const scale = fit === 'contain' ? Math.min(dw / sw, dh / sh) : Math.max(dw / sw, dh / sh);
  const w = sw * scale;
  const h = sh * scale;
  return { x: (dw - w) / 2, y: (dh - h) / 2, w, h };
}

function roundPath(ctx, w, h, radius) {
  const r = Math.max(0, Math.min(radius || 0, w / 2, h / 2));
  ctx.beginPath();
  ctx.roundRect(0, 0, w, h, r);
}

function wrapText(ctx, text, maxWidth) {
  const lines = [];
  for (const para of String(text || '').split('\n')) {
    if (!para) {
      lines.push('');
      continue;
    }
    let line = '';
    for (const ch of Array.from(para)) {
      const next = line + ch;
      if (line && ctx.measureText(next).width > maxWidth) {
        lines.push(line);
        line = ch;
      } else {
        line = next;
      }
    }
    lines.push(line);
  }
  return lines;
}

function drawText(ctx, layer) {
  const pad = 8;
  const size = layer.fontSize || 32;
  const family = FONT_STACKS[layer.fontFamily] || FONT_STACKS.sans;
  ctx.font = `${layer.fontWeight || 400} ${size}px ${family}`;
  if (layer.fill) {
    roundPath(ctx, layer.w, layer.h, layer.radius || 0);
    ctx.fillStyle = layer.fill;
    ctx.fill();
  }
  ctx.fillStyle = layer.color || '#1a1a1a';
  ctx.textBaseline = 'top';
  const align = layer.align || 'left';
  ctx.textAlign = align === 'center' ? 'center' : align === 'right' ? 'right' : 'left';
  const maxWidth = Math.max(4, layer.w - pad * 2);
  const lines = wrapText(ctx, layer.text, maxWidth);
  const lineHeight = size * (layer.lineHeight || 1.25);
  const block = lines.length * lineHeight;
  let y = Math.max(pad, (layer.h - block) / 2);
  let x = pad;
  if (align === 'center') x = layer.w / 2;
  if (align === 'right') x = layer.w - pad;
  ctx.save();
  ctx.beginPath();
  ctx.rect(0, 0, layer.w, layer.h);
  ctx.clip();
  for (const line of lines) {
    ctx.fillText(line, x, y);
    y += lineHeight;
  }
  ctx.restore();
}

function paintShape(ctx, layer, path) {
  path();
  if (layer.fill) {
    ctx.fillStyle = layer.fill;
    ctx.fill();
  }
  if (layer.stroke && layer.strokeWidth) {
    ctx.save();
    path();
    ctx.clip();
    ctx.lineWidth = layer.strokeWidth * 2;
    ctx.strokeStyle = layer.stroke;
    path();
    ctx.stroke();
    ctx.restore();
  }
}

async function drawLayer(ctx, layer) {
  ctx.save();
  ctx.globalAlpha = layer.opacity ?? 1;
  ctx.translate(layer.x + layer.w / 2, layer.y + layer.h / 2);
  ctx.rotate(((layer.rotation || 0) * Math.PI) / 180);
  ctx.translate(-layer.w / 2, -layer.h / 2);
  if (layer.type === 'text') {
    drawText(ctx, layer);
  } else if (layer.type === 'image' && layer.src) {
    try {
      const img = await loadImage(layer.src);
      ctx.save();
      roundPath(ctx, layer.w, layer.h, layer.radius || 0);
      ctx.clip();
      const box = coverRect(img.naturalWidth || img.width, img.naturalHeight || img.height, layer.w, layer.h, layer.fit);
      ctx.drawImage(img, box.x, box.y, box.w, box.h);
      ctx.restore();
    } catch {
      // Skip images that fail to decode.
    }
  } else if (layer.type === 'ellipse') {
    paintShape(ctx, layer, () => {
      ctx.beginPath();
      ctx.ellipse(layer.w / 2, layer.h / 2, layer.w / 2, layer.h / 2, 0, 0, Math.PI * 2);
    });
  } else if (layer.type === 'rect' || layer.type === 'line') {
    paintShape(ctx, layer, () => roundPath(ctx, layer.w, layer.h, layer.type === 'line' ? 999 : layer.radius || 0));
  }
  ctx.restore();
}

export async function exportPNG(doc) {
  await document.fonts.ready;
  const scale = Math.min(2, 4096 / Math.max(doc.width, doc.height));
  const canvas = document.createElement('canvas');
  canvas.width = Math.max(1, Math.round(doc.width * scale));
  canvas.height = Math.max(1, Math.round(doc.height * scale));
  const ctx = canvas.getContext('2d');
  ctx.scale(scale, scale);
  ctx.fillStyle = doc.background || '#ffffff';
  ctx.fillRect(0, 0, doc.width, doc.height);
  if (doc.backgroundImage) {
    try {
      const img = await loadImage(doc.backgroundImage);
      const box = coverRect(img.naturalWidth || img.width, img.naturalHeight || img.height, doc.width, doc.height, 'cover');
      ctx.drawImage(img, box.x, box.y, box.w, box.h);
    } catch {
      // Keep the flat background when the image is missing.
    }
  }
  for (const layer of doc.layers) {
    if (layer.visible === false) continue;
    await drawLayer(ctx, layer);
  }
  const blob = await new Promise((resolve) => canvas.toBlob(resolve, 'image/png'));
  if (!blob) throw new Error('导出失败');
  const name = (doc.name || 'poster').replace(/[\\/:*?"<>|]/g, '').trim() || 'poster';
  const link = document.createElement('a');
  const href = URL.createObjectURL(blob);
  link.href = href;
  link.download = name + '.png';
  document.body.append(link);
  link.click();
  link.remove();
  setTimeout(() => URL.revokeObjectURL(href), 2000);
}
