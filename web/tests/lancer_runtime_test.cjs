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
    '[data-gauge="cpu"]': new FakeElement(),
    '[data-gauge="memory"]': new FakeElement(),
  };
  const root = new FakeElement();
  root.querySelector = (selector) => elements[selector];

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
  return { timers, requests, document, window, elements, root };
}

test('starts immediately with one interval and never duplicates visible polling', () => {
  const runtime = createRuntime();
  assert.equal(runtime.requests.calls.length, 1);
  assert.equal(runtime.timers.intervals.size, 1);

  runtime.document.dispatch('visibilitychange');
  runtime.document.dispatch('visibilitychange');
  assert.equal(runtime.requests.calls.length, 1);
  assert.equal(runtime.timers.intervals.size, 1);
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
  assert.equal(runtime.elements['[data-gauge="cpu"]'].style.values.get('--gauge'), '100%');
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
  assert.equal(runtime.elements['[data-gauge="cpu"]'].style.values.has('--gauge'), false);
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