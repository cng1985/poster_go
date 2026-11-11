import { FONT_STACKS } from './fonts.js';
import { resizeLayer, snapMove } from './geom.js';
import {
  armHistory,
  disarmHistory,
  normalizeText,
  select,
  selectedLayer,
  setGesturing,
  state,
  touchDirty,
} from './model.js';

let stage;
let frame;
const nodes = new Map();
let dropAsset = async () => {};
let dropFiles = async () => {};

export function setDropHandlers(handlers) {
  dropAsset = handlers.asset || dropAsset;
  dropFiles = handlers.files || dropFiles;
}

export function mount(stageEl) {
  stage = stageEl;
  stage.replaceChildren();
  const inner = document.createElement('div');
  inner.className = 'stage-inner';
  stage.append(inner);
  stage.addEventListener('pointerdown', (event) => {
    if (event.target.closest('.layer') || event.target.closest('.sel-box')) return;
    select(null);
  });
  stage.addEventListener('dragover', (event) => {
    if (![...event.dataTransfer.types].some((type) => type === 'application/x-poster-asset' || type === 'Files')) return;
    event.preventDefault();
    stage.classList.add('is-drop');
  });
  stage.addEventListener('dragleave', () => stage.classList.remove('is-drop'));
  stage.addEventListener('drop', onDrop);
  window.addEventListener('resize', () => sync());
}

export function sync() {
  if (!state.doc || !stage) return;
  const inner = stage.querySelector('.stage-inner');
  if (!frame) {
    frame = document.createElement('div');
    frame.className = 'poster-frame';
    const clip = document.createElement('div');
    clip.className = 'poster-clip';
    const hint = document.createElement('div');
    hint.className = 'empty-hint';
    hint.append(Object.assign(document.createElement('span'), { textContent: '从左侧添加组件，或从图片库选一张图' }));
    clip.append(hint);
    frame.append(clip, guide('guide-v'), guide('guide-h'), selectionBox());
    inner.append(frame);
  }
  const scale = fitScale() * state.zoom;
  const scaler = frame.parentElement === inner ? ensureScaler(inner) : frame.parentElement;
  frame.style.width = state.doc.width + 'px';
  frame.style.height = state.doc.height + 'px';
  frame.style.transform = `scale(${scale})`;
  scaler.style.width = state.doc.width * scale + 'px';
  scaler.style.height = state.doc.height * scale + 'px';
  const clip = frame.querySelector('.poster-clip');
  clip.style.background = state.doc.background || '#ffffff';
  paintBackground(clip);
  const sig = state.doc.layers.map((layer) => layer.id + ':' + layer.type).join('|');
  if (clip.dataset.sig !== sig) {
    buildLayers(clip);
    clip.dataset.sig = sig;
  }
  for (const layer of state.doc.layers) paintLayer(nodes.get(layer.id), layer);
  const hint = clip.querySelector('.empty-hint');
  hint.hidden = state.doc.layers.length > 0;
  paintSelection();
  paintGuides();
}

function ensureScaler(inner) {
  let scaler = inner.querySelector('.scaler');
  if (!scaler) {
    scaler = document.createElement('div');
    scaler.className = 'scaler';
    inner.append(scaler);
    scaler.append(frame);
  }
  return scaler;
}

function guide(className) {
  const node = document.createElement('div');
  node.className = 'guide ' + className;
  return node;
}

function selectionBox() {
  const box = document.createElement('div');
  box.className = 'sel-box';
  box.hidden = true;
  return box;
}

function fitScale() {
  const pad = 96;
  const availW = Math.max(120, stage.clientWidth - pad);
  const availH = Math.max(120, stage.clientHeight - pad);
  return Math.min(availW / state.doc.width, availH / state.doc.height);
}

function buildLayers(clip) {
  nodes.clear();
  clip.querySelectorAll('.layer').forEach((node) => node.remove());
  for (const layer of state.doc.layers) {
    const node = createLayer(layer);
    nodes.set(layer.id, node);
    clip.append(node);
  }
}

function createLayer(layer) {
  const node = document.createElement('div');
  node.className = 'layer';
  node.dataset.id = layer.id;
  node.dataset.type = layer.type;
  let content;
  if (layer.type === 'text') {
    content = document.createElement('div');
    content.className = 'content layer-text';
  } else if (layer.type === 'image') {
    content = document.createElement('img');
    content.className = 'content';
    content.alt = '';
    content.draggable = false;
  } else {
    content = document.createElement('div');
    content.className = 'content layer-shape';
  }
  node.append(content);
  node.addEventListener('pointerdown', (event) => onLayerPointerDown(event, layer.id));
  node.addEventListener('dblclick', (event) => onDoubleClick(event, layer.id));
  return node;
}

function paintLayer(node, layer) {
  if (!node) return;
  node.style.left = layer.x + 'px';
  node.style.top = layer.y + 'px';
  node.style.width = layer.w + 'px';
  node.style.height = layer.h + 'px';
  node.style.transform = `rotate(${layer.rotation || 0}deg)`;
  node.style.opacity = String(layer.opacity ?? 1);
  node.style.display = layer.visible === false ? 'none' : 'block';
  node.classList.toggle('is-locked', !!layer.locked);
  node.style.borderRadius = layer.type === 'ellipse' ? '50%' : (layer.radius || 0) + 'px';
  node.style.boxShadow = layer.strokeWidth && layer.stroke ? `inset 0 0 0 ${layer.strokeWidth}px ${layer.stroke}` : 'none';
  const content = node.querySelector('.content');
  if (layer.type === 'text') {
    node.style.background = layer.fill || 'transparent';
    content.style.color = layer.color || '#1a1a1a';
    content.style.fontSize = (layer.fontSize || 32) + 'px';
    content.style.fontFamily = FONT_STACKS[layer.fontFamily] || FONT_STACKS.sans;
    content.style.fontWeight = layer.fontWeight || '400';
    content.style.lineHeight = String(layer.lineHeight || 1.25);
    content.dataset.align = layer.align || 'left';
    if (content.dataset.editing !== '1') content.textContent = layer.text || '';
  } else if (layer.type === 'image') {
    node.style.background = layer.src ? 'transparent' : '#efe8dc';
    content.style.objectFit = layer.fit === 'contain' ? 'contain' : 'cover';
    if (layer.src) {
      if (content.getAttribute('src') !== layer.src) content.src = layer.src;
      content.hidden = false;
    } else {
      content.removeAttribute('src');
      content.hidden = true;
    }
  } else {
    node.style.background = layer.fill || 'transparent';
  }
}

function paintBackground(clip) {
  let img = clip.querySelector('.poster-bg');
  if (!state.doc.backgroundImage) {
    img?.remove();
    return;
  }
  if (!img) {
    img = document.createElement('img');
    img.className = 'poster-bg';
    img.alt = '';
    clip.prepend(img);
  }
  if (img.getAttribute('src') !== state.doc.backgroundImage) img.src = state.doc.backgroundImage;
}

function paintSelection() {
  const box = frame.querySelector('.sel-box');
  const layer = selectedLayer();
  if (!layer || layer.visible === false) {
    box.hidden = true;
    return;
  }
  box.hidden = false;
  box.style.left = layer.x + 'px';
  box.style.top = layer.y + 'px';
  box.style.width = layer.w + 'px';
  box.style.height = layer.h + 'px';
  box.style.transform = `rotate(${layer.rotation || 0}deg)`;
  const locked = layer.locked ? '1' : '0';
  if (box.dataset.for !== layer.id || box.dataset.locked !== locked) {
    box.dataset.for = layer.id;
    box.dataset.locked = locked;
    box.classList.toggle('is-locked', !!layer.locked);
    box.replaceChildren();
    if (!layer.locked) {
      for (const name of ['nw', 'n', 'ne', 'e', 'se', 's', 'sw', 'w']) {
        const handle = document.createElement('button');
        handle.type = 'button';
        handle.className = 'handle';
        handle.dataset.h = name;
        handle.setAttribute('aria-label', '调整大小');
        handle.addEventListener('pointerdown', (event) => startResize(event, name));
        box.append(handle);
      }
      const rot = document.createElement('button');
      rot.type = 'button';
      rot.className = 'handle';
      rot.dataset.h = 'rot';
      rot.setAttribute('aria-label', '旋转');
      rot.addEventListener('pointerdown', startRotate);
      box.append(rot);
    }
  }
}

function paintGuides() {
  frame.querySelector('.guide-v').classList.toggle('is-on', !!state.guides.v);
  frame.querySelector('.guide-h').classList.toggle('is-on', !!state.guides.h);
  frame.querySelector('.guide-v').style.left = state.doc.width / 2 + 'px';
  frame.querySelector('.guide-h').style.top = state.doc.height / 2 + 'px';
}

function ratios() {
  const rect = frame.getBoundingClientRect();
  return { x: state.doc.width / rect.width, y: state.doc.height / rect.height };
}

function blurActive() {
  const el = document.activeElement;
  if (el instanceof HTMLElement && el !== document.body) el.blur();
}

function onLayerPointerDown(event, id) {
  if (event.button !== 0) return;
  const layer = state.doc.layers.find((item) => item.id === id);
  if (!layer || layer.visible === false) return;
  if (event.target.closest('[contenteditable="true"]')) return;
  event.stopPropagation();
  select(id);
  if (layer.locked) return;
  const origin = { x: event.clientX, y: event.clientY };
  const start = { x: layer.x, y: layer.y };
  const ratio = ratios();
  const target = event.currentTarget;
  target.setPointerCapture(event.pointerId);
  let armed = false;
  const move = (ev) => {
    let dx = (ev.clientX - origin.x) * ratio.x;
    let dy = (ev.clientY - origin.y) * ratio.y;
    if (ev.shiftKey) {
      if (Math.abs(dx) > Math.abs(dy)) dy = 0;
      else dx = 0;
    }
    if (!armed) {
      if (Math.hypot(dx, dy) < 3) return;
      blurActive();
      armHistory();
      armed = true;
      setGesturing(true);
      document.body.classList.add('is-dragging');
    }
    const snapped = snapMove(state.doc, start.x + dx, start.y + dy, layer.w, layer.h);
    layer.x = snapped.x;
    layer.y = snapped.y;
    state.guides = snapped.guides;
    touchDirty();
  };
  const up = () => {
    target.removeEventListener('pointermove', move);
    target.removeEventListener('pointerup', up);
    state.guides = { v: false, h: false };
    document.body.classList.remove('is-dragging');
    if (armed) {
      layer.x = Math.round(layer.x);
      layer.y = Math.round(layer.y);
      disarmHistory();
      setGesturing(false);
      touchDirty();
    } else {
      sync();
    }
  };
  target.addEventListener('pointermove', move);
  target.addEventListener('pointerup', up);
}

function startResize(event, handle) {
  event.stopPropagation();
  event.preventDefault();
  const layer = selectedLayer();
  if (!layer || layer.locked) return;
  const start = { x: layer.x, y: layer.y, w: layer.w, h: layer.h, rotation: layer.rotation || 0 };
  const origin = { x: event.clientX, y: event.clientY };
  const ratio = ratios();
  const target = event.currentTarget;
  target.setPointerCapture(event.pointerId);
  let armed = false;
  const move = (ev) => {
    const dx = (ev.clientX - origin.x) * ratio.x;
    const dy = (ev.clientY - origin.y) * ratio.y;
    if (!armed) {
      if (Math.hypot(dx, dy) < 2) return;
      blurActive();
      armHistory();
      armed = true;
      setGesturing(true);
      document.body.classList.add('is-dragging');
    }
    const next = resizeLayer(start, handle, dx, dy);
    layer.x = next.x;
    layer.y = next.y;
    layer.w = next.w;
    layer.h = next.h;
    touchDirty();
  };
  const up = () => {
    target.removeEventListener('pointermove', move);
    target.removeEventListener('pointerup', up);
    document.body.classList.remove('is-dragging');
    if (armed) {
      layer.x = Math.round(layer.x);
      layer.y = Math.round(layer.y);
      layer.w = Math.round(layer.w);
      layer.h = Math.round(layer.h);
      disarmHistory();
      setGesturing(false);
      touchDirty();
    }
  };
  target.addEventListener('pointermove', move);
  target.addEventListener('pointerup', up);
}

function startRotate(event) {
  event.stopPropagation();
  event.preventDefault();
  const layer = selectedLayer();
  if (!layer || layer.locked) return;
  const ratio = ratios();
  const rect = frame.getBoundingClientRect();
  const cx = rect.left + (layer.x + layer.w / 2) / ratio.x;
  const cy = rect.top + (layer.y + layer.h / 2) / ratio.y;
  const a0 = Math.atan2(event.clientY - cy, event.clientX - cx);
  const start = layer.rotation || 0;
  const target = event.currentTarget;
  target.setPointerCapture(event.pointerId);
  let armed = false;
  const move = (ev) => {
    const a1 = Math.atan2(ev.clientY - cy, ev.clientX - cx);
    let deg = start + ((a1 - a0) * 180) / Math.PI;
    if (!armed) {
      if (Math.abs(deg - start) < 1) return;
      blurActive();
      armHistory();
      armed = true;
      setGesturing(true);
    }
    deg = snapAngle(deg, ev.shiftKey);
    layer.rotation = deg;
    touchDirty();
  };
  const up = () => {
    target.removeEventListener('pointermove', move);
    target.removeEventListener('pointerup', up);
    if (armed) {
      layer.rotation = Math.round(layer.rotation * 10) / 10;
      disarmHistory();
      setGesturing(false);
      touchDirty();
    }
  };
  target.addEventListener('pointermove', move);
  target.addEventListener('pointerup', up);
}

function snapAngle(deg, force) {
  if (force) return Math.round(deg / 15) * 15;
  const mod = ((deg % 90) + 90) % 90;
  if (mod < 4) return deg - mod;
  if (mod > 86) return deg + (90 - mod);
  return deg;
}

function onDoubleClick(event, id) {
  const layer = state.doc.layers.find((item) => item.id === id);
  if (!layer || layer.type !== 'text' || layer.locked) return;
  const content = event.currentTarget.querySelector('.layer-text');
  if (!content || content.dataset.editing === '1') return;
  event.stopPropagation();
  content.contentEditable = 'true';
  content.dataset.editing = '1';
  state.editingText = true;
  armHistory();
  content.focus();
  const range = document.createRange();
  range.selectNodeContents(content);
  const selection = getSelection();
  selection.removeAllRanges();
  selection.addRange(range);
  const finish = () => {
    if (content.dataset.editing !== '1') return;
    content.contentEditable = 'false';
    delete content.dataset.editing;
    state.editingText = false;
    const next = normalizeText(content.innerText);
    const changed = next !== (layer.text || '');
    layer.text = next;
    disarmHistory();
    if (changed) touchDirty();
    else sync();
  };
  content.addEventListener('blur', finish, { once: true });
  content.addEventListener('keydown', (ev) => {
    if (ev.key === 'Escape') {
      ev.preventDefault();
      ev.stopPropagation();
      content.blur();
    }
  });
}

export function exitTextEdit() {
  const editing = document.querySelector('.layer-text[contenteditable="true"]');
  editing?.blur();
}

function pointFromEvent(event) {
  const rect = frame.getBoundingClientRect();
  return {
    x: ((event.clientX - rect.left) * state.doc.width) / rect.width,
    y: ((event.clientY - rect.top) * state.doc.height) / rect.height,
  };
}

async function onDrop(event) {
  event.preventDefault();
  stage.classList.remove('is-drop');
  const file = [...event.dataTransfer.files].find((item) => item.type.startsWith('image/'));
  if (file) {
    await dropFiles(file, pointFromEvent(event));
    return;
  }
  const src = event.dataTransfer.getData('application/x-poster-asset');
  if (!src) return;
  await dropAsset(src, pointFromEvent(event));
}

export function zoomLabel() {
  if (!state.doc || !stage) return '100%';
  return Math.round(fitScale() * state.zoom * 100) + '%';
}
