// Isla spoiler: oculta el texto hasta pulsar el botón.
export function mount(el, props) {
	var btn = el.querySelector('.spoiler-btn');
	var body = el.querySelector('.spoiler-body');
	if (!btn || !body) return;
	btn.addEventListener('click', function () {
		var open = body.hidden;
		body.hidden = !open;
		el.classList.toggle('is-open', open);
		btn.setAttribute('aria-expanded', open ? 'true' : 'false');
	});
}
