document.querySelectorAll('[data-expand-list]').forEach((root) => {
  const items = [...root.querySelectorAll('[data-expand-item]')];
  const trigger = root.querySelector('[data-expand-trigger]');
  const initialCount = Math.max(1, Number.parseInt(root.dataset.expandInitial || '6', 10) || 6);
  const collapsedItems = items.slice(initialCount);
  if (!trigger || collapsedItems.length === 0) return;

  collapsedItems.forEach((item) => { item.hidden = true; });
  trigger.hidden = false;
  trigger.setAttribute('aria-expanded', 'false');

  trigger.addEventListener('click', () => {
    const hiddenItems = items.filter((item) => item.hidden);
    hiddenItems.forEach((item, index) => {
      item.hidden = false;
      requestAnimationFrame(() => {
        item.style.transitionDelay = `${Math.min(index * 45, 225)}ms`;
        item.classList.add('is-in');
      });
    });
    root.dataset.expanded = 'true';
    trigger.setAttribute('aria-expanded', 'true');
    trigger.hidden = true;
  }, { once: true });
});
document.querySelectorAll('[data-archive-sort]').forEach((control) => {
  const archive = control.closest('.lancer-archive');
  const seasonLine = archive?.querySelector('.season-line');
  const label = control.querySelector('[data-archive-order-label]');
  if (!archive || !seasonLine || !label) return;

  control.dataset.sortReady = 'true';

  const syncExpansion = (year) => {
    const items = [...year.querySelectorAll('[data-expand-item]')];
    const trigger = year.querySelector('[data-expand-trigger]');
    const initialCount = Math.max(1, Number.parseInt(year.dataset.expandInitial || '6', 10) || 6);
    const expanded = year.dataset.expanded === 'true';
    items.forEach((item, index) => { item.hidden = !expanded && index >= initialCount; });
    if (trigger) {
      trigger.hidden = expanded || items.length <= initialCount;
      trigger.setAttribute('aria-expanded', String(expanded));
    }
  };

  const applyOrder = (order) => {
    const direction = order === 'oldest' ? 1 : -1;
    const years = [...archive.querySelectorAll('[data-archive-year]')];
    years.sort((a, b) => (Number(a.dataset.archiveYear) - Number(b.dataset.archiveYear)) * direction);
    years.forEach((year) => {
      const notes = year.querySelector('.season-notes');
      const posts = [...year.querySelectorAll('[data-published]')];
      posts.sort((a, b) => a.dataset.published.localeCompare(b.dataset.published) * direction);
      posts.forEach((post) => notes?.append(post));
      syncExpansion(year);
      seasonLine.append(year);
    });

    const isOldest = order === 'oldest';
    control.dataset.order = order;
    archive.dataset.archiveOrder = order;
    label.textContent = isOldest ? 'OLDEST ↑' : 'NEWEST ↓';
    control.setAttribute('aria-label', isOldest ? '切换文章排序，当前从旧到新' : '切换文章排序，当前从新到旧');
  };

  control.addEventListener('click', () => {
    applyOrder(control.dataset.order === 'newest' ? 'oldest' : 'newest');
  });
});
document.querySelectorAll('[data-drawer-handle]').forEach((handle) => {
  const drawer = handle.closest('.page-drawer');
  if (!drawer) return;

  let pointerID = null;
  let startY = 0;
  let startScrollY = 0;
  let pendingScrollY = 0;
  let frameID = null;
  let suppressClick = false;
  let previousScrollBehavior = '';

  const applyDrag = () => {
    window.scrollTo({ top: pendingScrollY, behavior: 'auto' });
    frameID = null;
  };

  handle.addEventListener('pointerdown', (event) => {
    if (event.button !== undefined && event.button !== 0) return;
    pointerID = event.pointerId;
    startY = event.clientY;
    startScrollY = window.scrollY;
    pendingScrollY = startScrollY;
    suppressClick = false;
    previousScrollBehavior = document.documentElement.style.scrollBehavior;
    document.documentElement.style.scrollBehavior = 'auto';
    handle.setPointerCapture(pointerID);
    drawer.classList.add('is-dragging');
  });

  handle.addEventListener('pointermove', (event) => {
    if (event.pointerId !== pointerID) return;
    const dragDistance = startY - event.clientY;
    if (Math.abs(dragDistance) > 4) suppressClick = true;
    pendingScrollY = Math.max(0, startScrollY + dragDistance);
    if (frameID === null) frameID = requestAnimationFrame(applyDrag);
    event.preventDefault();
  });

  const finishDrag = (event) => {
    if (event.pointerId !== pointerID) return;
    if (handle.hasPointerCapture(pointerID)) handle.releasePointerCapture(pointerID);
    pointerID = null;
    document.documentElement.style.scrollBehavior = previousScrollBehavior;
    drawer.classList.remove('is-dragging');
  };

  handle.addEventListener('pointerup', finishDrag);
  handle.addEventListener('pointercancel', finishDrag);
  handle.addEventListener('click', (event) => {
    if (suppressClick) {
      event.preventDefault();
      suppressClick = false;
      return;
    }
    const reduceMotion = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches;
    drawer.scrollIntoView({ behavior: reduceMotion ? 'auto' : 'smooth', block: 'start' });
  });
});

document.querySelectorAll('.interactive-influences').forEach((root) => {
  const cards = [...root.querySelectorAll('[data-influence]')];
  const preview = root.querySelector('.influence-preview');
  const layers = [...root.querySelectorAll('.influence-preview-image')];
  const index = root.querySelector('#influence-preview-index');
  const title = root.querySelector('#influence-preview-title');
  const copy = root.querySelector('#influence-preview-copy');
  let activeLayer = 0;

  const selected = cards.find((card) => card.getAttribute('aria-pressed') === 'true');
  if (selected && layers[0]) {
    layers[0].style.setProperty('--influence-image', `url("${selected.dataset.image}")`);
    layers[0].style.backgroundPosition = selected.dataset.position || 'center';
    layers[0].style.backgroundSize = selected.dataset.size || 'cover';
  }

  const activate = (card) => {
    if (root.dataset.active === card.dataset.influence) return;

    cards.forEach((item) => item.setAttribute('aria-pressed', String(item === card)));
    root.dataset.active = card.dataset.influence;
    index.textContent = card.dataset.index;
    title.textContent = card.querySelector('h3').textContent;
    copy.textContent = card.querySelector('p').textContent;

    if (preview && layers.length === 2 && card.dataset.image) {
      const nextLayer = activeLayer === 0 ? 1 : 0;
      layers[nextLayer].style.setProperty('--influence-image', `url("${card.dataset.image}")`);
      layers[nextLayer].style.backgroundPosition = card.dataset.position || 'center';
      layers[nextLayer].style.backgroundSize = card.dataset.size || 'cover';
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

  const percentFormat = new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 1 });
  const numberFormat = new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 0 });
  const sampleLimit = 40;
  const chartHistory = { cpu: [], frequency: [], memory: [], uptime: [] };
  const chartViews = new Map();
  let frequencyCeiling = 5000;
  root.querySelectorAll('[data-chart]').forEach((chart) => {
    chartViews.set(chart.dataset.chart, {
      line: chart.querySelector('[data-chart-line]'),
      area: chart.querySelector('[data-chart-area]'),
      point: chart.querySelector('[data-chart-point]'),
      samples: root.querySelector(`[data-chart-samples="${chart.dataset.chart}"]`),
    });
  });

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

  const renderChart = (key) => {
    const view = chartViews.get(key);
    const history = chartHistory[key];
    if (!view || !history.length) return;

    const width = 240;
    const height = 92;
    const padding = 6;
    const usableWidth = width - (padding * 2);
    const usableHeight = height - (padding * 2);
    let min = 0;
    let max = 100;
    if (key === 'frequency') {
      const observedPeak = Math.max(...history);
      frequencyCeiling = Math.max(
        frequencyCeiling,
        Math.ceil((observedPeak * 1.15) / 500) * 500,
      );
      max = Math.max(5000, frequencyCeiling);
    } else if (key === 'uptime') {
      min = 0;
      max = 1;
    }

    const points = history.map((value, index) => {
      const x = width - padding - (((history.length - 1) - index) / (sampleLimit - 1) * usableWidth);
      const y = key === 'uptime'
        ? height / 2
        : height - padding - (((value - min) / (max - min)) * usableHeight);
      return [x, y];
    });
    const linePath = points.map(([x, y], index) => `${index === 0 ? 'M' : 'L'}${x.toFixed(2)} ${y.toFixed(2)}`).join(' ');
    const [firstX] = points[0];
    const [lastX, lastY] = points[points.length - 1];
    const areaPath = key !== 'uptime' && points.length > 1
      ? `${linePath} L${lastX.toFixed(2)} ${height - padding} L${firstX.toFixed(2)} ${height - padding} Z`
      : '';

    view.line.setAttribute('d', linePath);
    view.area.setAttribute('d', areaPath);
    view.point.setAttribute('cx', lastX.toFixed(2));
    view.point.setAttribute('cy', lastY.toFixed(2));
    view.point.setAttribute('opacity', '1');
    view.samples.textContent = `${String(history.length).padStart(2, '0')} / ${sampleLimit}`;
  };

  const appendChartPoint = (key, value) => {
    const history = chartHistory[key];
    if (!history || !Number.isFinite(value)) return;
    history.push(value);
    if (history.length > sampleLimit) history.shift();
    renderChart(key);
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
    appendChartPoint('cpu', cpuPercent);
    appendChartPoint('frequency', snapshot.cpu_frequency_mhz);
    appendChartPoint('memory', memoryPercent);
    appendChartPoint('uptime', snapshot.uptime_seconds);
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
