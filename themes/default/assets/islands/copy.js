// Isla copiar: props.text (texto a copiar; por defecto, la URL actual).
export function mount(el, props) {
	var btn = el.querySelector('.copy-btn') || el;
	btn.addEventListener('click', function () {
		var text = btn.getAttribute('data-copy') || location.href;
		var done = function () {
			var old = btn.textContent;
			btn.textContent = (window.__i18n && window.__i18n.copied) || 'Copied!';
			btn.classList.add('is-copied');
			setTimeout(function () { btn.textContent = old; btn.classList.remove('is-copied'); }, 1500);
		};
		if (navigator.clipboard) {
			navigator.clipboard.writeText(text).then(done, done);
		}
	});
}
