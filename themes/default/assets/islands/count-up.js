import { prefersReducedMotion } from '/assets/js/motion.js';

// Isla count-up: número que se anima hasta el valor indicado.
export function mount(el, props) {
  var target = Number(props.value || 0);
  var duration = Number(props.duration || 1200);
  var prefix = props.prefix || '';
  var suffix = props.suffix || '';
  el.classList.add('count-up');
  var span = document.createElement('span');
  span.className = 'count-up-value';
  span.textContent = prefix + '0' + suffix;
  el.appendChild(span);
  if (prefersReducedMotion()) {
    span.textContent = prefix + target.toLocaleString() + suffix;
    return;
  }
  var start = (typeof performance !== 'undefined') ? performance.now() : Date.now();
  function step(now) {
    var t = Math.min(1, (now - start) / duration);
    var eased = 1 - Math.pow(1 - t, 3);
    var val = Math.round(target * eased);
    span.textContent = prefix + val.toLocaleString() + suffix;
    if (t < 1) requestAnimationFrame(step);
  }
  requestAnimationFrame(step);
}
