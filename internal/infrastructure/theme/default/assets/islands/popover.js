// Isla popover: muestra un popover al hover o al clic.
export function mount(el, props) {
  var btn = el.querySelector('.popover-trigger');
  var pop = el.querySelector('.popover-content');
  if (!btn || !pop) return;
  var mode = props.mode || 'hover';
  function show() { pop.hidden = false; pop.classList.add('is-open'); }
  function hide() { pop.hidden = true; pop.classList.remove('is-open'); }
  hide();
  if (mode === 'hover') {
    btn.addEventListener('mouseenter', show);
    btn.addEventListener('mouseleave', hide);
    btn.addEventListener('focus', show);
    btn.addEventListener('blur', hide);
  } else {
    btn.addEventListener('click', function (e) { e.stopPropagation(); if (pop.hidden) show(); else hide(); });
    document.addEventListener('click', function (e) { if (!el.contains(e.target)) hide(); });
  }
}
