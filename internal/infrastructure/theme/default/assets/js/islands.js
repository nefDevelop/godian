// Runtime de islas: carga bajo demanda los componentes [data-island].
(function () {
	var manifest = null;
	var loaded = {};

	function loadManifest() {
		if (manifest) return Promise.resolve(manifest);
		return fetch('/assets/islands/manifest.json')
			.then(function (r) { return r.ok ? r.json() : {}; })
			.then(function (m) { manifest = m; return m; })
			.catch(function () { manifest = {}; return manifest; });
	}

	function injectCSS(url) {
		if (!url || document.querySelector('link[data-island-css="' + url + '"]')) return;
		var l = document.createElement('link');
		l.rel = 'stylesheet';
		l.href = url;
		l.setAttribute('data-island-css', url);
		document.head.appendChild(l);
	}

	function loadModule(name, entry) {
		if (loaded[name]) return loaded[name];
		if (entry.css) injectCSS(entry.css);
		loaded[name] = entry.js ? import(entry.js) : Promise.resolve(null);
		return loaded[name];
	}

	function mount(el) {
		var name = el.getAttribute('data-island');
		var props = {};
		try { props = JSON.parse(el.getAttribute('data-props') || '{}'); } catch (e) {}
		var strategy = el.getAttribute('data-island-load') || 'idle';

		var run = function () {
			loadManifest().then(function (m) {
				var entry = m[name];
				if (!entry || !entry.js) return;
				return loadModule(name, entry).then(function (mod) {
					if (!mod) return;
					var mountFn = mod.mount || (mod.default && mod.default.mount);
					if (typeof mountFn === 'function') mountFn(el, props);
				});
			});
		};

		if (strategy === 'visible' && 'IntersectionObserver' in window) {
			var obs = new IntersectionObserver(function (entries) {
				if (entries.some(function (e) { return e.isIntersecting; })) {
					obs.disconnect();
					run();
				}
			});
			obs.observe(el);
			return;
		}
		if (strategy === 'idle' && 'requestIdleCallback' in window) {
			requestIdleCallback(run);
			return;
		}
		run();
	}

	function init() {
		document.querySelectorAll('[data-island]').forEach(mount);
	}

	if (document.readyState === 'loading') {
		document.addEventListener('DOMContentLoaded', init);
	} else {
		init();
	}
})();
