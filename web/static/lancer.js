document.querySelectorAll('[data-drawer-target]').forEach((cue) => {
  cue.addEventListener('click', () => cue.classList.add('is-used'), { once: true });
});
