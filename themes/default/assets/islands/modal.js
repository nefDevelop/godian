// Isla modal: abre un <dialog> nativo al pulsar el botón disparador.
export function mount(el, props) {
	var dialog = el.querySelector('dialog');
	var btn = el.querySelector('[data-modal-open]');
	if (!dialog) return;

	// Etiqueta accesible a partir del título.
	var title = dialog.querySelector('.modal-title');
	if (title) {
		if (!title.id) title.id = 'modal-title-' + Math.random().toString(36).slice(2, 7);
		dialog.setAttribute('aria-labelledby', title.id);
	}

	var lastFocus = null;
	function open() {
		lastFocus = document.activeElement;
		if (typeof dialog.showModal === 'function') dialog.showModal(); else dialog.setAttribute('open', '');
		var target = dialog.querySelector('[data-modal-close]') || dialog;
		if (target && typeof target.focus === 'function') target.focus();
	}
	function close() {
		if (typeof dialog.close === 'function') dialog.close(); else dialog.removeAttribute('open');
		// Devolver el foco al elemento que abrió el modal.
		if (lastFocus && typeof lastFocus.focus === 'function') lastFocus.focus();
	}

	if (btn) btn.addEventListener('click', open);
	dialog.querySelectorAll('[data-modal-close]').forEach(function (b) { b.addEventListener('click', close); });
	dialog.addEventListener('click', function (e) { if (e.target === dialog) close(); });
	dialog.addEventListener('close', function () {
		if (lastFocus && typeof lastFocus.focus === 'function') lastFocus.focus();
	});
}
