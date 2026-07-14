document.querySelectorAll('[data-drawer-target]').forEach((cue) => {
  cue.addEventListener('click', () => cue.classList.add('is-used'), { once: true });
});

document.querySelectorAll('.interactive-influences').forEach((root) => {
  const cards = [...root.querySelectorAll('[data-influence]')];
  const preview = root.querySelector('.influence-preview');
  const layers = [...root.querySelectorAll('.influence-preview-image')];
  const index = root.querySelector('#influence-preview-index');
  const title = root.querySelector('#influence-preview-title');
  const copy = root.querySelector('#influence-preview-copy');
  let activeLayer = 0;

  const selected = cards.find((card) => card.getAttribute('aria-selected') === 'true');
  if (selected && layers[0]) {
    layers[0].style.setProperty('--influence-image', `url("${selected.dataset.image}")`);
    layers[0].style.backgroundPosition = selected.dataset.position || 'center';
  }

  const activate = (card) => {
    if (root.dataset.active === card.dataset.influence) return;

    cards.forEach((item) => item.setAttribute('aria-selected', String(item === card)));
    root.dataset.active = card.dataset.influence;
    index.textContent = card.dataset.index;
    title.textContent = card.querySelector('h3').textContent;
    copy.textContent = card.querySelector('p').textContent;

    if (preview && layers.length === 2 && card.dataset.image) {
      const nextLayer = activeLayer === 0 ? 1 : 0;
      layers[nextLayer].style.setProperty('--influence-image', `url("${card.dataset.image}")`);
      layers[nextLayer].style.backgroundPosition = card.dataset.position || 'center';
      layers[nextLayer].classList.add('is-active');
      layers[activeLayer].classList.remove('is-active');
      activeLayer = nextLayer;
    }
  };

  cards.forEach((card) => {
    card.addEventListener('pointerenter', () => activate(card));
    card.addEventListener('focusin', () => activate(card));
    card.addEventListener('click', () => activate(card));
  });
});
// Home-only telemetry client. Aggregate values are rendered without exposing host identity.
document.querySelectorAll('[data-system-status]').forEach((root) => {
  const status = root.querySelector('[data-status]');
  const cpu = root.querySelector('[data-metric="cpu"]');
  const frequency = root.querySelector('[data-metric="frequency"]');
  const memory = root.querySelector('[data-metric="memory"]');
  const memoryDetail = root.querySelector('[data-memory-detail]');
  const uptime = root.querySelector('[data-metric="uptime"]');
  const cpuGauge = root.querySelector('[data-gauge="cpu"]');
  const memoryGauge = root.querySelector('[data-gauge="memory"]');
  const percentFormat = new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 1 });
  const numberFormat = new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 0 });

  let pollTimer = null;
  let activeController = null;
  let activeTimeout = null;
  let requestToken = 0;
  let polling = false;
  let destroyed = false;
  let lastAnnouncedStatus = status.textContent.trim();

  const clampPercent = (value) => Math.max(0, Math.min(100, value));

  const announceStatus = (nextStatus) => {
    if (lastAnnouncedStatus === nextStatus) return;
    lastAnnouncedStatus = nextStatus;
    status.textContent = nextStatus;
  };

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B';
    const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB'];
    const unitIndex = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
    const value = bytes / (1024 ** unitIndex);
    return `${new Intl.NumberFormat('zh-CN', { maximumFractionDigits: unitIndex === 0 ? 0 : 1 }).format(value)} ${units[unitIndex]}`;
  };

  const formatUptime = (seconds) => {
    const totalMinutes = Math.floor(seconds / 60);
    if (totalMinutes < 1) return '不足 1 分钟';

    const days = Math.floor(totalMinutes / 1440);
    const hours = Math.floor((totalMinutes % 1440) / 60);
    const minutes = totalMinutes % 60;
    const parts = [];
    if (days > 0) parts.push(`${days} 天`);
    if (hours > 0) parts.push(`${hours} 小时`);
    if (days === 0 && minutes > 0) parts.push(`${minutes} 分钟`);
    return parts.join(' ');
  };

  const setGauge = (gauge, value) => {
    gauge.style.setProperty('--gauge', `${clampPercent(value)}%`);
  };

  const renderUnavailable = () => {
    root.dataset.state = 'unavailable';
    root.setAttribute('aria-busy', 'false');
    announceStatus('UNAVAILABLE');
    cpu.textContent = '--%';
    frequency.textContent = '-- MHz';
    memory.textContent = '--%';
    memoryDetail.textContent = '-- / --';
    uptime.textContent = '--';
    cpuGauge.style.removeProperty('--gauge');
    memoryGauge.style.removeProperty('--gauge');
  };

  const renderSnapshot = (snapshot) => {
    const keys = ['cpu_percent', 'cpu_frequency_mhz', 'memory_used_bytes', 'memory_total_bytes', 'memory_percent', 'uptime_seconds'];
    if (!snapshot || snapshot.online !== true || !keys.every((key) => typeof snapshot[key] === 'number' && Number.isFinite(snapshot[key]) && snapshot[key] >= 0)) {
      throw new Error('invalid telemetry response');
    }
    if (snapshot.memory_total_bytes === 0 || snapshot.memory_used_bytes > snapshot.memory_total_bytes) {
      throw new Error('invalid memory telemetry');
    }

    const cpuPercent = clampPercent(snapshot.cpu_percent);
    const memoryPercent = clampPercent(snapshot.memory_percent);
    root.dataset.state = 'online';
    root.setAttribute('aria-busy', 'false');
    announceStatus('ONLINE');
    cpu.textContent = `${percentFormat.format(cpuPercent)}%`;
    frequency.textContent = `${numberFormat.format(snapshot.cpu_frequency_mhz)} MHz`;
    memory.textContent = `${percentFormat.format(memoryPercent)}%`;
    memoryDetail.textContent = `${formatBytes(snapshot.memory_used_bytes)} / ${formatBytes(snapshot.memory_total_bytes)}`;
    uptime.textContent = formatUptime(snapshot.uptime_seconds);
    setGauge(cpuGauge, cpuPercent);
    setGauge(memoryGauge, memoryPercent);
  };

  const cancelRequest = () => {
    requestToken += 1;
    if (activeTimeout !== null) clearTimeout(activeTimeout);
    if (activeController !== null) activeController.abort();
    activeTimeout = null;
    activeController = null;
  };

  const fetchSnapshot = async () => {
    cancelRequest();
    const token = ++requestToken;
    const controller = new AbortController();
    activeController = controller;
    root.setAttribute('aria-busy', 'true');
    const timeout = setTimeout(() => controller.abort(), 8000);
    activeTimeout = timeout;

    try {
      const response = await fetch('/api/system-status', {
        cache: 'no-store',
        headers: { Accept: 'application/json' },
        signal: controller.signal,
      });
      if (!response.ok) throw new Error('telemetry unavailable');
      const snapshot = await response.json();
      if (token === requestToken) renderSnapshot(snapshot);
    } catch (error) {
      if (token === requestToken) renderUnavailable();
    } finally {
      clearTimeout(timeout);
      if (token === requestToken) {
        activeController = null;
        activeTimeout = null;
      }
    }
  };

  const startPolling = () => {
    if (destroyed || polling) return;
    if (document.visibilityState !== 'visible') {
      root.setAttribute('aria-busy', 'false');
      return;
    }
    polling = true;
    void fetchSnapshot();
    pollTimer = window.setInterval(() => {
      if (document.visibilityState === 'visible') void fetchSnapshot();
    }, 15000);
  };

  const stopPolling = () => {
    polling = false;
    if (pollTimer !== null) clearInterval(pollTimer);
    pollTimer = null;
    cancelRequest();
    root.setAttribute('aria-busy', 'false');
  };

  const handleVisibility = () => {
    if (destroyed) return;
    if (document.visibilityState === 'visible') startPolling();
    else stopPolling();
  };

  const handlePageShow = (event) => {
    if (!destroyed && event.persisted && document.visibilityState === 'visible') startPolling();
  };

  const handlePageHide = (event) => {
    stopPolling();
    if (event.persisted) return;

    destroyed = true;
    document.removeEventListener('visibilitychange', handleVisibility);
    window.removeEventListener('pageshow', handlePageShow);
    window.removeEventListener('pagehide', handlePageHide);
  };

  document.addEventListener('visibilitychange', handleVisibility);
  window.addEventListener('pageshow', handlePageShow);
  window.addEventListener('pagehide', handlePageHide);
  startPolling();
});
