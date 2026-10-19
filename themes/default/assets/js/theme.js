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
			var id = decodeURIComponent(a.getAttribute('href').slice(1));
			(map[id] = map[id] || []).push(a);
		});
		var headings = Array.prototype.slice.call(document.querySelectorAll('.prose h2[id], .prose h3[id]'));
		var observer = new IntersectionObserver(function (entries) {
			entries.forEach(function (entry) {
				if (!entry.isIntersecting) return;
				tocLinks.forEach(function (a) { a.classList.remove('is-active'); });
				var links = map[entry.target.id];
				if (links) links.forEach(function (a) { a.classList.add('is-active'); });
			});
		}, { rootMargin: '-80px 0px -70% 0px' });
		headings.forEach(function (h) { observer.observe(h); });
	}
})();

// Aplica un cambio de tema con transición de colores y la retira al terminar.
var themeTransitionTimer = null;
window.godianThemeTransition = function (apply) {
	var root = document.documentElement;
	root.classList.add('theme-transition');
	try { apply(); } catch (e) {}
	clearTimeout(themeTransitionTimer);
	var duration = 260;
	try {
		var raw = getComputedStyle(root).getPropertyValue('--motion-base').trim();
		var n = parseFloat(raw);
		if (!isNaN(n)) duration = raw.indexOf('ms') >= 0 ? n : n * 1000;
	} catch (e) {}
	themeTransitionTimer = setTimeout(function () {
		root.classList.remove('theme-transition');
	}, duration + 30);
};

// Toggle de tema claro/oscuro.
document.addEventListener('click', function (e) {
	var btn = e.target.closest('.theme-toggle');
	if (!btn) return;
	var root = document.documentElement;
	var next = root.classList.contains('dark') ? 'light' : 'dark';
	var apply = function () {
		root.classList.remove('light', 'dark');
		root.classList.add(next);
		root.dataset.theme = next;
	};
	if (window.godianThemeTransition) window.godianThemeTransition(apply);
	else apply();
	try { localStorage.setItem('theme', next); } catch (err) {}
});

// Carga diferida de librerías (vendorizadas localmente, sin CDN).
(function () {
	var BASE = (window.__BASE_PATH__ || '') + '/assets/vendor/';

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

	// Mermaid (diagramas): se carga SOLO cuando el primer diagrama entra en pantalla,
	// y con el build ESM moderno (cada tipo de diagrama se descarga en un chunk aparte).
	var mermaidEl = document.querySelector('.mermaid');
	if (mermaidEl && !window.mermaid) {
		var loadMermaid = function () {
			if (window.mermaid) return;
			import(BASE + 'mermaid/mermaid.esm.min.mjs').then(function (mod) {
				var mermaid = mod.default || mod.mermaid || window.mermaid;
				if (!mermaid) return;
				window.mermaid = mermaid;
				var dark = document.documentElement.classList.contains('dark');
				mermaid.initialize({ startOnLoad: false, theme: dark ? 'dark' : 'default', securityLevel: 'loose' });
				mermaid.run({ querySelector: '.mermaid' });
			}).catch(function () {});
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

	// KaTeX (matemáticas): se carga cuando el contenido entra en pantalla.
	// El marcador [data-math] está oculto, así que se observa .prose (tiene tamaño).
	var mathEl = document.querySelector('[data-math]');
	if (mathEl && !window.renderMathInElement) {
		var loadMath = function () {
			if (window.renderMathInElement) return;
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
		};
		// Carga diferida pero garantizada: idle con timeout; si no hay API, inmediata.
		if ('requestIdleCallback' in window) {
			requestIdleCallback(loadMath, { timeout: 1200 });
		} else {
			setTimeout(loadMath, 0);
		}
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

// Facade de vídeo (shortcode {{youtube}}): carga el iframe (youtube-nocookie)
// solo al pulsar play, sin peticiones ni cookies de terceros hasta entonces.
document.addEventListener('click', function (e) {
	var btn = e.target.closest && e.target.closest('.video-embed .video-play');
	if (!btn) return;
	var box = btn.closest('.video-embed');
	var id = box && box.getAttribute('data-youtube-id');
	if (!id) return;
	var iframe = document.createElement('iframe');
	iframe.src = 'https://www.youtube-nocookie.com/embed/' + id + '?autoplay=1&mute=1&rel=0&playsinline=1';
	iframe.title = btn.getAttribute('aria-label') || 'Video';
	iframe.allow = 'autoplay; fullscreen; picture-in-picture';
	iframe.allowFullscreen = true;
	iframe.className = 'video-iframe';
	box.appendChild(iframe);
	btn.style.display = 'none';
});

// TOC flotante (móvil): el botón abre/cierra el índice.
(function () {
	var fab = document.querySelector('.toc-fab');
	var panel = document.querySelector('.toc-float-panel');
	if (!fab || !panel) return;
	function set(open) {
		fab.setAttribute('aria-expanded', open ? 'true' : 'false');
		panel.classList.toggle('is-open', open);
	}
	set(false);
	fab.addEventListener('click', function (e) {
		e.stopPropagation();
		set(fab.getAttribute('aria-expanded') !== 'true');
	});
	panel.addEventListener('click', function (e) {
		if (e.target.closest && e.target.closest('a')) set(false);
	});
	document.addEventListener('click', function (e) {
		if (!panel.contains(e.target) && !fab.contains(e.target)) set(false);
	});
	document.addEventListener('keydown', function (e) {
		if (e.key === 'Escape') set(false);
	});
})();

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

// Animación de altura para <details> de contenido (callouts plegables, bloques
// de código plegables y <details> nativos). La isla accordion tiene su propio
// controlador y se excluye.
(function () {
	if (typeof Element === 'undefined' || !Element.prototype.animate) return;
	if (window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;

	function token(name, fallback) {
		var value = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
		return value || fallback;
	}
	function ms(name, fallback) {
		var value = token(name, '');
		var n = parseFloat(value);
		if (!isNaN(n)) return value.indexOf('ms') >= 0 ? n : n * 1000;
		return fallback;
	}

	var details = document.querySelectorAll('.prose details:not(.accordion-item), details.smooth:not(.accordion-item)');
	details.forEach(function (d) {
		var summary = d.querySelector(':scope > summary');
		if (!summary || d.classList.contains('smooth-js')) return;

		// El contenido nativo de <details> son los nodos posteriores al summary.
		// Se agrupan para poder animar su altura.
		var body = document.createElement('div');
		body.className = 'details-body';
		var node = summary.nextSibling;
		while (node) {
			var next = node.nextSibling;
			body.appendChild(node);
			node = next;
		}
		d.appendChild(body);
		d.classList.add('smooth-js');

		var anim = null;
		summary.addEventListener('click', function (e) {
			e.preventDefault();
			if (anim) return;

			var opening = !d.open;
			if (opening) d.open = true;
			var height = body.scrollHeight;
			body.style.overflow = 'hidden';

			var duration = opening ? ms('--motion-base', 220) : ms('--motion-fast', 140);
			var easing = opening
				? token('--motion-ease-enter', 'cubic-bezier(0, 0, 0.2, 1)')
				: token('--motion-ease-exit', 'cubic-bezier(0.4, 0, 1, 1)');
			var frames = opening
				? [{ height: '0px', opacity: 0 }, { height: height + 'px', opacity: 1 }]
				: [{ height: height + 'px', opacity: 1 }, { height: '0px', opacity: 0 }];

			anim = body.animate(frames, { duration: duration, easing: easing });
			anim.onfinish = function () {
				if (!opening) d.open = false;
				body.style.overflow = '';
				body.style.height = '';
				anim = null;
			};
			anim.oncancel = function () {
				body.style.overflow = '';
				body.style.height = '';
				anim = null;
			};
		});
	});
})();

// Task lists interactivas: solo se activan con la clase checklist-interactive.
// El estado se guarda en localStorage; el Markdown fuente no cambia.
(function () {
	var lists = document.querySelectorAll('.prose ul.checklist-interactive');
	lists.forEach(function (ul, listIndex) {
		var boxes = ul.querySelectorAll('input[type="checkbox"]');
		if (!boxes.length) return;
		var key = 'checklist:' + location.pathname + ':' + listIndex;
		var saved = [];
		try { saved = JSON.parse(localStorage.getItem(key) || '[]'); } catch (e) {}
		boxes.forEach(function (box, i) {
			box.disabled = false;
			var li = box.closest('li');
			var text = li ? li.textContent.trim() : '';
			if (text) box.setAttribute('aria-label', text);
			if (saved[i] !== undefined) box.checked = !!saved[i];
			box.addEventListener('change', function () {
				var state = [];
				boxes.forEach(function (b) { state.push(b.checked); });
				try { localStorage.setItem(key, JSON.stringify(state)); } catch (e) {}
			});
		});
	});
})();
