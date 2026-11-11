import assert from 'node:assert/strict';
import test from 'node:test';
import { localToParent, reorderLayers, resizeLayer, scaleLayers, snapMove } from './geom.js';

test('reorder moves the visual list and keeps paint order', () => {
  assert.deepEqual(reorderLayers(['A', 'B', 'C'], 0, 2), ['C', 'A', 'B']);
  assert.deepEqual(reorderLayers(['A', 'B', 'C'], 2, 0), ['B', 'C', 'A']);
  assert.deepEqual(reorderLayers(['A', 'B', 'C'], 0, 1), ['A', 'C', 'B']);
  assert.deepEqual(reorderLayers(['A', 'B', 'C'], 1, 2), ['B', 'A', 'C']);
  assert.deepEqual(reorderLayers(['A', 'B', 'C'], 1, 1), ['A', 'B', 'C']);
});

test('resize keeps the opposite edge fixed', () => {
  const handles = ['n', 's', 'e', 'w', 'ne', 'nw', 'se', 'sw'];
  for (const rotation of [0, 30, 90]) {
    for (const handle of handles) {
      const start = { x: 40, y: 60, w: 180, h: 90, rotation };
      const next = resizeLayer(start, handle, 36, -18);
      if (next.w <= 8.01 || next.h <= 8.01) continue;
      const anchor = {
        e: [0, start.h / 2],
        w: [start.w, start.h / 2],
        n: [start.w / 2, start.h],
        s: [start.w / 2, 0],
        ne: [0, start.h],
        nw: [start.w, start.h],
        se: [0, 0],
        sw: [start.w, 0],
      }[handle];
      const before = localToParent(start.x, start.y, start.w, start.h, rotation, anchor[0], anchor[1]);
      const moved = {
        e: [0, next.h / 2],
        w: [next.w, next.h / 2],
        n: [next.w / 2, next.h],
        s: [next.w / 2, 0],
        ne: [0, next.h],
        nw: [next.w, next.h],
        se: [0, 0],
        sw: [next.w, 0],
      }[handle];
      const after = localToParent(next.x, next.y, next.w, next.h, rotation, moved[0], moved[1]);
      assert.ok(Math.abs(before.x - after.x) < 1e-6, handle + ' x ' + rotation);
      assert.ok(Math.abs(before.y - after.y) < 1e-6, handle + ' y ' + rotation);
    }
  }
});

test('east resize without rotation grows width in place', () => {
  const next = resizeLayer({ x: 10, y: 20, w: 100, h: 40, rotation: 0 }, 'e', 25, 0);
  assert.equal(next.x, 10);
  assert.equal(next.y, 20);
  assert.equal(next.w, 125);
  assert.equal(next.h, 40);
});

test('snap centers and edges', () => {
  const centered = snapMove({ width: 100, height: 80 }, 28, 10, 40, 20);
  assert.equal(centered.x, 30);
  assert.equal(centered.guides.v, true);
  const edge = snapMove({ width: 100, height: 80 }, 2, 64, 40, 20);
  assert.equal(edge.x, 0);
  assert.equal(edge.y, 60);
});

test('scaleLayers scales geometry with the canvas', () => {
  const doc = {
    width: 500,
    height: 800,
    layers: [{ x: 50, y: 80, w: 100, h: 200, fontSize: 20, radius: 8, strokeWidth: 2 }],
  };
  scaleLayers(doc, 1000, 1600);
  assert.equal(doc.width, 1000);
  assert.equal(doc.height, 1600);
  assert.deepEqual(doc.layers[0], { x: 100, y: 160, w: 200, h: 400, fontSize: 40, radius: 16, strokeWidth: 4 });
});
