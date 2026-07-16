import assert from 'node:assert/strict';
import { readFileSync, existsSync } from 'node:fs';
import test from 'node:test';

const htmlPath = new URL('./apex-clutch.html', import.meta.url);
const assets = [
  new URL('./apex-clutch-assets/f1-verstappen.png', import.meta.url),
  new URL('./apex-clutch-assets/basketball-curry.png', import.meta.url),
  new URL('./apex-clutch-assets/cs-niko.png', import.meta.url),
];

test('apex-clutch prototype has its required scenes and motion fallback', () => {
  const html = readFileSync(htmlPath, 'utf8');
  for (const asset of assets) assert.equal(existsSync(asset), true);
  assert.match(html, /<main[\s>]/);
  assert.match(html, /f1-verstappen\.png/);
  assert.match(html, /basketball-curry\.png/);
  assert.match(html, /cs-niko\.png/);
  assert.match(html, /@media\s*\(prefers-reduced-motion:\s*reduce\)/);
  assert.match(html, /aria-label=/);
  assert.match(html, /data-scene="f1"/);
  assert.match(html, /data-scene="basketball"/);
  assert.match(html, /data-scene="cs"/);
  assert.match(html, /class="post-link"/);
  assert.match(html, /IntersectionObserver/);
});
