import * as api from './api.js';
import { COMPONENT_GROUPS, insertAsset } from './components.js';
import { clear, h, icon } from './dom.js';
import { FONT_LABELS } from './fonts.js';
import { reorderLayers, scaleLayers } from './geom.js';
import { uid } from './id.js';
import {
  applyCopy,
  applyLayout,
  applyTemplate,
  armHistory,
  clone,
  commit,
  currentTexts,
  disarmHistory,
  patch,
  saveNow,
  select,
  selectedLayer,
  setLibraryMode,
  setTab,
  state,
  loadDoc,
} from './model.js';

const CATS = [
  ['all', '全部'],
  ['background', '背景'],
  ['texture', '纹理'],
  ['illustration', '插画'],
  ['icon', '装饰'],
  ['upload', '我的'],
];
const SIZES = [
  ['竖版', 1080, 1520],
  ['方形', 1080, 1080],
  ['故事', 1080, 1920],
  ['横版', 1920, 1080],
  ['小红书', 1242, 1660],
  ['A4', 794, 1123],
];
const PRESETS = [
  ['custom', '自定义', '', ''],
  ['openai', 'OpenAI', 'https://api.openai.com/v1', 'gpt-4o-mini'],
  ['deepseek', 'DeepSeek', 'https://api.deepseek.com/v1', 'deepseek-chat'],
  ['moonshot', 'Moonshot', 'https://api.moonshot.cn/v1', 'moonshot-v1-8k'],
  ['qwen', '通义千问兼容', 'https://dashscope.aliyuncs.com/compatible-mode/v1', 'qwen-plus'],
  ['ollama', 'Ollama 本地', 'http://127.0.0.1:11434/v1', 'llama3.2'],
  ['lmstudio', 'LM Studio', 'http://127.0.0.1:1234/v1', ''],
];

let layerQuery = '';
let libQuery = '';
let libCat = 'all';
let aiPrompt = '';
const closers = [];

export function isModalOpen() {
  return closers.length > 0;
}

export function closeTopModal() {
  const close = closers[closers.length - 1];
  if (!close) return false;
  close();
  return true;
}

function openModal(className, build) {
  const root = document.getElementById('modal-root');
  const back = h('div', { class: 'modal-back' });
  const modal = h('div', { class: 'modal ' + className, role: 'dialog', 'aria-modal': 'true' });
  back.append(modal);
  let closed = false;
  function close() {
    if (closed) return;
    closed = true;
    const index = closers.indexOf(close);
    if (index >= 0) closers.splice(index, 1);
    back.remove();
  }
  closers.push(close);
  back.addEventListener('click', (event) => {
    if (event.target === back) close();
  });
  root.append(back);
  try {
    build(modal, close);
  } catch (err) {
    close();
    throw err;
  }
  if (!closed) modal.querySelector('input, textarea, button')?.focus();
  return close;
}

export function toast(message, kind = 'info') {
  const root = document.getElementById('toast-root');
  const node = h('div', { class: 'toast' + (kind === 'error' ? ' is-error' : ''), role: 'status', text: message });
  root.append(node);
  while (root.children.length > 3) root.firstElementChild.remove();
  setTimeout(() => node.remove(), 3200);
}

export function confirmDialog({ title, body, confirmText = '确定', danger = false }) {
  return new Promise((resolve) => {
    openModal('modal-sm', (modal, close) => {
      const done = (value) => {
        close();
        resolve(value);
      };
      modal.append(
        h('h2', { text: title }),
        h('p', { text: body }),
        h('div', { class: 'modal-actions' }, [
          h('button', { type: 'button', class: 'btn', text: '取消', onClick: () => done(false) }),
          h('button', { type: 'button', class: danger ? 'btn btn-danger' : 'btn btn-primary', text: confirmText, onClick: () => done(true) }),
        ]),
      );
    });
  });
}

export function renderLeft() {
  const root = document.getElementById('left-panel');
  if (!root || !state.doc) return;
  const focus = captureFocus();
  clear(root);
  if (state.leftTab === 'layers') root.append(renderLayers());
  if (state.leftTab === 'components') root.append(renderComponents());
  if (state.leftTab === 'library') root.append(renderLibrary());
  if (state.leftTab === 'templates') root.append(renderTemplates());
  restoreFocus(focus);
}

export function renderInspector() {
  const root = document.getElementById('inspector');
  if (!root || !state.doc) return;
  clear(root);
  const layer = selectedLayer();
  root.append(layer ? layerInspector(layer) : posterInspector());
}

export function syncSelection() {
  document.querySelectorAll('.layer-row').forEach((row) => {
    row.classList.toggle('is-selected', row.dataset.id === state.selectedId);
  });
}

export function syncLive() {
  const layer = selectedLayer();
  const root = document.getElementById('inspector');
  if (!layer || !root) return;
  root.querySelectorAll('[data-bind]').forEach((input) => {
    if (document.activeElement === input) return;
    const value = layer[input.dataset.bind];
    if (input.type === 'color') {
      const next = toColorInput(value);
      if (input.value !== next) input.value = next;
      return;
    }
    const shown = value == null ? '' : String(roundIfNumber(value));
    if (input.value !== shown) input.value = shown;
  });
  root.querySelectorAll('[data-choice]').forEach((btn) => {
    const [key, val] = btn.dataset.choice.split(':');
    btn.classList.toggle('is-active', String(layer[key] ?? '') === val);
  });
  const label = root.querySelector('[data-opacity-label]');
  if (label) label.textContent = Math.round((layer.opacity ?? 1) * 100) + '%';
}

function renderLayers() {
  const view = state.doc.layers.map((layer, index) => ({ layer, index })).reverse();
  const q = layerQuery.trim();
  const rows = view.filter(({ layer }) => !q || layer.name.includes(q) || (layer.text || '').includes(q));
  const list = h('div', { class: 'layer-list' });
  if (!rows.length) list.append(h('p', { class: 'panel-note', text: state.doc.layers.length ? '没有匹配的图层' : '还没有图层' }));
  rows.forEach(({ layer }, viewIndex) => {
    const realIndex = view.findIndex((item) => item.layer.id === layer.id);
    const row = h('div', {
      class: 'layer-row' + (layer.id === state.selectedId ? ' is-selected' : '') + (layer.visible === false ? ' is-hidden' : ''),
      dataset: { id: layer.id },
      onClick: () => select(layer.id),
    }, [
      h('button', { type: 'button', class: 'mini grip', title: '拖动排序', 'aria-label': '拖动排序' }, [icon('grip')]),
      h('span', { class: 'type-mark', text: typeMark(layer) }),
      h('span', { class: 'layer-name', text: layer.name || '未命名' }),
      h('button', {
        type: 'button', class: 'mini', title: layer.visible === false ? '显示' : '隐藏',
        'aria-label': layer.visible === false ? '显示' : '隐藏', 'aria-pressed': layer.visible !== false ? 'true' : 'false',
        onClick: (event) => { event.stopPropagation(); commit(() => { layer.visible = layer.visible === false; }); },
      }, [icon(layer.visible === false ? 'eyeOff' : 'eye')]),
      h('button', {
        type: 'button', class: 'mini', title: layer.locked ? '解锁' : '锁定',
        'aria-label': layer.locked ? '解锁' : '锁定', 'aria-pressed': layer.locked ? 'true' : 'false',
        onClick: (event) => { event.stopPropagation(); commit(() => { layer.locked = !layer.locked; }); },
      }, [icon(layer.locked ? 'lock' : 'unlock')]),
    ]);
    if (!q) bindReorder(row.querySelector('.grip'), realIndex, list);
    list.append(row);
  });
  return h('div', {}, [
    h('div', { class: 'panel-head' }, [
      h('h2', { class: 'panel-title', text: '图层' }),
      h('p', { class: 'panel-note', text: '上面的图层盖住下面的。拖动手柄可以调整顺序。' }),
    ]),
    h('input', {
      id: 'layer-filter', class: 'search', placeholder: '搜索图层', value: layerQuery,
      onInput: (event) => { layerQuery = event.target.value; renderLeft(); },
    }),
    list,
  ]);
}

function bindReorder(grip, fromView, list) {
  grip.addEventListener('pointerdown', (event) => {
    event.preventDefault();
    event.stopPropagation();
    const rows = [...list.querySelectorAll('.layer-row')];
    grip.setPointerCapture(event.pointerId);
    let target = fromView;
    const move = (ev) => {
      target = rows.length - 1;
      rows.forEach((row, index) => {
        const rect = row.getBoundingClientRect();
        if (ev.clientY < rect.top + rect.height / 2 && target === rows.length - 1) target = index;
        row.classList.toggle('is-drop', false);
      });
      rows.forEach((row, index) => row.classList.toggle('is-drop', index === target));
    };
    const up = () => {
      grip.removeEventListener('pointermove', move);
      grip.removeEventListener('pointerup', up);
      if (target !== fromView) commit(() => { state.doc.layers = reorderLayers(state.doc.layers, fromView, target); });
    };
    grip.addEventListener('pointermove', move);
    grip.addEventListener('pointerup', up);
  });
}

function renderComponents() {
  const q = '';
  const box = h('div');
  box.append(h('div', { class: 'panel-head' }, [
    h('h2', { class: 'panel-title', text: '组件' }),
    h('p', { class: 'panel-note', text: '点一下就加到画布中间，然后拖动、改字、换色。' }),
  ]));
  for (const group of COMPONENT_GROUPS) {
    const items = group.items.filter((item) => !q || item.name.includes(q));
    if (!items.length) continue;
    box.append(h('h3', { class: 'group-title', text: group.title }));
    box.append(h('div', { class: 'comp-grid' }, items.map((item) => h('button', {
      type: 'button', class: 'comp-card', onClick: () => addComponent(item),
    }, [
      h('span', { class: 'preview', 'data-kind': item.preview }, [h('span', { class: 'sw', text: item.preview === 'title' ? 'Aa' : item.preview === 'sub' ? 'Aa' : item.preview === 'body' ? '文' : '' })]),
      h('span', {}, [h('strong', { text: item.name }), h('small', { text: item.hint })]),
    ]))));
  }
  return box;
}

function addComponent(item) {
  if (item.action === 'library') {
    setTab('library');
    return;
  }
  if (item.action === 'upload') {
    pickFile();
    return;
  }
  commit(() => {
    const layers = item.build(state.doc);
    state.doc.layers.push(...layers);
    state.selectedId = layers.at(-1).id;
  });
}

function renderLibrary() {
  const q = libQuery.trim();
  const items = state.library.filter((item) => (libCat === 'all' || item.category === libCat) && (!q || item.name.includes(q)));
  const banner = state.libraryMode ? h('div', { class: 'banner' }, [
    h('span', { text: state.libraryMode.type === 'background' ? '选择一张图片作为画布背景' : '选择图片替换当前图层' }),
    h('button', { type: 'button', class: 'btn btn-small', text: '取消', onClick: () => setLibraryMode(null) }),
  ]) : null;
  return h('div', { id: 'library-panel' }, [
    h('div', { class: 'panel-head' }, [
      h('h2', { class: 'panel-title', text: '图片库' }),
      h('p', { class: 'panel-note', text: '点图片加入画布，也可以拖到海报上。背景类图片会铺满并放到最底层。' }),
    ]),
    banner,
    h('div', { class: 'lib-actions' }, [
      h('button', { type: 'button', class: 'btn btn-small', text: '上传图片', onClick: pickFile }),
    ]),
    h('input', {
      id: 'lib-filter', class: 'search', placeholder: '搜索图片', value: libQuery,
      onInput: (event) => { libQuery = event.target.value; renderLeft(); },
    }),
    h('div', { class: 'chips' }, CATS.map(([id, label]) => h('button', {
      type: 'button', class: 'chip' + (libCat === id ? ' is-active' : ''), text: label,
      onClick: () => { libCat = id; renderLeft(); },
    }))),
    items.length ? h('div', { class: 'lib-grid' }, items.map((item) => libraryCard(item))) : h('p', { class: 'panel-note', text: '这个分类里还没有图片' }),
  ]);
}

function libraryCard(item) {
  const card = h('div', { class: 'lib-card', draggable: 'true', tabindex: '0', role: 'button' }, [
    h('img', { src: item.src, alt: item.name }),
    h('span', { class: 'lib-meta' }, [
      h('strong', { text: item.name }),
      item.builtin === false ? h('button', {
        type: 'button', class: 'mini danger-text', text: '删除', title: '删除上传的图片',
        onClick: (event) => { event.stopPropagation(); removeUpload(item); },
      }) : null,
    ]),
  ]);
  card.addEventListener('click', () => useAsset(item));
  card.addEventListener('keydown', (event) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      useAsset(item);
    }
  });
  card.addEventListener('dragstart', (event) => {
    event.dataTransfer.setData('application/x-poster-asset', item.src);
    event.dataTransfer.effectAllowed = 'copy';
  });
  return card;
}

function renderTemplates() {
  return h('div', {}, [
    h('div', { class: 'panel-head' }, [
      h('h2', { class: 'panel-title', text: '模板' }),
      h('p', { class: 'panel-note', text: '套用后会替换当前图层，可以用撤销退回。' }),
    ]),
    h('div', { class: 'tpl-list' }, state.templates.map((tpl) => {
      const swatch = h('span', { class: 'tpl-swatch' });
      swatch.style.background = tpl.accent || '#e23d2b';
      return h('button', {
        type: 'button', class: 'tpl-card', onClick: () => applyTemplateCard(tpl),
      }, [
        swatch,
        h('strong', { text: tpl.name }),
        h('small', { text: tpl.description + ' · ' + tpl.document.width + '×' + tpl.document.height }),
      ]);
    })),
  ]);
}

function applyTemplateCard(tpl) {
  const run = () => {
    commit(() => applyTemplate(state.doc, tpl.document));
    toast('已套用「' + tpl.name + '」');
  };
  if (!state.doc.layers.length) {
    run();
    return;
  }
  confirmDialog({ title: '套用「' + tpl.name + '」', body: '当前图层会被替换。这个操作可以撤销。', confirmText: '替换' }).then((ok) => {
    if (ok) run();
  });
}

function posterInspector() {
  const bg = h('input', { type: 'color', value: toColorInput(state.doc.background) });
  bg.addEventListener('focus', armHistory);
  bg.addEventListener('input', () => patch(() => { state.doc.background = bg.value; }));
  bg.addEventListener('blur', disarmHistory);
  const width = h('input', { type: 'number', min: '240', max: '4096', value: state.doc.width });
  const height = h('input', { type: 'number', min: '240', max: '4096', value: state.doc.height });
  width.addEventListener('change', () => resizeCanvas(Number(width.value), state.doc.height));
  height.addEventListener('change', () => resizeCanvas(state.doc.width, Number(height.value)));
  return h('div', {}, [
    h('div', { class: 'insp-head' }, [h('h2', { text: '画布' })]),
    h('p', { class: 'panel-note', text: '没有选中图层时，这里调整整张海报。' }),
    h('div', { class: 'size-presets' }, SIZES.map(([label, w, hgt]) => h('button', {
      type: 'button',
      class: 'chip' + (state.doc.width === w && state.doc.height === hgt ? ' is-active' : ''),
      text: label,
      onClick: () => resizeCanvas(w, hgt),
    }))),
    h('div', { class: 'grid-2' }, [field('宽度', width), field('高度', height)]),
    field('背景色', bg),
    h('div', { class: 'field' }, [
      h('span', { text: '背景图' }),
      state.doc.backgroundImage ? h('img', { class: 'bg-preview', src: state.doc.backgroundImage, alt: '背景图' }) : h('p', { class: 'panel-note', text: '背景图铺在所有图层后面。' }),
      h('div', { class: 'inline-actions' }, [
        h('button', { type: 'button', class: 'btn btn-small', text: '从图片库选择', onClick: () => setLibraryMode({ type: 'background' }) }),
        state.doc.backgroundImage ? h('button', { type: 'button', class: 'btn btn-small', text: '清除', onClick: () => commit(() => { state.doc.backgroundImage = ''; }) }) : null,
      ]),
    ]),
  ]);
}

function layerInspector(layer) {
  const box = h('div');
  box.append(h('div', { class: 'insp-head' }, [
    h('h2', { text: '图层' }),
    h('button', { type: 'button', class: 'btn btn-small', text: '画布', onClick: () => select(null) }),
  ]));
  box.append(field('名称', bindText(h('input', { type: 'text', 'data-bind': 'name', value: layer.name, maxlength: '40' }), layer.id, 'name')));
  if (layer.type === 'text') {
    box.append(field('文字', bindText(h('textarea', { 'data-bind': 'text', rows: '4', value: layer.text || '' }), layer.id, 'text')));
    const family = h('select', { 'data-bind': 'fontFamily' }, FONT_LABELS.map(([value, label]) => h('option', { value, text: label })));
    family.value = layer.fontFamily || 'sans';
    family.addEventListener('change', () => commit(() => { layer.fontFamily = family.value; }));
    const weight = h('select', {}, [['400', '常规'], ['500', '中等'], ['600', '半粗'], ['700', '粗体'], ['800', '特粗']].map(([value, label]) => h('option', { value, text: label })));
    weight.value = layer.fontWeight || '400';
    weight.addEventListener('change', () => commit(() => { layer.fontWeight = weight.value; }));
    box.append(h('div', { class: 'grid-2' }, [field('字体', family), field('字重', weight)]));
    box.append(h('div', { class: 'grid-2' }, [
      field('字号', bindNumber(h('input', { type: 'number', 'data-bind': 'fontSize', min: '8', max: '400', value: layer.fontSize || 32 }), layer.id, 'fontSize')),
      field('行高', bindNumber(h('input', { type: 'number', 'data-bind': 'lineHeight', min: '0.8', max: '3', step: '0.05', value: layer.lineHeight || 1.25 }), layer.id, 'lineHeight')),
    ]));
    box.append(field('对齐', seg(layer, 'align', [['left', '左'], ['center', '中'], ['right', '右']])));
    box.append(h('div', { class: 'grid-2' }, [colorField('文字颜色', layer, 'color'), colorField('底色', layer, 'fill', true)]));
  }
  if (layer.type === 'image') {
    box.append(h('div', { class: 'field' }, [
      h('span', { text: '图片' }),
      layer.src ? h('img', { class: 'bg-preview', src: layer.src, alt: layer.name }) : null,
      h('button', { type: 'button', class: 'btn btn-small', text: '从图片库更换', onClick: () => setLibraryMode({ type: 'layer', layerId: layer.id }) }),
    ]));
    box.append(field('适配', seg(layer, 'fit', [['cover', '填充'], ['contain', '完整']])));
  }
  if (layer.type === 'rect' || layer.type === 'ellipse' || layer.type === 'line') {
    box.append(colorField(layer.type === 'line' ? '颜色' : '填充', layer, 'fill', true));
    if (layer.type !== 'line') box.append(colorField('描边', layer, 'stroke', true));
  }
  if (layer.type !== 'line' && layer.type !== 'ellipse') {
    box.append(field('圆角', bindNumber(h('input', { type: 'number', 'data-bind': 'radius', min: '0', max: '2048', value: layer.radius || 0 }), layer.id, 'radius')));
  }
  if (layer.type !== 'line' && layer.type !== 'text' && layer.type !== 'image') {
    box.append(field('描边粗细', bindNumber(h('input', { type: 'number', 'data-bind': 'strokeWidth', min: '0', max: '64', value: layer.strokeWidth || 0 }), layer.id, 'strokeWidth')));
  }
  box.append(h('div', { class: 'grid-2' }, [
    field('X', bindNumber(h('input', { type: 'number', 'data-bind': 'x', value: Math.round(layer.x) }), layer.id, 'x')),
    field('Y', bindNumber(h('input', { type: 'number', 'data-bind': 'y', value: Math.round(layer.y) }), layer.id, 'y')),
    field('宽', bindNumber(h('input', { type: 'number', 'data-bind': 'w', min: '1', value: Math.round(layer.w) }), layer.id, 'w')),
    field('高', bindNumber(h('input', { type: 'number', 'data-bind': 'h', min: '1', value: Math.round(layer.h) }), layer.id, 'h')),
  ]));
  box.append(field('旋转', bindNumber(h('input', { type: 'number', 'data-bind': 'rotation', step: '1', value: Math.round(layer.rotation || 0) }), layer.id, 'rotation')));
  const opacity = h('input', { type: 'range', 'data-bind': 'opacity', min: '0', max: '1', step: '0.01', value: layer.opacity ?? 1 });
  opacity.addEventListener('focus', armHistory);
  opacity.addEventListener('input', () => patch(() => { layer.opacity = Number(opacity.value); }));
  opacity.addEventListener('blur', disarmHistory);
  box.append(field('不透明度', h('div', { class: 'opacity-line' }, [opacity, h('span', { 'data-opacity-label': '1', text: Math.round((layer.opacity ?? 1) * 100) + '%' })])));
  box.append(field('对齐到画布', h('div', { class: 'seg wrap' }, [
    ['left', '左'], ['hcenter', '水平居中'], ['right', '右'], ['top', '上'], ['vcenter', '垂直居中'], ['bottom', '下'],
  ].map(([mode, label]) => h('button', { type: 'button', class: 'btn btn-small', text: label, onClick: () => alignLayer(layer.id, mode) })))));
  box.append(field('叠放', h('div', { class: 'seg wrap' }, [
    ['up', '上移一层'], ['down', '下移一层'], ['top', '置顶'], ['bottom', '置底'],
  ].map(([dir, label]) => h('button', { type: 'button', class: 'btn btn-small', text: label, onClick: () => moveLayer(layer.id, dir) })))));
  box.append(h('div', { class: 'inline-actions' }, [
    h('button', { type: 'button', class: 'btn btn-small', text: '复制图层', onClick: () => duplicateLayer(layer.id) }),
    h('button', { type: 'button', class: 'btn btn-danger btn-small', text: '删除', onClick: () => deleteLayer(layer.id) }),
  ]));
  return box;
}

function colorField(label, layer, key, allowEmpty) {
  const color = h('input', { type: 'color', value: toColorInput(layer[key]) });
  const text = h('input', { type: 'text', 'data-bind': key, value: layer[key] || '', placeholder: allowEmpty ? '留空为透明' : '#000000' });
  color.addEventListener('pointerdown', armHistory);
  color.addEventListener('input', () => {
    text.value = color.value;
    patch(() => { layer[key] = color.value; });
  });
  color.addEventListener('blur', disarmHistory);
  text.addEventListener('focus', armHistory);
  text.addEventListener('input', () => {
    if (!validColor(text.value, allowEmpty)) return;
    patch(() => { layer[key] = text.value; });
  });
  text.addEventListener('blur', () => {
    if (!validColor(text.value, allowEmpty)) text.value = layer[key] || '';
    disarmHistory();
  });
  return field(label, h('div', { class: 'color-line' }, [color, text]));
}

function seg(layer, key, options) {
  return h('div', { class: 'seg' }, options.map(([value, label]) => h('button', {
    type: 'button',
    class: (layer[key] || (key === 'fit' ? 'cover' : '')) === value ? 'is-active' : '',
    dataset: { choice: key + ':' + value },
    text: label,
    onClick: () => commit(() => { layer[key] = value; }),
  })));
}

function field(label, control) {
  return h('label', { class: 'field' }, [h('span', { text: label }), control]);
}

function bindText(input, id, key) {
  input.addEventListener('focus', armHistory);
  input.addEventListener('input', () => {
    const layer = state.doc.layers.find((item) => item.id === id);
    if (!layer) return;
    patch(() => { layer[key] = input.value; });
  });
  input.addEventListener('blur', disarmHistory);
  return input;
}

function bindNumber(input, id, key) {
  input.addEventListener('focus', armHistory);
  input.addEventListener('input', () => {
    if (input.value.trim() === '' || !Number.isFinite(Number(input.value))) return;
    const layer = state.doc.layers.find((item) => item.id === id);
    if (!layer) return;
    patch(() => { layer[key] = Number(input.value); });
  });
  input.addEventListener('blur', () => {
    const layer = state.doc.layers.find((item) => item.id === id);
    if (layer && (input.value.trim() === '' || !Number.isFinite(Number(input.value)))) input.value = layer[key];
    disarmHistory();
  });
  return input;
}

function resizeCanvas(width, height) {
  if (!Number.isFinite(width) || !Number.isFinite(height)) return;
  if (width === state.doc.width && height === state.doc.height) return;
  commit(() => scaleLayers(state.doc, width, height));
}

function alignLayer(id, mode) {
  const layer = state.doc.layers.find((item) => item.id === id);
  if (!layer) return;
  commit(() => {
    if (mode === 'left') layer.x = 0;
    if (mode === 'hcenter') layer.x = (state.doc.width - layer.w) / 2;
    if (mode === 'right') layer.x = state.doc.width - layer.w;
    if (mode === 'top') layer.y = 0;
    if (mode === 'vcenter') layer.y = (state.doc.height - layer.h) / 2;
    if (mode === 'bottom') layer.y = state.doc.height - layer.h;
  });
}

function moveLayer(id, dir) {
  const index = state.doc.layers.findIndex((item) => item.id === id);
  if (index < 0) return;
  if (dir === 'up' && index === state.doc.layers.length - 1) return;
  if (dir === 'down' && index === 0) return;
  commit(() => {
    const layers = state.doc.layers;
    const current = layers.findIndex((item) => item.id === id);
    if (dir === 'top') {
      const [item] = layers.splice(current, 1);
      layers.push(item);
      return;
    }
    if (dir === 'bottom') {
      const [item] = layers.splice(current, 1);
      layers.unshift(item);
      return;
    }
    const dest = dir === 'up' ? current + 1 : current - 1;
    [layers[current], layers[dest]] = [layers[dest], layers[current]];
  });
}

function duplicateLayer(id) {
  commit(() => {
    const index = state.doc.layers.findIndex((item) => item.id === id);
    if (index < 0) return;
    const copy = clone(state.doc.layers[index]);
    copy.id = uid('l');
    copy.name = [...((copy.name || '图层') + ' 副本')].slice(0, 40).join('');
    copy.x += 24;
    copy.y += 24;
    state.doc.layers.splice(index + 1, 0, copy);
    state.selectedId = copy.id;
  });
}

function deleteLayer(id) {
  const layer = state.doc.layers.find((item) => item.id === id);
  if (!layer) return;
  if (layer.locked) {
    toast('图层已锁定', 'error');
    return;
  }
  commit(() => {
    state.doc.layers = state.doc.layers.filter((item) => item.id !== id);
    if (state.selectedId === id) state.selectedId = null;
  });
}

function useAsset(item, point) {
  const mode = state.libraryMode;
  commit(() => {
    if (mode?.type === 'background') {
      state.doc.backgroundImage = item.src;
    } else if (mode?.type === 'layer') {
      const layer = state.doc.layers.find((entry) => entry.id === mode.layerId);
      if (layer) {
        layer.src = item.src;
        layer.name = item.name;
        state.selectedId = layer.id;
      }
    } else {
      state.selectedId = insertAsset(state.doc, item, point);
    }
    state.libraryMode = null;
  });
  if (mode?.type === 'background') toast('已设为背景');
  else if (mode?.type === 'layer') toast('已更换图片');
  else toast(item.placement === 'backdrop' ? '已放入背景图层' : '已添加到画布');
}

export function placeAssetSrc(src, point) {
  const item = state.library.find((entry) => entry.src === src);
  if (item) useAsset(item, point);
}

export async function placeUpload(file, point) {
  await uploadFile(file, point);
}

function pickFile() {
  const input = document.createElement('input');
  input.type = 'file';
  input.accept = 'image/png,image/jpeg,image/gif,image/webp,image/svg+xml,.svg';
  input.addEventListener('change', () => {
    if (input.files?.[0]) uploadFile(input.files[0]);
  });
  input.click();
}

async function uploadFile(file, point) {
  try {
    const item = await api.upload(file);
    state.library = (await api.library()).items;
    useAsset(item, point);
  } catch (err) {
    toast(err.message || '上传失败', 'error');
  }
}

async function removeUpload(item) {
  const ok = await confirmDialog({ title: '删除图片', body: '「' + item.name + '」会从图片库移除。', confirmText: '删除', danger: true });
  if (!ok) return;
  try {
    await api.deleteUpload(item.id);
    state.library = (await api.library()).items;
    renderLeft();
    toast('已删除图片');
  } catch (err) {
    toast(err.message || '删除失败', 'error');
  }
}

export async function openPosters() {
  state.posters = await api.listPosters();
  openModal('modal-md', (modal, close) => {
    modal.append(h('div', { class: 'modal-head' }, [
      h('h2', { text: '我的海报' }),
      h('button', { type: 'button', class: 'btn btn-small', text: '关闭', onClick: close }),
    ]));
    const list = h('div', { class: 'poster-list' });
    if (!state.posters.length) list.append(h('p', { text: '还没有海报' }));
    for (const poster of state.posters) {
      list.append(h('div', { class: 'poster-row' + (poster.id === state.doc.id ? ' is-current' : '') }, [
        h('div', {}, [
          h('strong', { text: poster.name }),
          h('small', { text: poster.width + '×' + poster.height + ' · ' + when(poster.updatedAt) }),
        ]),
        h('div', { class: 'inline-actions' }, [
          h('button', { type: 'button', class: 'btn btn-small', text: '打开', onClick: () => openOne(poster.id, close) }),
          h('button', { type: 'button', class: 'btn btn-small', text: '复制', onClick: () => copyOne(poster.id, close) }),
          h('button', { type: 'button', class: 'btn btn-small btn-danger', text: '删除', onClick: () => removeOne(poster, close) }),
        ]),
      ]));
    }
    modal.append(list);
  });
}

async function openOne(id, close) {
  if (id === state.doc.id) {
    close();
    return;
  }
  await saveNow();
  loadDoc(await api.getPoster(id));
  close();
}

async function copyOne(id, close) {
  await saveNow();
  const doc = await api.duplicatePoster(id);
  state.posters = await api.listPosters();
  loadDoc(doc);
  close();
  toast('已复制海报');
}

async function removeOne(poster, close) {
  const ok = await confirmDialog({ title: '删除海报', body: '「' + poster.name + '」删除后不能恢复。', confirmText: '删除', danger: true });
  if (!ok) return;
  await api.deletePoster(poster.id);
  const list = await api.listPosters();
  state.posters = list;
  close();
  if (poster.id === state.doc.id) {
    if (list.length) loadDoc(await api.getPoster(list[0].id));
    else loadDoc(await api.createPoster({}));
  }
  toast('已删除海报');
}

export function openModels() {
  openModal('modal-lg', (modal, close) => {
    let editingId = '';
    modal.append(h('div', { class: 'modal-head' }, [
      h('h2', { text: '模型配置' }),
      h('button', { type: 'button', class: 'btn btn-small', text: '关闭', onClick: close }),
    ]));
    modal.append(h('p', { class: 'panel-note', text: '使用 OpenAI Chat Completions 协议。可以接入 OpenAI、DeepSeek、通义兼容模式、Moonshot、Ollama、LM Studio 或任意兼容服务。密钥只保存在本机，由服务器转发请求。' }));
    const list = h('div', { class: 'profile-list' });
    const form = h('form', { class: 'llm-form' });
    const name = h('input', { required: 'true', maxlength: '40' });
    const preset = h('select');
    PRESETS.forEach(([value, label]) => preset.append(h('option', { value, text: label })));
    const base = h('input', { placeholder: 'https://api.openai.com/v1', required: 'true' });
    const key = h('input', { type: 'password', autocomplete: 'off', placeholder: 'sk-...' });
    const auth = h('select');
    [['bearer', 'Bearer Token（OpenAI 兼容）'], ['api-key', 'api-key 请求头（Azure）']].forEach(([value, label]) => auth.append(h('option', { value, text: label })));
    const model = h('input', { list: 'llm-models', placeholder: 'gpt-4o-mini', required: 'true' });
    const dataList = h('datalist', { id: 'llm-models' });
    const temp = h('input', { type: 'number', min: '0', max: '2', step: '0.1', value: '0.7' });
    const tokens = h('input', { type: 'number', min: '1', max: '8192', step: '1', value: '2048' });
    const tokenParam = h('select');
    [['max_tokens', 'max_tokens（大多数兼容服务）'], ['max_completion_tokens', 'max_completion_tokens（OpenAI 新模型）']].forEach(([value, label]) => tokenParam.append(h('option', { value, text: label })));
    const activate = h('input', { type: 'checkbox' });
    const status = h('p', { class: 'panel-note', text: '' });
    preset.addEventListener('change', () => {
      const found = PRESETS.find((item) => item[0] === preset.value);
      if (!found || found[0] === 'custom') return;
      base.value = found[2];
      if (found[3]) model.value = found[3];
    });
    form.append(
      field('配置名称', name),
      field('服务商预设', preset),
      field('接口根地址', base),
      field('密钥', key),
      field('认证方式', auth),
      field('模型', h('div', { class: 'inline-field' }, [model, h('button', { type: 'button', class: 'btn btn-small', text: '获取模型', onClick: () => fetchModels() })])),
      dataList,
      h('div', { class: 'grid-2' }, [field('温度', temp), field('最大输出', tokens)]),
      field('长度参数', tokenParam),
      h('label', { class: 'check-line' }, [activate, h('span', { text: '保存后设为当前模型' })]),
      status,
      h('div', { class: 'modal-actions' }, [
        h('button', { type: 'button', class: 'btn', text: '测试连接', onClick: () => testConnection() }),
        h('button', { type: 'submit', class: 'btn btn-primary', text: '保存' }),
        h('button', { type: 'button', class: 'btn btn-danger', text: '删除', id: 'llm-delete', onClick: () => removeProfile() }),
      ]),
    );
    form.addEventListener('submit', (event) => {
      event.preventDefault();
      saveProfile();
    });
    modal.append(h('div', { class: 'llm-layout' }, [list, form]));

    function payload() {
      return {
        id: editingId,
        name: name.value.trim(),
        baseUrl: base.value.trim(),
        model: model.value.trim(),
        apiKey: key.value,
        temperature: Number(temp.value),
        maxTokens: Number(tokens.value),
        authType: auth.value,
        tokenParam: tokenParam.value,
        activate: activate.checked,
      };
    }
    function fill(profile) {
      editingId = profile?.id || '';
      name.value = profile?.name || '';
      base.value = profile?.baseUrl || '';
      model.value = profile?.model || '';
      key.value = '';
      key.placeholder = profile?.apiKeySet ? '已保存，留空则不修改' : '本地模型可以留空';
      auth.value = profile?.authType || 'bearer';
      tokenParam.value = profile?.tokenParam || 'max_tokens';
      temp.value = profile ? profile.temperature : 0.7;
      tokens.value = profile ? profile.maxTokens : 2048;
      activate.checked = profile ? profile.id === state.activeProfileId : true;
      preset.value = 'custom';
      modal.querySelector('#llm-delete').hidden = !editingId;
      status.textContent = '';
    }
    function renderList() {
      clear(list);
      list.append(h('button', { type: 'button', class: 'btn btn-small', text: '新建配置', onClick: () => fill(null) }));
      if (!state.profiles.length) list.append(h('p', { class: 'panel-note', text: '还没有配置。先选一个服务商，填上密钥。' }));
      for (const profile of state.profiles) {
        const current = profile.id === state.activeProfileId;
        list.append(h('button', { type: 'button', class: 'profile-item' + (profile.id === editingId ? ' is-selected' : ''), onClick: () => fill(profile) }, [
          h('strong', { text: profile.name }),
          h('small', { text: profile.model + (current ? ' · 当前' : '') }),
        ]));
      }
    }
    async function refresh() {
      const data = await api.profiles();
      state.profiles = data.profiles;
      state.activeProfileId = data.activeProfileId;
    }
    async function saveProfile() {
      const body = payload();
      try {
        if (editingId) await api.updateProfile(editingId, body);
        else {
          const created = await api.createProfile(body);
          editingId = created.id;
        }
        await refresh();
        fill(state.profiles.find((item) => item.id === editingId) || null);
        renderList();
        toast('模型配置已保存');
      } catch (err) {
        status.textContent = err.message;
      }
    }
    async function testConnection() {
      status.textContent = '正在连接…';
      try {
        const res = await api.testProfile(payload());
        status.textContent = '连接成功：' + res.reply;
      } catch (err) {
        status.textContent = err.message;
      }
    }
    async function fetchModels() {
      status.textContent = '正在获取模型…';
      try {
        const res = await api.listModels(payload());
        clear(dataList);
        for (const id of res.models || []) dataList.append(h('option', { value: id }));
        if (res.models?.length && !model.value) model.value = res.models[0];
        status.textContent = res.models?.length ? '已获取 ' + res.models.length + ' 个模型' : '接口没有返回模型，请手动填写';
      } catch (err) {
        status.textContent = err.message;
      }
    }
    async function removeProfile() {
      if (!editingId) return;
      const ok = await confirmDialog({ title: '删除模型配置', body: '密钥会从本机配置里删除。', confirmText: '删除', danger: true });
      if (!ok) return;
      await api.deleteProfile(editingId);
      await refresh();
      fill(state.profiles.find((item) => item.id === state.activeProfileId) || null);
      renderList();
      toast('已删除配置');
    }
    const initial = state.profiles.find((item) => item.id === state.activeProfileId) || state.profiles[0] || null;
    fill(initial);
    renderList();
  });
}

export function openAI() {
  openModal('modal-md', (modal) => {
    const active = state.profiles.find((item) => item.id === state.activeProfileId);
    modal.append(h('div', { class: 'modal-head' }, [h('h2', { text: 'AI 生成' })]));
    modal.append(h('p', { class: 'panel-note', text: active ? '当前模型：' + active.name + ' / ' + active.model : '还没有启用模型。' }));
    if (!active) {
      modal.append(h('button', { type: 'button', class: 'btn btn-primary', text: '去配置模型', onClick: () => { closeTopModal(); openModels(); } }));
      return;
    }
    const textarea = h('textarea', { rows: '5', placeholder: '例如：一张咖啡新品海报，暖色，标题「秋日拿铁」，副标题「第二杯半价」' });
    textarea.value = aiPrompt;
    textarea.addEventListener('input', () => { aiPrompt = textarea.value; });
    const result = h('pre', { class: 'ai-result' });
    const copyBtn = h('button', { type: 'button', class: 'btn', text: '生成文案', onClick: () => run('copy') });
    const layoutBtn = h('button', { type: 'button', class: 'btn btn-primary', text: '生成排版', onClick: () => run('layout') });
    modal.append(textarea, h('div', { class: 'modal-actions' }, [copyBtn, layoutBtn]), result);
    async function run(mode) {
      const prompt = textarea.value.trim();
      if (!prompt) {
        toast('请先描述海报内容');
        return;
      }
      if (mode === 'layout' && state.doc.layers.length) {
        const ok = await confirmDialog({ title: '生成排版', body: '会替换当前图层，可以撤销。', confirmText: '生成并替换' });
        if (!ok) return;
      }
      copyBtn.disabled = true;
      layoutBtn.disabled = true;
      result.textContent = '正在请求模型…';
      try {
        const res = await api.generate({ mode, prompt, width: state.doc.width, height: state.doc.height, texts: currentTexts() });
        if (mode === 'copy') {
          commit(() => applyCopy(state.doc, res.copy));
          const lines = [res.copy.title, res.copy.subtitle, res.copy.body, ...(res.copy.bullets || [])].filter(Boolean);
          result.textContent = lines.join('\n');
          toast('文案已放到画布，可撤销');
        } else {
          commit(() => applyLayout(state.doc, res.layout));
          result.textContent = '已生成 ' + (res.layout.layers?.length || 0) + ' 个图层';
          toast('排版已应用到画布，可撤销');
        }
      } catch (err) {
        result.textContent = err.message;
        toast(err.message, 'error');
      } finally {
        copyBtn.disabled = false;
        layoutBtn.disabled = false;
      }
    }
  });
}

export function openHelp() {
  openModal('modal-sm', (modal, close) => {
    modal.append(h('div', { class: 'modal-head' }, [
      h('h2', { text: '快捷键' }),
      h('button', { type: 'button', class: 'btn btn-small', text: '关闭', onClick: close }),
    ]));
    modal.append(h('ul', { class: 'help-list' }, [
      ['Ctrl/⌘ Z', '撤销'],
      ['Ctrl/⌘ Shift Z', '重做'],
      ['Ctrl/⌘ S', '立即保存'],
      ['Ctrl/⌘ D', '复制图层'],
      ['Delete', '删除图层'],
      ['方向键', '微调 1 像素，按住 Shift 为 10 像素'],
      ['Ctrl/⌘ 滚轮', '缩放'],
      ['双击文字', '编辑文字'],
    ].map(([key, label]) => h('li', {}, [h('kbd', { text: key }), h('span', { text: label })]))));
  });
}

export function duplicateSelected() {
  if (state.selectedId) duplicateLayer(state.selectedId);
}

export function deleteSelected() {
  if (state.selectedId) deleteLayer(state.selectedId);
}

let nudging = false;

export function nudgeSelected(dx, dy) {
  const layer = selectedLayer();
  if (!layer || layer.locked) return;
  if (!nudging) {
    armHistory();
    nudging = true;
  }
  patch(() => {
    layer.x += dx;
    layer.y += dy;
  });
}

export function endNudge() {
  if (!nudging) return;
  nudging = false;
  disarmHistory();
}

function captureFocus() {
  const el = document.activeElement;
  if (!(el instanceof HTMLInputElement)) return null;
  if (el.id !== 'layer-filter' && el.id !== 'lib-filter') return null;
  return { id: el.id, start: el.selectionStart, end: el.selectionEnd };
}

function restoreFocus(saved) {
  if (!saved) return;
  const el = document.getElementById(saved.id);
  if (!el) return;
  el.focus();
  el.setSelectionRange(saved.start, saved.end);
}

function typeMark(layer) {
  if (layer.type === 'text') return '文';
  if (layer.type === 'image') return '图';
  if (layer.type === 'ellipse') return '圆';
  if (layer.type === 'line') return '线';
  return '形';
}

function toColorInput(value) {
  if (typeof value === 'string' && /^#[0-9a-fA-F]{6}$/.test(value)) return value;
  if (typeof value === 'string' && /^#[0-9a-fA-F]{3}$/.test(value)) {
    return '#' + [...value.slice(1)].map((ch) => ch + ch).join('');
  }
  return '#000000';
}

function validColor(value, allowEmpty) {
  if (allowEmpty && (value === '' || value === 'transparent')) return true;
  return /^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$/.test(value);
}

function roundIfNumber(value) {
  return typeof value === 'number' ? Math.round(value * 1000) / 1000 : value;
}

function when(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  return date.toLocaleString('zh-CN', { hour12: false, month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit' });
}
