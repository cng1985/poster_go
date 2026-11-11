import { savePoster } from './api.js';
import { uid } from './id.js';

export const state = {
  doc: null,
  selectedId: null,
  libraryMode: null,
  leftTab: 'layers',
  zoom: 1,
  guides: { v: false, h: false },
  editingText: false,
  posters: [],
  library: [],
  templates: [],
  profiles: [],
  activeProfileId: '',
  saveState: 'saved',
  saveError: '',
};

const listeners = new Set();
let history = [];
let future = [];
let historyArmed = false;
let gesturing = false;
let rev = 0;
let savedRev = 0;
let timer = 0;
let lastSaveError = '';
let reportSaveError = () => {};

export function subscribe(fn) {
  listeners.add(fn);
  return () => listeners.delete(fn);
}

export function emit(type) {
  for (const fn of listeners) fn(type);
}

export function onSaveError(fn) {
  reportSaveError = fn;
}

export function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

export function isGesturing() {
  return gesturing;
}

export function setGesturing(value) {
  gesturing = value;
}

export function canUndo() {
  return history.length > 0;
}

export function canRedo() {
  return future.length > 0;
}

export function selectedLayer() {
  return state.doc?.layers.find((layer) => layer.id === state.selectedId) || null;
}

function pushHistory() {
  if (!state.doc) return;
  history.push(clone(state.doc));
  if (history.length > 60) history.shift();
  future = [];
}

export function armHistory() {
  if (historyArmed || !state.doc) return;
  pushHistory();
  historyArmed = true;
}

export function disarmHistory() {
  if (!historyArmed) return;
  historyArmed = false;
  const prev = history[history.length - 1];
  if (prev && JSON.stringify(prev) === JSON.stringify(state.doc)) history.pop();
}

function markDirty(type) {
  rev += 1;
  state.saveState = 'dirty';
  emit(type);
  scheduleSave();
}

export function touchDirty() {
  markDirty('ephemeral');
}

export function patch(mutator) {
  mutator();
  markDirty('ephemeral');
}

export function commit(mutator) {
  if (historyArmed) disarmHistory();
  pushHistory();
  mutator();
  markDirty('structure');
}

export function undo() {
  if (historyArmed) disarmHistory();
  if (!history.length || !state.doc) return;
  future.push(clone(state.doc));
  state.doc = history.pop();
  if (!state.doc.layers.some((layer) => layer.id === state.selectedId)) state.selectedId = null;
  markDirty('structure');
}

export function redo() {
  if (!future.length || !state.doc) return;
  history.push(clone(state.doc));
  state.doc = future.pop();
  if (!state.doc.layers.some((layer) => layer.id === state.selectedId)) state.selectedId = null;
  markDirty('structure');
}

export function loadDoc(doc) {
  state.doc = doc;
  if (!Array.isArray(state.doc.layers)) state.doc.layers = [];
  state.selectedId = null;
  state.libraryMode = null;
  state.guides = { v: false, h: false };
  state.editingText = false;
  history = [];
  future = [];
  historyArmed = false;
  rev = 0;
  savedRev = 0;
  state.saveState = 'saved';
  state.saveError = '';
  localStorage.setItem('postergo-id', doc.id);
  emit('structure');
}

export function select(id) {
  if (state.selectedId === id) return;
  state.selectedId = id;
  emit('selection');
}

export function setTab(tab) {
  state.leftTab = tab;
  emit('panel');
}

export function setLibraryMode(mode) {
  state.libraryMode = mode;
  if (mode) state.leftTab = 'library';
  emit('panel');
}

export function normalizeText(value) {
  return String(value || '').replace(/\u00a0/g, ' ').replace(/\r/g, '').replace(/\n$/, '');
}

export function currentTexts() {
  return state.doc.layers
    .filter((layer) => layer.type === 'text' && layer.text)
    .map((layer) => layer.text)
    .slice(0, 20);
}

function textLayer(role, text, preset) {
  return {
    id: uid('l'),
    name: preset.name,
    type: 'text',
    role,
    text,
    x: preset.x,
    y: preset.y,
    w: preset.w,
    h: preset.h,
    rotation: 0,
    opacity: 1,
    visible: true,
    locked: false,
    fontSize: preset.fontSize,
    fontFamily: 'sans',
    fontWeight: preset.fontWeight,
    color: preset.color,
    align: preset.align,
    lineHeight: preset.lineHeight || 1.25,
    fill: '',
    radius: 0,
  };
}

function upsertRole(doc, role, text, preset) {
  const found = doc.layers.find((layer) => layer.type === 'text' && layer.role === role);
  if (found) {
    if (text) found.text = text;
    return;
  }
  if (!text) return;
  doc.layers.push(textLayer(role, text, preset));
}

export function applyCopy(doc, copy) {
  const w = doc.width;
  const h = doc.height;
  upsertRole(doc, 'title', copy.title || '', {
    name: '主标题',
    fontSize: Math.max(32, Math.round(w * 0.075)),
    fontWeight: '800',
    w: Math.round(w * 0.84),
    h: Math.round(w * 0.22),
    x: Math.round(w * 0.08),
    y: Math.round(h * 0.16),
    align: 'center',
    color: '#1c1915',
    lineHeight: 1.1,
  });
  upsertRole(doc, 'subtitle', copy.subtitle || '', {
    name: '副标题',
    fontSize: Math.max(18, Math.round(w * 0.03)),
    fontWeight: '500',
    w: Math.round(w * 0.8),
    h: Math.round(w * 0.1),
    x: Math.round(w * 0.1),
    y: Math.round(h * 0.36),
    align: 'center',
    color: '#5c564e',
  });
  let body = copy.body || '';
  if (copy.bullets?.length) {
    const lines = copy.bullets.map((item) => '· ' + item).join('\n');
    body = body ? body + '\n' + lines : lines;
  }
  upsertRole(doc, 'body', body, {
    name: '正文',
    fontSize: Math.max(16, Math.round(w * 0.026)),
    fontWeight: '400',
    w: Math.round(w * 0.76),
    h: Math.round(h * 0.24),
    x: Math.round(w * 0.12),
    y: Math.round(h * 0.48),
    align: 'center',
    color: '#2a2722',
    lineHeight: 1.45,
  });
  state.selectedId = doc.layers.find((layer) => layer.role === 'title')?.id || state.selectedId;
}

export function applyLayout(doc, layout) {
  if (layout.name) doc.name = layout.name;
  if (layout.background) doc.background = layout.background;
  doc.backgroundImage = layout.backgroundImage || '';
  doc.layers = Array.isArray(layout.layers) ? layout.layers : [];
  state.selectedId = doc.layers.find((layer) => layer.role === 'title')?.id || doc.layers.at(-1)?.id || null;
}

export function applyTemplate(doc, templateDoc) {
  const next = clone(templateDoc);
  doc.name = next.name;
  doc.width = next.width;
  doc.height = next.height;
  doc.background = next.background;
  doc.backgroundImage = next.backgroundImage || '';
  doc.layers = next.layers || [];
  state.selectedId = null;
}

function scheduleSave() {
  clearTimeout(timer);
  timer = setTimeout(() => {
    saveNow().catch(() => {});
  }, 700);
}

let saveTail = Promise.resolve();

export function saveNow() {
  clearTimeout(timer);
  const run = saveTail.then(persist, persist);
  saveTail = run.then(() => {}, () => {});
  return run;
}

async function persist() {
  if (!state.doc || rev === savedRev) {
    state.saveState = 'saved';
    emit('chrome');
    return;
  }
  const my = rev;
  const body = clone(state.doc);
  state.saveState = 'saving';
  emit('chrome');
  let saved;
  try {
    saved = await savePoster(body);
  } catch (err) {
    state.saveState = 'error';
    state.saveError = err.message || '保存失败';
    if (state.saveError !== lastSaveError) {
      lastSaveError = state.saveError;
      reportSaveError(state.saveError);
    }
    emit('chrome');
    throw err;
  }
  if (state.doc && state.doc.id === body.id) {
    state.doc.updatedAt = saved.updatedAt;
    if (saved.createdAt) state.doc.createdAt = saved.createdAt;
  }
  if (my >= savedRev) savedRev = my;
  if (rev !== savedRev) return persist();
  state.saveState = 'saved';
  state.saveError = '';
  lastSaveError = '';
  emit('chrome');
}

export function beaconSave() {
  if (!state.doc || rev === savedRev) return;
  fetch('/api/posters/' + encodeURIComponent(state.doc.id), {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(state.doc),
    keepalive: true,
  });
}
