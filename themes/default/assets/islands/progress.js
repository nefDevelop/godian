// Isla barra de progreso: props.value / props.max / props.label.
export function mount(el, props) {
	var value = Number(el.getAttribute('data-value') || props.value || 0);
	var max = Number(el.getAttribute('data-max') || props.max || 100);
	var pct = max > 0 ? Math.max(0, Math.min(100, (value / max) * 100)) : 0;
	var bar = el.querySelector('.progress-bar');
	if (bar) {
		bar.style.width = pct + '%';
		bar.setAttribute('aria-valuenow', String(value));
		bar.setAttribute('aria-valuemax', String(max));
	}
}
