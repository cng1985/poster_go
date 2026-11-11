import { uid } from './id.js';

function base(partial) {
  return {
    id: uid('l'),
    rotation: 0,
    opacity: 1,
    visible: true,
    locked: false,
    ...partial,
  };
}

function centered(doc, w, h, yRatio) {
  const offset = (doc.layers.length % 5) * 16;
  return {
    x: Math.round((doc.width - w) / 2 + offset),
    y: Math.round(doc.height * yRatio + offset),
  };
}

export function insertAsset(doc, item, point) {
  if (item.placement === 'backdrop') {
    const layer = base({
      name: item.name,
      type: 'image',
      role: 'backdrop',
      x: 0,
      y: 0,
      w: doc.width,
      h: doc.height,
      src: item.src,
      fit: 'cover',
      radius: 0,
    });
    const first = doc.layers[0];
    if (first && first.type === 'image' && first.role === 'backdrop') {
      doc.layers[0] = { ...layer, id: first.id };
      return first.id;
    }
    doc.layers.unshift(layer);
    return layer.id;
  }
  const size = item.category === 'icon' ? 180 : Math.round(Math.min(doc.width, doc.height) * 0.42);
  const x = point ? point.x - size / 2 : (doc.width - size) / 2;
  const y = point ? point.y - size / 2 : (doc.height - size) / 2;
  const layer = base({
    name: item.name,
    type: 'image',
    role: 'image',
    x: Math.round(x),
    y: Math.round(y),
    w: size,
    h: size,
    src: item.src,
    fit: 'cover',
    radius: 0,
  });
  doc.layers.push(layer);
  return layer.id;
}

export const COMPONENT_GROUPS = [
  {
    title: '文本',
    items: [
      {
        id: 'title',
        name: '主标题',
        hint: '大号标题',
        preview: 'title',
        build(doc) {
          const w = Math.round(doc.width * 0.8);
          const pos = centered(doc, w, 140, 0.16);
          return [base({ name: '主标题', type: 'text', role: 'title', text: '主标题', w, h: 140, ...pos, fontSize: Math.max(48, Math.round(doc.width * 0.072)), fontWeight: '800', fontFamily: 'sans', color: '#1c1915', align: 'center', lineHeight: 1.1 })];
        },
      },
      {
        id: 'subtitle',
        name: '副标题',
        hint: '补充一句',
        preview: 'sub',
        build(doc) {
          const w = Math.round(doc.width * 0.72);
          const pos = centered(doc, w, 72, 0.32);
          return [base({ name: '副标题', type: 'text', role: 'subtitle', text: '写一句副标题', w, h: 72, ...pos, fontSize: 32, fontWeight: '500', fontFamily: 'sans', color: '#5c564e', align: 'center', lineHeight: 1.3 })];
        },
      },
      {
        id: 'body',
        name: '正文',
        hint: '多行说明',
        preview: 'body',
        build(doc) {
          const w = Math.round(doc.width * 0.72);
          const pos = centered(doc, w, 160, 0.42);
          return [base({ name: '正文', type: 'text', role: 'body', text: '在这里写下正文。\n短句更适合海报。', w, h: 160, ...pos, fontSize: 26, fontWeight: '400', fontFamily: 'sans', color: '#2a2722', align: 'left', lineHeight: 1.45 })];
        },
      },
      {
        id: 'price',
        name: '价格',
        hint: '圆角色块',
        preview: 'badge',
        build(doc) {
          const pos = centered(doc, 220, 72, 0.62);
          return [base({ name: '价格', type: 'text', role: 'price', text: '¥ 99', w: 220, h: 72, ...pos, fontSize: 36, fontWeight: '800', fontFamily: 'sans', color: '#ffffff', align: 'center', fill: '#e23d2b', radius: 999, lineHeight: 1 })];
        },
      },
    ],
  },
  {
    title: '形状',
    items: [
      {
        id: 'rect',
        name: '矩形',
        hint: '色块',
        preview: 'rect',
        build(doc) {
          const pos = centered(doc, 420, 200, 0.36);
          return [base({ name: '矩形', type: 'rect', role: 'decoration', w: 420, h: 200, ...pos, fill: '#1c1915', radius: 0 })];
        },
      },
      {
        id: 'card',
        name: '圆角卡片',
        hint: '白底卡片',
        preview: 'card',
        build(doc) {
          const pos = centered(doc, 460, 260, 0.32);
          return [base({ name: '圆角卡片', type: 'rect', role: 'decoration', w: 460, h: 260, ...pos, fill: '#ffffff', stroke: '#e6dfd4', strokeWidth: 2, radius: 28 })];
        },
      },
      {
        id: 'ellipse',
        name: '圆形',
        hint: '圆与椭圆',
        preview: 'ellipse',
        build(doc) {
          const pos = centered(doc, 220, 220, 0.36);
          return [base({ name: '圆形', type: 'ellipse', role: 'decoration', w: 220, h: 220, ...pos, fill: '#e23d2b' })];
        },
      },
      {
        id: 'line',
        name: '直线',
        hint: '分隔',
        preview: 'line',
        build(doc) {
          const pos = centered(doc, 420, 8, 0.48);
          return [base({ name: '直线', type: 'line', role: 'decoration', w: 420, h: 8, ...pos, fill: '#1c1915', radius: 999 })];
        },
      },
    ],
  },
  {
    title: '图片',
    items: [
      { id: 'pick-image', name: '从图片库', hint: '选择一张图片', preview: 'image', action: 'library' },
      { id: 'upload-image', name: '上传图片', hint: 'PNG、JPG、WEBP、SVG', preview: 'upload', action: 'upload' },
    ],
  },
  {
    title: '装饰',
    items: [
      {
        id: 'badge',
        name: '徽章',
        hint: '标签和文字',
        preview: 'badge',
        build(doc) {
          const pos = centered(doc, 200, 64, 0.12);
          return [
            base({ name: '徽章底', type: 'rect', role: 'decoration', w: 200, h: 64, ...pos, fill: '#1f6b4a', radius: 999 }),
            base({ name: '徽章字', type: 'text', role: 'kicker', text: '新品', w: 200, h: 64, ...pos, fontSize: 26, fontWeight: '700', fontFamily: 'sans', color: '#ffffff', align: 'center', lineHeight: 1 }),
          ];
        },
      },
      {
        id: 'quote',
        name: '引用',
        hint: '色条加文字',
        preview: 'quote',
        build(doc) {
          const pos = centered(doc, 520, 120, 0.4);
          return [
            base({ name: '引用条', type: 'rect', role: 'decoration', x: pos.x, y: pos.y, w: 10, h: 120, fill: '#e23d2b', radius: 8 }),
            base({ name: '引用', type: 'text', role: 'body', text: '引用一句让人停住的话', x: pos.x + 28, y: pos.y, w: 492, h: 120, fontSize: 30, fontWeight: '600', fontFamily: 'serif', color: '#1c1915', align: 'left', lineHeight: 1.35 }),
          ];
        },
      },
    ],
  },
];
