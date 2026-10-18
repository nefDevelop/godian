// Isla carrusel: props.images = "url1,url2,url3", props.alt (opcional).
export function mount(el, props) {
	var imgs = (props.images || '')
		.split(',')
		.map(function (s) { return s.trim(); })
		.filter(Boolean);
	if (!imgs.length) return;

	var i = 0;
	el.innerHTML = '';

	var wrap = document.createElement('div');
	wrap.className = 'carousel';
	wrap.tabIndex = 0;
	wrap.setAttribute('role', 'group');
	wrap.setAttribute('aria-roledescription', 'carrusel');

	var prev = document.createElement('button');
	prev.type = 'button';
	prev.className = 'carousel-nav';
	prev.setAttribute('aria-label', 'Anterior');
	prev.textContent = '‹';

	var img = document.createElement('img');
	img.className = 'carousel-img';
	img.loading = 'lazy';
	img.decoding = 'async';
	img.alt = props.alt || '';

	var next = document.createElement('button');
	next.type = 'button';
	next.className = 'carousel-nav';
	next.setAttribute('aria-label', 'Siguiente');
	next.textContent = '›';

	var dots = document.createElement('div');
	dots.className = 'carousel-dots';

	var dotsBtns = imgs.map(function (_, idx) {
		var d = document.createElement('button');
		d.type = 'button';
		d.className = 'carousel-dot';
		d.setAttribute('aria-label', 'Ir a la imagen ' + (idx + 1));
		d.addEventListener('click', function () { go(idx); });
		dots.appendChild(d);
		return d;
	});

	function render() {
		img.src = imgs[i];
		dotsBtns.forEach(function (d, idx) { d.classList.toggle('is-active', idx === i); });
		// Texto alternativo (no aria-label, para no chocar con alt vacío).
		img.alt = (props.alt ? props.alt + ' — ' : '') + 'Imagen ' + (i + 1) + ' de ' + imgs.length;
	}
	function go(n) { i = (n + imgs.length) % imgs.length; render(); }

	prev.addEventListener('click', function () { go(i - 1); });
	next.addEventListener('click', function () { go(i + 1); });
	wrap.addEventListener('keydown', function (e) {
		if (e.key === 'ArrowLeft') { e.preventDefault(); go(i - 1); }
		else if (e.key === 'ArrowRight') { e.preventDefault(); go(i + 1); }
	});

	// Swipe táctil.
	var startX = null;
	wrap.addEventListener('touchstart', function (e) { startX = e.touches[0].clientX; }, { passive: true });
	wrap.addEventListener('touchend', function (e) {
		if (startX === null) return;
		var dx = e.changedTouches[0].clientX - startX;
		if (Math.abs(dx) > 40) { go(dx < 0 ? i + 1 : i - 1); }
		startX = null;
	});

	wrap.appendChild(prev);
	wrap.appendChild(img);
	wrap.appendChild(next);
	el.appendChild(wrap);
	el.appendChild(dots);
	render();
}
