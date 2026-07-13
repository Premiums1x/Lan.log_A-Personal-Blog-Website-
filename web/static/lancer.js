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
