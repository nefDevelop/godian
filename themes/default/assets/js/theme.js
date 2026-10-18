(function () {
	// Botones de copiar en bloques de código.
	document.addEventListener('click', function (e) {
		var btn = e.target.closest('.code-copy');
		if (!btn) return;
		var block = btn.closest('.code-block');
		var pre = block && block.querySelector('pre');
		if (!pre) return;
		var text = pre.innerText;
		var done = function () {
			var old = btn.textContent;
			btn.textContent = (btn.getAttribute('data-copied') || (window.__i18n && window.__i18n.copied) || 'Copied!');
			setTimeout(function () { btn.textContent = old; }, 1500);
		};
		if (navigator.clipboard) {
			navigator.clipboard.writeText(text).then(done, done);
		}
	});

	// Resalta la sección activa en la tabla de contenidos.
	var tocLinks = Array.prototype.slice.call(document.querySelectorAll('.toc a[href^="#"]'));
	if (tocLinks.length && 'IntersectionObserver' in window) {
		var map = {};
		tocLinks.forEach(function (a) {
			map[decodeURIComponent(a.getAttribute('href').slice(1))] = a;
		});
		var headings = Array.prototype.slice.call(document.querySelectorAll('.prose h2[id], .prose h3[id]'));
		var observer = new IntersectionObserver(function (entries) {
			entries.forEach(function (entry) {
				if (!entry.isIntersecting) return;
				tocLinks.forEach(function (a) { a.classList.remove('is-active'); });
				var link = map[entry.target.id];
				if (link) link.classList.add('is-active');
			});
		}, { rootMargin: '-80px 0px -70% 0px' });
		headings.forEach(function (h) { observer.observe(h); });
	}
})();

// Toggle de tema claro/oscuro.
document.addEventListener('click', function (e) {
	var btn = e.target.closest('.theme-toggle');
	if (!btn) return;
	var root = document.documentElement;
	var next = root.classList.contains('dark') ? 'light' : 'dark';
	root.classList.remove('light', 'dark');
	root.classList.add(next);
	root.dataset.theme = next;
	try { localStorage.setItem('theme', next); } catch (err) {}
});

// Carga diferida de librerías (vendorizadas localmente, sin CDN).
(function () {
	var BASE = '/assets/vendor/';

	function loadStyle(href) {
		var l = document.createElement('link');
		l.rel = 'stylesheet';
		l.href = href;
		document.head.appendChild(l);
	}
	function loadScript(src, onload) {
		var s = document.createElement('script');
		s.src = src;
		if (onload) s.onload = onload;
		document.head.appendChild(s);
	}

	// Mermaid (diagramas): se carga SOLO cuando el primer diagrama entra en pantalla.
	// Evita descargar ~2,8 MB de JS si el usuario no llega a verlo.
	var mermaidEl = document.querySelector('.mermaid');
	if (mermaidEl && !window.mermaid) {
		var loadMermaid = function () {
			if (window.mermaid) return;
			loadScript(BASE + 'mermaid/mermaid.min.js', function () {
				if (!window.mermaid) return;
				var dark = document.documentElement.classList.contains('dark');
				window.mermaid.initialize({ startOnLoad: false, theme: dark ? 'dark' : 'default', securityLevel: 'loose' });
				window.mermaid.run({ querySelector: '.mermaid' });
			});
		};
		if ('IntersectionObserver' in window) {
			var mObs = new IntersectionObserver(function (entries) {
				if (entries.some(function (e) { return e.isIntersecting; })) { mObs.disconnect(); loadMermaid(); }
			}, { rootMargin: '300px' });
			mObs.observe(mermaidEl);
		} else {
			loadMermaid();
		}
	}

	// KaTeX (matemáticas)
	if (document.querySelector('[data-math]') && !window.renderMathInElement) {
		loadStyle(BASE + 'katex/katex.min.css');
		loadScript(BASE + 'katex/katex.min.js', function () {
			loadScript(BASE + 'katex/contrib/auto-render.min.js', function () {
				if (!window.renderMathInElement) return;
				window.renderMathInElement(document.querySelector('.prose') || document.body, {
					delimiters: [
						{ left: '$$', right: '$$', display: true },
						{ left: '$', right: '$', display: false },
						{ left: '\\(', right: '\\)', display: false },
						{ left: '\\[', right: '\\]', display: true }
					],
					throwOnError: false,
					ignoredTags: ['script', 'noscript', 'style', 'textarea', 'pre', 'code', 'option']
				});
				document.querySelectorAll('[data-math]').forEach(function (el) { el.remove(); });
			});
		});
	}

	// PhotoSwipe (lightbox)
	var pswpImgs = document.querySelectorAll('.prose picture img, .prose > img');
	if (pswpImgs.length && !window.__pswpInit) {
		window.__pswpInit = true;
		loadStyle(BASE + 'photoswipe/photoswipe.css');
		import(BASE + 'photoswipe/photoswipe.esm.min.js').then(function (mod) {
			var PhotoSwipe = mod.default || mod.PhotoSwipe;
			pswpImgs.forEach(function (img) {
				img.style.cursor = 'zoom-in';
				img.addEventListener('click', function () {
					var pswp = new PhotoSwipe({
						dataSource: [{ src: img.currentSrc || img.src, w: img.naturalWidth, h: img.naturalHeight, alt: img.alt }],
						index: 0
					});
					pswp.init();
				});
			});
		}).catch(function () {});
	}
})();

// Menú de navegación en móvil.
(function () {
	var btn = document.querySelector('.menu-toggle');
	if (!btn) return;
	btn.addEventListener('click', function () {
		var header = btn.closest('.site-header');
		var open = header.classList.toggle('nav-open');
		btn.setAttribute('aria-expanded', open ? 'true' : 'false');
	});
	document.querySelectorAll('.site-nav a').forEach(function (a) {
		a.addEventListener('click', function () {
			btn.closest('.site-header').classList.remove('nav-open');
			btn.setAttribute('aria-expanded', 'false');
		});
	});
})();

// Barra de progreso de lectura.
// Se usa transform (scaleX) en lugar de width y se limita a un frame por scroll
// (requestAnimationFrame) para evitar "layout thrash" y el aviso de Firefox sobre
// efectos ligados al scroll.
(function () {
	var bar = document.getElementById('reading-progress');
	if (!bar) return;
	var ticking = false;
	function update() {
		ticking = false;
		var h = document.documentElement;
		var max = h.scrollHeight - h.clientHeight;
		var ratio = max > 0 ? (h.scrollTop / max) : 0;
		if (ratio < 0) ratio = 0;
		if (ratio > 1) ratio = 1;
		bar.style.transform = 'scaleX(' + ratio + ')';
	}
	function onScroll() {
		if (ticking) return;
		ticking = true;
		requestAnimationFrame(update);
	}
	update();
	document.addEventListener('scroll', onScroll, { passive: true });
	window.addEventListener('resize', onScroll);
})();

// Botón volver arriba.
(function () {
	var btn = document.getElementById('back-to-top');
	if (!btn) return;
	function toggle() { btn.classList.toggle('is-visible', window.scrollY > 400); }
	toggle();
	document.addEventListener('scroll', toggle, { passive: true });
	btn.addEventListener('click', function () { window.scrollTo({ top: 0, behavior: 'smooth' }); });
})();

// Copiar enlace para compartir.
document.addEventListener('click', function (e) {
	var btn = e.target.closest('[data-copy-url]');
	if (!btn) return;
	var url = btn.getAttribute('data-copy-url');
	var done = function () {
		var old = btn.textContent;
		btn.textContent = (btn.getAttribute('data-copied') || (window.__i18n && window.__i18n.copied) || 'Copied!');
		setTimeout(function () { btn.textContent = old; }, 1500);
	};
	if (navigator.clipboard) navigator.clipboard.writeText(url).then(done, done);
});

// Tablas ordenables (table.sortable).
document.addEventListener('click', function (e) {
	var th = e.target.closest('table.sortable th');
	if (!th) return;
	var table = th.closest('table');
	var tbody = table.querySelector('tbody');
	if (!tbody) return;
	var col = Array.prototype.indexOf.call(th.parentElement.children, th);
	var dir = table.dataset.sortCol === String(col) && table.dataset.sortDir === 'asc' ? 'desc' : 'asc';
	table.dataset.sortCol = String(col);
	table.dataset.sortDir = dir;
	var rows = Array.prototype.slice.call(tbody.querySelectorAll('tr'));
	rows.sort(function (a, b) {
		var va = (a.children[col] ? a.children[col].textContent : '').trim();
		var vb = (b.children[col] ? b.children[col].textContent : '').trim();
		var na = parseFloat(va), nb = parseFloat(vb);
		if (!isNaN(na) && !isNaN(nb)) return dir === 'asc' ? na - nb : nb - na;
		return dir === 'asc' ? va.localeCompare(vb) : vb.localeCompare(va);
	});
	rows.forEach(function (r) { tbody.appendChild(r); });
});
