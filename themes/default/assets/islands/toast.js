// Isla toast: botón que lanza una notificación efímera.
// También expone window.godianToast(message, type, duration) para uso programático.
function ensureContainer() {
	var c = document.getElementById('toast-container');
	if (!c) {
		c = document.createElement('div');
		c.id = 'toast-container';
		c.className = 'toast-container';
		c.setAttribute('aria-live', 'polite');
		document.body.appendChild(c);
	}
	return c;
}

function showToast(message, type, duration) {
	var c = ensureContainer();
	var t = document.createElement('div');
	t.className = 'toast toast-' + (type || 'info');
	t.setAttribute('role', 'status');
	t.textContent = message;
	c.appendChild(t);
	requestAnimationFrame(function () { t.classList.add('is-visible'); });
	setTimeout(function () {
		t.classList.remove('is-visible');
		setTimeout(function () { t.remove(); }, 300);
	}, duration || 3000);
}

// API global (se define una sola vez).
if (typeof window !== 'undefined' && !window.godianToast) {
	window.godianToast = function (message, type, duration) {
		showToast(message || 'Notificación', type || 'info', duration || 3000);
	};
}

export function mount(el, props) {
	var btn = el.querySelector('[data-toast-trigger]') || el;
	var message = props.message || 'Notificación';
	var type = props.type || 'info';
	var duration = Number(props.duration || 3000);
	if (btn && btn.tagName === 'BUTTON') {
		btn.addEventListener('click', function () { showToast(message, type, duration); });
	}
}
