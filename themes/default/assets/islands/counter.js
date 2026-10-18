export function mount(el, props) {
	var n = Number(props.start || 0);
	var out = el.querySelector('[data-counter-out]');
	var inc = el.querySelector('[data-counter-inc]');
	var dec = el.querySelector('[data-counter-dec]');
	function render() { if (out) out.textContent = n; }
	if (inc) inc.addEventListener('click', function () { n++; render(); });
	if (dec) dec.addEventListener('click', function () { n--; render(); });
	render();
}
