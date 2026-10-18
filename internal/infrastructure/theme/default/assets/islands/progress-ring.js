// Isla anillo de progreso (SVG circular).
export function mount(el, props) {
  var value = Number(props.value || 0);
  var max = Number(props.max || 100);
  var size = Number(props.size || 80);
  var pct = max > 0 ? Math.min(1, value / max) : 0;
  var r = (size - 8) / 2;
  var c = 2 * Math.PI * r;
  var offset = c * (1 - pct);
  el.classList.add('progress-ring');
  el.setAttribute('role', 'progressbar');
  el.setAttribute('aria-label', props.label || 'Progreso');
  el.setAttribute('aria-valuemin', '0');
  el.setAttribute('aria-valuemax', String(max));
  el.setAttribute('aria-valuenow', String(value));
  el.innerHTML = '';
  var svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
  svg.setAttribute('width', size);
  svg.setAttribute('height', size);
  svg.setAttribute('viewBox', '0 0 ' + size + ' ' + size);
  var bg = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
  bg.setAttribute('cx', size / 2); bg.setAttribute('cy', size / 2); bg.setAttribute('r', r);
  bg.setAttribute('fill', 'none'); bg.setAttribute('stroke', 'var(--border)'); bg.setAttribute('stroke-width', '6');
  var fg = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
  fg.setAttribute('cx', size / 2); fg.setAttribute('cy', size / 2); fg.setAttribute('r', r);
  fg.setAttribute('fill', 'none'); fg.setAttribute('stroke', 'var(--accent)'); fg.setAttribute('stroke-width', '6');
  fg.setAttribute('stroke-linecap', 'round');
  fg.setAttribute('stroke-dasharray', c);
  fg.setAttribute('stroke-dashoffset', offset);
  fg.setAttribute('transform', 'rotate(-90 ' + size / 2 + ' ' + size / 2 + ')');
  var ttl = document.createElementNS('http://www.w3.org/2000/svg', 'title');
  ttl.textContent = props.label || (value + ' / ' + max);
  svg.appendChild(ttl);
  svg.appendChild(bg); svg.appendChild(fg);
  el.appendChild(svg);
  if (props.label) {
    var lab = document.createElement('span');
    lab.className = 'progress-ring-label';
    lab.textContent = props.label;
    el.appendChild(lab);
  }
}
