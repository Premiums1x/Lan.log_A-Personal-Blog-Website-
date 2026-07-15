const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');

const CLIENT_SOURCE = fs.readFileSync(path.join(__dirname, '..', 'static', 'lancer.js'), 'utf8');

class FakeElement {
  constructor(text = '') {
    this._textContent = text;
    this.textWrites = 0;
    this.dataset = {};
    this.attributes = {};
    this.style = {
      values: new Map(),
      setProperty: (name, value) => this.style.values.set(name, value),
      removeProperty: (name) => this.style.values.delete(name),
    };
  }

  get textContent() { return this._textContent; }

  set textContent(value) {
    this._textContent = String(value);
    this.textWrites += 1;
  }

  setAttribute(name, value) { this.attributes[name] = String(value); }
}

class FakeEventTarget {
  constructor() { this.listeners = new Map(); }

  addEventListener(type, listener, options = {}) {
    if (!this.listeners.has(type)) this.listeners.set(type, new Set());
    this.listeners.get(type).add({ listener, once: Boolean(options && options.once) });
  }

  removeEventListener(type, listener) {
    const entries = this.listeners.get(type);
    if (!entries) return;
    for (const entry of entries) {
      if (entry.listener === listener) entries.delete(entry);
    }
  }

  dispatch(type, init = {}) {
    const event = { type, persisted: false, ...init };
    const entries = [...(this.listeners.get(type) || [])];
    for (const entry of entries) {
      entry.listener(event);
      if (entry.once) this.listeners.get(type).delete(entry);
    }
  }

  listenerCount(type) { return this.listeners.get(type)?.size || 0; }
}

class FakeTimers {
  constructor() {
    this.nextID = 1;
    this.intervals = new Map();
    this.timeouts = new Map();
  }

  setInterval(callback, delay) {
    const id = this.nextID++;
    this.intervals.set(id, { callback, delay });
    return id;
  }

  clearInterval(id) { this.intervals.delete(id); }

  setTimeout(callback, delay) {
    const id = this.nextID++;
    this.timeouts.set(id, { callback, delay });
    return id;
  }

  clearTimeout(id) { this.timeouts.delete(id); }

  runIntervals() {
    for (const { callback } of [...this.intervals.values()]) callback();
  }

  runTimeout(id) {
    const timer = this.timeouts.get(id);
    if (!timer) return;
    this.timeouts.delete(id);
    timer.callback();
  }
}

class ControlledFetch {
  constructor({ rejectOnAbort = true } = {}) {
    this.calls = [];
    this.rejectOnAbort = rejectOnAbort;
  }

  fetch = (url, options = {}) => new Promise((resolve, reject) => {
    const call = { url, options, resolve, reject };
    this.calls.push(call);
    if (this.rejectOnAbort && options.signal) {
      options.signal.addEventListener('abort', () => {
        const error = new Error('aborted');
        error.name = 'AbortError';
        reject(error);
      }, { once: true });
    }
  });
}

const snapshot = (overrides = {}) => ({
  online: true,
  cpu_percent: 24.5,
  cpu_frequency_mhz: 2450.4,
  memory_used_bytes: 1610612736,
  memory_total_bytes: 4294967296,
  memory_percent: 37.5,
  uptime_seconds: 93784,
  ...overrides,
});

const ok = (body) => ({ ok: true, json: async () => body });
const flush = async () => {
  await Promise.resolve();
  await Promise.resolve();
  await new Promise((resolve) => setImmediate(resolve));
};

function createRuntime(options = {}) {
  const timers = new FakeTimers();
  const requests = new ControlledFetch(options);
  const document = new FakeEventTarget();
  const window = new FakeEventTarget();
  document.visibilityState = options.visibilityState || 'visible';

  const elements = {
    '[data-status]': new FakeElement('CONNECTING'),
    '[data-metric="cpu"]': new FakeElement('--%'),
    '[data-metric="frequency"]': new FakeElement('-- MHz'),
    '[data-metric="memory"]': new FakeElement('--%'),
    '[data-memory-detail]': new FakeElement('-- / --'),
    '[data-metric="uptime"]': new FakeElement('--'),
    '[data-chart-samples="cpu"]': new FakeElement('00 / 40'),
    '[data-chart-samples="frequency"]': new FakeElement('00 / 40'),
    '[data-chart-samples="memory"]': new FakeElement('00 / 40'),
    '[data-chart-samples="uptime"]': new FakeElement('00 / 40'),
  };
  const chartParts = {};
  const charts = ['cpu', 'frequency', 'memory', 'uptime'].map((key) => {
    const line = new FakeElement();
    const area = new FakeElement();
    const point = new FakeElement();
    const chart = new FakeElement();
    chart.dataset.chart = key;
    chart.querySelector = (selector) => ({
      '[data-chart-line]': line,
      '[data-chart-area]': area,
      '[data-chart-point]': point,
    })[selector];
    chartParts[key] = { line, area, point };
    return chart;
  });
  const root = new FakeElement();
  root.querySelector = (selector) => elements[selector];
  root.querySelectorAll = (selector) => selector === '[data-chart]' ? charts : [];

  document.querySelectorAll = (selector) => selector === '[data-system-status]' ? [root] : [];
  window.setInterval = timers.setInterval.bind(timers);
  window.addEventListener = window.addEventListener.bind(window);

  const context = {
    document,
    window,
    fetch: requests.fetch,
    AbortController,
    Intl,
    Math,
    Number,
    Error,
    Promise,
    console,
    setTimeout: timers.setTimeout.bind(timers),
    clearTimeout: timers.clearTimeout.bind(timers),
    clearInterval: timers.clearInterval.bind(timers),
  };
  vm.runInNewContext(CLIENT_SOURCE, context, { filename: 'lancer.js' });
  return { timers, requests, document, window, elements, root, chartParts };
}

function createInfluenceRuntime() {
  const element = (text = '') => ({
    textContent: text,
    dataset: {},
    attributes: {},
    listeners: new Map(),
    classNames: new Set(),
    style: {
      values: new Map(),
      setProperty(name, value) { this.values.set(name, value); },
    },
    classList: {
      add(name) { this.owner.classNames.add(name); },
      remove(name) { this.owner.classNames.delete(name); },
      owner: null,
    },
    addEventListener(type, listener) { this.listeners.set(type, listener); },
    dispatch(type) { this.listeners.get(type)?.({ type }); },
    getAttribute(name) { return this.attributes[name]; },
    setAttribute(name, value) { this.attributes[name] = String(value); },
    querySelector(selector) { return this.children?.[selector] || null; },
  });

  const cards = [
    ['f1', '01 / PADDOCK', 'F1', 'precision', '/f1.webp'],
    ['sport', '02 / PITCH', 'FOOTBALL', 'vision', '/football.webp'],
    ['server', '03 / SERVER', 'SERVER', 'entry', '/server.webp'],
  ].map(([influence, index, title, copy, image], cardIndex) => {
    const card = element();
    card.dataset = { influence, index, image, position: 'center' };
    card.attributes['aria-pressed'] = String(cardIndex === 0);
    card.children = { h3: element(title), p: element(copy) };
    card.classList.owner = card;
    return card;
  });
  const layers = [element(), element()];
  layers.forEach((layer, index) => {
    layer.classList.owner = layer;
    if (index === 0) layer.classNames.add('is-active');
  });
  const index = element('01 / PADDOCK');
  const title = element('F1');
  const copy = element('precision');
  const preview = element();
  const root = element();
  root.dataset.active = 'f1';
  root.querySelectorAll = (selector) => selector === '[data-influence]' ? cards : layers;
  root.querySelector = (selector) => ({
    '.influence-preview': preview,
    '#influence-preview-index': index,
    '#influence-preview-title': title,
    '#influence-preview-copy': copy,
  })[selector] || null;

  const document = {
    querySelectorAll(selector) {
      return selector === '.interactive-influences' ? [root] : [];
    },
  };
  vm.runInNewContext(CLIENT_SOURCE, { document }, { filename: 'lancer.js' });
  return { cards, layers, root, index, title, copy };
}

function createDrawerRuntime() {
  const listeners = new Map();
  const classNames = new Set();
  const captured = new Set();
  const scrollCalls = [];
  const openCalls = [];
  const frames = [];
  const drawer = {
    classList: {
      add(name) { classNames.add(name); },
      remove(name) { classNames.delete(name); },
    },
    scrollIntoView(options) { openCalls.push(options); },
  };
  const handle = {
    closest(selector) { return selector === '.page-drawer' ? drawer : null; },
    addEventListener(type, listener) { listeners.set(type, listener); },
    setPointerCapture(id) { captured.add(id); },
    hasPointerCapture(id) { return captured.has(id); },
    releasePointerCapture(id) { captured.delete(id); },
    dispatch(type, init = {}) {
      const event = {
        button: 0,
        pointerId: 1,
        clientY: 0,
        defaultPrevented: false,
        preventDefault() { this.defaultPrevented = true; },
        ...init,
      };
      listeners.get(type)?.(event);
      return event;
    },
  };
  const document = {
    documentElement: { style: { scrollBehavior: '' } },
    querySelectorAll(selector) { return selector === '[data-drawer-handle]' ? [handle] : []; },
  };
  const window = {
    scrollY: 200,
    scrollTo(options) { scrollCalls.push(options); this.scrollY = options.top; },
    matchMedia() { return { matches: false }; },
  };
  const requestAnimationFrame = (callback) => {
    frames.push(callback);
    return frames.length;
  };
  vm.runInNewContext(CLIENT_SOURCE, { document, window, requestAnimationFrame, Math }, { filename: 'lancer.js' });
  return {
    handle,
    listeners,
    classNames,
    scrollCalls,
    openCalls,
    documentElement: document.documentElement,
    flushFrame() { frames.shift()?.(); },
  };
}

function createArticleExpandRuntime(itemCount = 8, initialCount = 6) {
  const items = Array.from({ length: itemCount }, () => ({ hidden: false, style: {}, classList: { add() {} } }));
  const trigger = new FakeEventTarget();
  trigger.hidden = true;
  trigger.attributes = {};
  trigger.setAttribute = (name, value) => { trigger.attributes[name] = String(value); };
  const root = {
    dataset: { expandInitial: String(initialCount) },
    querySelectorAll(selector) { return selector === '[data-expand-item]' ? items : []; },
    querySelector(selector) { return selector === '[data-expand-trigger]' ? trigger : null; },
  };
  const document = {
    querySelectorAll(selector) { return selector === '[data-expand-list]' ? [root] : []; },
  };
  vm.runInNewContext(CLIENT_SOURCE, {
    document,
    window: {},
    requestAnimationFrame(callback) { callback(); return 1; },
  }, { filename: 'lancer.js' });
  return { items, trigger, root };
}
test('article lists show six items then reveal every remaining item', () => {
  const runtime = createArticleExpandRuntime(9, 6);
  assert.deepEqual(runtime.items.map((item) => item.hidden), [false, false, false, false, false, false, true, true, true]);
  assert.equal(runtime.trigger.hidden, false);
  assert.equal(runtime.trigger.attributes['aria-expanded'], 'false');

  runtime.trigger.dispatch('click');
  assert.equal(runtime.items.every((item) => item.hidden === false), true);
  assert.equal(runtime.trigger.hidden, true);
  assert.equal(runtime.trigger.attributes['aria-expanded'], 'true');
});
test('drawer handle drags native scroll and keeps click-to-open behavior', () => {
  const runtime = createDrawerRuntime();
  for (const eventType of ['pointerdown', 'pointermove', 'pointerup', 'pointercancel', 'click']) {
    assert.equal(typeof runtime.listeners.get(eventType), 'function');
  }

  runtime.handle.dispatch('pointerdown', { pointerId: 7, clientY: 500 });
  assert.equal(runtime.classNames.has('is-dragging'), true);
  assert.equal(runtime.documentElement?.style?.scrollBehavior ?? 'auto', 'auto');
  const move = runtime.handle.dispatch('pointermove', { pointerId: 7, clientY: 400 });
  runtime.flushFrame();
  assert.equal(move.defaultPrevented, true);
  assert.equal(runtime.scrollCalls.at(-1).top, 300);

  runtime.handle.dispatch('pointerup', { pointerId: 7, clientY: 400 });
  assert.equal(runtime.classNames.has('is-dragging'), false);
  assert.equal(runtime.documentElement.style.scrollBehavior, '');
  const suppressedClick = runtime.handle.dispatch('click');
  assert.equal(suppressedClick.defaultPrevented, true);
  assert.equal(runtime.openCalls.length, 0);

  runtime.handle.dispatch('pointerdown', { pointerId: 8, clientY: 500 });
  runtime.handle.dispatch('pointerup', { pointerId: 8, clientY: 500 });
  runtime.handle.dispatch('click');
  assert.equal(runtime.openCalls[0].behavior, 'smooth');
  assert.equal(runtime.openCalls[0].block, 'start');
});
test('influence buttons preserve pointer, focus, click, pressed state, and crossfade', () => {
  const runtime = createInfluenceRuntime();
  assert.equal(runtime.layers[0].style.values.get('--influence-image'), 'url("/f1.webp")');
  for (const eventType of ['pointerenter', 'focusin', 'click']) {
    assert.equal(typeof runtime.cards[1].listeners.get(eventType), 'function');
  }

  runtime.cards[1].dispatch('focusin');
  assert.equal(runtime.cards[0].attributes['aria-pressed'], 'false');
  assert.equal(runtime.cards[1].attributes['aria-pressed'], 'true');
  assert.equal(runtime.root.dataset.active, 'sport');
  assert.equal(runtime.title.textContent, 'FOOTBALL');
  assert.equal(runtime.layers[0].classNames.has('is-active'), false);
  assert.equal(runtime.layers[1].classNames.has('is-active'), true);

  runtime.cards[2].dispatch('pointerenter');
  assert.equal(runtime.cards[2].attributes['aria-pressed'], 'true');
  assert.equal(runtime.layers[0].classNames.has('is-active'), true);
  assert.equal(runtime.layers[1].classNames.has('is-active'), false);

  runtime.cards[0].dispatch('click');
  assert.equal(runtime.cards[0].attributes['aria-pressed'], 'true');
  assert.equal(runtime.layers[1].classNames.has('is-active'), true);
});

test('starts immediately with one interval and never duplicates visible polling', () => {
  const runtime = createRuntime();
  assert.equal(runtime.requests.calls.length, 1);
  assert.equal(runtime.timers.intervals.size, 1);

  runtime.document.dispatch('visibilitychange');
  runtime.document.dispatch('visibilitychange');
  assert.equal(runtime.requests.calls.length, 1);
  assert.equal(runtime.timers.intervals.size, 1);
});

test('telemetry charts keep stable honest scales instead of amplifying tiny changes', async () => {
  const runtime = createRuntime();
  runtime.requests.calls[0].resolve(ok(snapshot({
    cpu_frequency_mhz: 3000,
    uptime_seconds: 93784,
  })));
  await flush();

  runtime.timers.runIntervals();
  runtime.requests.calls[1].resolve(ok(snapshot({
    cpu_frequency_mhz: 3500,
    uptime_seconds: 93799,
  })));
  await flush();

  const uptimePath = runtime.chartParts.uptime.line.attributes.d;
  const uptimeY = [...uptimePath.matchAll(/[ML][\d.]+ ([\d.]+)/g)].map((match) => Number(match[1]));
  assert.deepEqual(uptimeY, [46, 46], 'uptime should advance horizontally without fake vertical volatility');
  assert.equal(runtime.chartParts.uptime.area.attributes.d, '', 'uptime trace should not imply a measured area');

  const frequencyPath = runtime.chartParts.frequency.line.attributes.d;
  const frequencyY = [...frequencyPath.matchAll(/[ML][\d.]+ ([\d.]+)/g)].map((match) => Number(match[1]));
  assert.equal(frequencyY.length, 2);
  assert.ok(frequencyY.every((value) => value > 6 && value < 86), 'ordinary clock changes should stay inside the fixed 0-5 GHz scale');
});
test('hidden state aborts work, stops polling, and settles aria-busy', async () => {
  const runtime = createRuntime();
  const first = runtime.requests.calls[0];
  assert.equal(runtime.root.attributes['aria-busy'], 'true');

  runtime.document.visibilityState = 'hidden';
  runtime.document.dispatch('visibilitychange');
  await flush();

  assert.equal(first.options.signal.aborted, true);
  assert.equal(runtime.timers.intervals.size, 0);
  assert.equal(runtime.timers.timeouts.size, 0);
  assert.equal(runtime.root.attributes['aria-busy'], 'false');
});

test('clamps valid percentages and invalid or stale responses cannot leave fake values', async () => {
  const runtime = createRuntime({ rejectOnAbort: false });
  const stale = runtime.requests.calls[0];

  runtime.timers.runIntervals();
  const current = runtime.requests.calls[1];
  current.resolve(ok(snapshot({ cpu_percent: 125, memory_percent: 140 })));
  await flush();
  assert.equal(runtime.elements['[data-metric="cpu"]'].textContent, '100%');
  assert.match(runtime.chartParts.cpu.line.attributes.d, /^M/);
  assert.equal(runtime.elements['[data-chart-samples="cpu"]'].textContent, '01 / 40');
  assert.equal(runtime.elements['[data-metric="memory"]'].textContent, '100%');

  stale.resolve(ok(snapshot({ cpu_percent: 2 })));
  await flush();
  assert.equal(runtime.elements['[data-metric="cpu"]'].textContent, '100%', 'stale response must not apply');

  runtime.timers.runIntervals();
  runtime.requests.calls[2].resolve(ok({ online: true, cpu_percent: Number.NaN }));
  await flush();
  assert.equal(runtime.elements['[data-status]'].textContent, 'UNAVAILABLE');
  assert.equal(runtime.elements['[data-metric="cpu"]'].textContent, '--%');
  assert.equal(runtime.elements['[data-memory-detail]'].textContent, '-- / --');
  assert.match(runtime.chartParts.cpu.line.attributes.d, /^M/);
});

test('an old request finalizer cannot clear the replacement request timeout', async () => {
  const runtime = createRuntime({ rejectOnAbort: false });
  const oldRequest = runtime.requests.calls[0];
  const oldTimeout = [...runtime.timers.timeouts.keys()][0];

  runtime.document.visibilityState = 'hidden';
  runtime.document.dispatch('visibilitychange');
  assert.equal(runtime.timers.timeouts.has(oldTimeout), false);

  runtime.document.visibilityState = 'visible';
  runtime.document.dispatch('visibilitychange');
  const replacementTimeout = [...runtime.timers.timeouts.keys()][0];
  assert.ok(replacementTimeout);

  oldRequest.reject(new Error('late abort rejection'));
  await flush();
  assert.equal(runtime.timers.timeouts.has(replacementTimeout), true);
});

test('persisted pagehide suspends and pageshow restarts once without destroying listeners', async () => {
  const runtime = createRuntime();
  const first = runtime.requests.calls[0];

  runtime.window.dispatch('pagehide', { persisted: true });
  await flush();
  assert.equal(first.options.signal.aborted, true);
  assert.equal(runtime.timers.intervals.size, 0);
  assert.equal(runtime.document.listenerCount('visibilitychange'), 1);
  assert.equal(runtime.window.listenerCount('pageshow'), 1);
  assert.equal(runtime.root.attributes['aria-busy'], 'false');

  runtime.window.dispatch('pageshow', { persisted: true });
  assert.equal(runtime.requests.calls.length, 2);
  assert.equal(runtime.timers.intervals.size, 1);

  runtime.window.dispatch('pageshow', { persisted: true });
  assert.equal(runtime.requests.calls.length, 2);
  assert.equal(runtime.timers.intervals.size, 1);
});

test('non-persisted pagehide permanently removes lifecycle listeners', () => {
  const runtime = createRuntime();
  runtime.window.dispatch('pagehide', { persisted: false });
  assert.equal(runtime.document.listenerCount('visibilitychange'), 0);
  assert.equal(runtime.window.listenerCount('pageshow'), 0);
  assert.equal(runtime.timers.intervals.size, 0);
});

test('same status refresh updates metrics without re-announcing the live region', async () => {
  const runtime = createRuntime();
  runtime.requests.calls[0].resolve(ok(snapshot()));
  await flush();
  const status = runtime.elements['[data-status]'];
  assert.equal(status.textContent, 'ONLINE');
  assert.equal(status.textWrites, 1);

  runtime.timers.runIntervals();
  runtime.requests.calls[1].resolve(ok(snapshot({ cpu_percent: 42 })));
  await flush();
  assert.equal(runtime.elements['[data-metric="cpu"]'].textContent, '42%');
  assert.equal(status.textContent, 'ONLINE');
  assert.equal(status.textWrites, 1, 'ONLINE must not be announced twice');
});