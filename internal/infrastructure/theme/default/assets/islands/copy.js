// Isla copiar: props.text (texto a copiar; por defecto, la URL actual).
export function mount(el, props) {
	var btn = el.querySelector('.copy-btn') || el;
	btn.addEventListener('click', function () {
		var text = btn.getAttribute('data-copy') || location.href;
		var done = function () {
			var old = btn.textContent;
			btn.textContent = 'Copied!';
			setTimeout(function () { btn.textContent = old; }, 1500);
		};
		if (navigator.clipboard) {
			navigator.clipboard.writeText(text).then(done, done);
		}
	});
}
