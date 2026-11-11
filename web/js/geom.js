// Layers are stored back-to-front. The panel shows them front-to-back.
export function reorderLayers(layers, fromView, toView) {
  if (fromView === toView || fromView < 0 || toView < 0 || fromView >= layers.length || toView >= layers.length) {
    return layers.slice();
  }
  const view = layers.slice().reverse();
  const [item] = view.splice(fromView, 1);
  view.splice(toView, 0, item);
  return view.reverse();
}

const EDGE = {
  n: ['n', 'ne', 'nw'],
  s: ['s', 'se', 'sw'],
  e: ['e', 'ne', 'se'],
  w: ['w', 'nw', 'sw'],
};

function hasEdge(handle, dir) {
  return EDGE[dir].includes(handle);
}

export function localToParent(x, y, w, h, rotationDeg, lx, ly) {
  const rad = (rotationDeg * Math.PI) / 180;
  const c = Math.cos(rad);
  const s = Math.sin(rad);
  const dx = lx - w / 2;
  const dy = ly - h / 2;
  return {
    x: x + w / 2 + dx * c - dy * s,
    y: y + h / 2 + dx * s + dy * c,
  };
}

// Resize in parent space. Rotation is around the layer center, matching CSS.
export function resizeLayer(start, handle, dx, dy) {
  const rot = start.rotation || 0;
  const rad = (-rot * Math.PI) / 180;
  const c = Math.cos(rad);
  const s = Math.sin(rad);
  const lx = dx * c - dy * s;
  const ly = dx * s + dy * c;
  let left = 0;
  let top = 0;
  let right = start.w;
  let bottom = start.h;
  if (hasEdge(handle, 'e')) right = start.w + lx;
  if (hasEdge(handle, 's')) bottom = start.h + ly;
  if (hasEdge(handle, 'w')) left = lx;
  if (hasEdge(handle, 'n')) top = ly;
  const min = 8;
  if (right - left < min) {
    if (hasEdge(handle, 'w')) left = right - min;
    else right = left + min;
  }
  if (bottom - top < min) {
    if (hasEdge(handle, 'n')) top = bottom - min;
    else bottom = top + min;
  }
  const newW = right - left;
  const newH = bottom - top;
  const anchors = {
    e: [start.w, start.h / 2],
    w: [0, start.h / 2],
    n: [start.w / 2, 0],
    s: [start.w / 2, start.h],
    ne: [start.w, 0],
    nw: [0, 0],
    se: [start.w, start.h],
    sw: [0, start.h],
  };
  const anchor = anchors[handle];
  const anchorNew = [anchor[0] - left, anchor[1] - top];
  const p = localToParent(start.x, start.y, start.w, start.h, rot, anchor[0], anchor[1]);
  const rad2 = (rot * Math.PI) / 180;
  const c2 = Math.cos(rad2);
  const s2 = Math.sin(rad2);
  const ox = anchorNew[0] - newW / 2;
  const oy = anchorNew[1] - newH / 2;
  const rx = ox * c2 - oy * s2;
  const ry = ox * s2 + oy * c2;
  return {
    x: p.x - rx - newW / 2,
    y: p.y - ry - newH / 2,
    w: newW,
    h: newH,
    rotation: rot,
  };
}

export function snapMove(doc, x, y, w, h, threshold = 6) {
  const guides = { v: false, h: false };
  if (Math.abs(x) <= threshold) x = 0;
  if (Math.abs(y) <= threshold) y = 0;
  if (Math.abs(x + w - doc.width) <= threshold) x = doc.width - w;
  if (Math.abs(y + h - doc.height) <= threshold) y = doc.height - h;
  if (Math.abs(x + w / 2 - doc.width / 2) <= threshold) {
    x = (doc.width - w) / 2;
    guides.v = true;
  }
  if (Math.abs(y + h / 2 - doc.height / 2) <= threshold) {
    y = (doc.height - h) / 2;
    guides.h = true;
  }
  return { x, y, guides };
}

export function scaleLayers(doc, width, height) {
  width = Math.max(240, Math.min(4096, Math.round(width)));
  height = Math.max(240, Math.min(4096, Math.round(height)));
  const sx = width / doc.width;
  const sy = height / doc.height;
  const fontScale = Math.min(sx, sy);
  for (const layer of doc.layers) {
    layer.x = Math.round(layer.x * sx);
    layer.y = Math.round(layer.y * sy);
    layer.w = Math.max(4, Math.round(layer.w * sx));
    layer.h = Math.max(4, Math.round(layer.h * sy));
    if (layer.fontSize) layer.fontSize = Math.max(8, Math.round(layer.fontSize * fontScale));
    if (layer.strokeWidth) layer.strokeWidth = Math.round(layer.strokeWidth * fontScale * 10) / 10;
    if (layer.radius) layer.radius = Math.round(layer.radius * fontScale);
  }
  doc.width = width;
  doc.height = height;
}
