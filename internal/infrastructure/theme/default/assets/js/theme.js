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

	// Mermaid (diagramas) — build IIFE que expone window.mermaid.
	if (document.querySelector('.mermaid') && !window.mermaid) {
		loadScript(BASE + 'mermaid/mermaid.min.js', function () {
			if (!window.mermaid) return;
			var dark = document.documentElement.classList.contains('dark');
			window.mermaid.initialize({ startOnLoad: false, theme: dark ? 'dark' : 'default', securityLevel: 'loose' });
			window.mermaid.run({ querySelector: '.mermaid' });
		});
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
(function () {
	var bar = document.getElementById('reading-progress');
	if (!bar) return;
	function update() {
		var h = document.documentElement;
		var max = h.scrollHeight - h.clientHeight;
		bar.style.width = (max > 0 ? (h.scrollTop / max) * 100 : 0) + '%';
	}
	update();
	document.addEventListener('scroll', update, { passive: true });
	window.addEventListener('resize', update);
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
