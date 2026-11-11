import * as api from './api.js';
import { exitTextEdit, mount, setDropHandlers, sync, zoomLabel } from './canvas.js';
import { icon } from './dom.js';
import { exportPNG } from './export.js';
import {
  armHistory,
  beaconSave,
  canRedo,
  canUndo,
  disarmHistory,
  loadDoc,
  onSaveError,
  patch,
  redo,
  saveNow,
  setTab,
  state,
  subscribe,
  undo,
} from './model.js';
import {
  closeTopModal,
  deleteSelected,
  duplicateSelected,
  endNudge,
  isModalOpen,
  nudgeSelected,
  openAI,
  openHelp,
  openModels,
  openPosters,
  placeAssetSrc,
  placeUpload,
  renderInspector,
  renderLeft,
  syncLive,
  syncSelection,
  toast,
} from './panels.js';

const nameInput = document.getElementById('poster-name');

function updateChrome() {
  if (!state.doc) return;
  if (document.activeElement !== nameInput) nameInput.value = state.doc.name || '';
  document.title = (state.doc.name || '未命名海报') + ' · PosterGo';
  document.getElementById('btn-undo').disabled = !canUndo();
  document.getElementById('btn-redo').disabled = !canRedo();
  document.getElementById('status-size').textContent = state.doc.width + ' × ' + state.doc.height;
  const labels = { saved: '已保存', saving: '保存中', dirty: '未保存', error: '保存失败' };
  const save = document.getElementById('status-save');
  save.textContent = labels[state.saveState] || '';
  save.classList.toggle('is-error', state.saveState === 'error');
  document.getElementById('status-zoom').textContent = zoomLabel();
  const profile = state.profiles.find((item) => item.id === state.activeProfileId);
  document.getElementById('status-model').textContent = profile ? profile.name : '未配置模型';
  document.querySelectorAll('.rail button').forEach((btn) => {
    btn.setAttribute('aria-pressed', btn.dataset.tab === state.leftTab ? 'true' : 'false');
  });
}

function onEvent(type) {
  if (type === 'structure' || type === 'selection' || type === 'ephemeral') sync();
  if (type === 'structure' || type === 'panel') renderLeft();
  if (type === 'structure' || type === 'selection') renderInspector();
  if (type === 'selection') syncSelection();
  if (type === 'ephemeral') syncLive();
  updateChrome();
}

function isTyping(target) {
  if (!target || !target.tagName) return false;
  const tag = target.tagName;
  return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || target.isContentEditable;
}

async function createNew() {
  try {
    await saveNow();
    const doc = await api.createPoster({
      name: '未命名海报',
      width: state.doc?.width || 1080,
      height: state.doc?.height || 1520,
      background: '#f6f1e8',
      layers: [],
    });
    loadDoc(doc);
    toast('已新建海报');
  } catch (err) {
    toast(err.message, 'error');
  }
}

async function doExport() {
  try {
    await saveNow();
    await exportPNG(state.doc);
    toast('已导出 PNG');
  } catch (err) {
    toast(err.message || '导出失败', 'error');
  }
}

function onKey(event) {
  if (isModalOpen()) {
    if (event.key === 'Escape') {
      event.preventDefault();
      closeTopModal();
    }
    return;
  }
  const meta = event.ctrlKey || event.metaKey;
  const key = event.key.toLowerCase();
  if (meta && key === 's') {
    event.preventDefault();
    saveNow().catch((err) => toast(err.message, 'error'));
    return;
  }
  if (meta && (key === 'z' || key === 'y')) {
    event.preventDefault();
    exitTextEdit();
    if (key === 'y' || event.shiftKey) redo();
    else undo();
    return;
  }
  if (meta && key === 'd') {
    event.preventDefault();
    duplicateSelected();
    return;
  }
  if (meta && event.key === '0') {
    event.preventDefault();
    state.zoom = 1;
    sync();
    updateChrome();
    return;
  }
  if (isTyping(event.target)) return;
  if (event.key === 'Escape') {
    exitTextEdit();
    return;
  }
  if (event.key === 'Delete' || event.key === 'Backspace') {
    event.preventDefault();
    deleteSelected();
    return;
  }
  const step = event.shiftKey ? 10 : 1;
  const delta = {
    ArrowLeft: [-step, 0],
    ArrowRight: [step, 0],
    ArrowUp: [0, -step],
    ArrowDown: [0, step],
  }[event.key];
  if (delta) {
    event.preventDefault();
    nudgeSelected(delta[0], delta[1]);
  }
}

function bind() {
  document.querySelectorAll('[data-ico]').forEach((el) => el.append(icon(el.dataset.ico)));
  nameInput.addEventListener('focus', armHistory);
  nameInput.addEventListener('input', () => patch(() => { state.doc.name = nameInput.value; }));
  nameInput.addEventListener('blur', disarmHistory);
  document.getElementById('btn-posters').addEventListener('click', () => { openPosters(); });
  document.getElementById('btn-new').addEventListener('click', () => { createNew(); });
  document.getElementById('btn-undo').addEventListener('click', () => { exitTextEdit(); undo(); });
  document.getElementById('btn-redo').addEventListener('click', () => { exitTextEdit(); redo(); });
  document.getElementById('btn-ai').addEventListener('click', () => { openAI(); });
  document.getElementById('btn-models').addEventListener('click', () => { openModels(); });
  document.getElementById('btn-export').addEventListener('click', () => { doExport(); });
  document.getElementById('btn-help').addEventListener('click', () => { openHelp(); });
  document.getElementById('status-model').addEventListener('click', () => { openModels(); });
  document.querySelectorAll('.rail button').forEach((btn) => {
    btn.addEventListener('click', () => setTab(btn.dataset.tab));
  });
  const stage = document.getElementById('stage');
  stage.addEventListener('wheel', (event) => {
    if (!event.ctrlKey && !event.metaKey) return;
    event.preventDefault();
    const factor = event.deltaY > 0 ? 0.92 : 1.08;
    state.zoom = Math.min(6, Math.max(0.2, state.zoom * factor));
    sync();
    updateChrome();
  }, { passive: false });
  window.addEventListener('dragover', (event) => {
    if ([...event.dataTransfer.types].includes('Files')) event.preventDefault();
  });
  window.addEventListener('drop', (event) => {
    if (event.target.closest && event.target.closest('#stage')) return;
    if ([...event.dataTransfer.types].includes('Files')) event.preventDefault();
  });
  document.addEventListener('keydown', onKey);
  document.addEventListener('keyup', (event) => {
    if (event.key.startsWith('Arrow')) endNudge();
  });
  window.addEventListener('beforeunload', (event) => {
    if (state.saveState === 'saved') return;
    beaconSave();
    event.preventDefault();
    event.returnValue = '';
  });
  window.addEventListener('pagehide', () => beaconSave());
}

async function boot() {
  const stage = document.getElementById('stage');
  mount(stage);
  setDropHandlers({ asset: placeAssetSrc, files: placeUpload });
  bind();
  onSaveError((message) => toast(message, 'error'));
  subscribe(onEvent);
  try {
    const [posters, library, templates, profileData] = await Promise.all([
      api.listPosters(),
      api.library(),
      api.templates(),
      api.profiles(),
    ]);
    state.posters = posters || [];
    state.library = library.items || [];
    state.templates = templates || [];
    state.profiles = profileData.profiles || [];
    state.activeProfileId = profileData.activeProfileId || '';
    const saved = localStorage.getItem('postergo-id');
    const known = (posters || []).some((item) => item.id === saved);
    const id = known ? saved : posters?.[0]?.id;
    const doc = id ? await api.getPoster(id) : await api.createPoster({});
    loadDoc(doc);
  } catch (err) {
    stage.replaceChildren();
    const note = document.createElement('p');
    note.className = 'boot-error';
    note.textContent = '无法打开编辑器：' + (err.message || '未知错误');
    stage.append(note);
  }
}

boot();
