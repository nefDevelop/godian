(function () {
	'use strict';

	var csrfMeta = document.querySelector('meta[name="admin-csrf"]');
	var csrf = csrfMeta ? csrfMeta.getAttribute('content') : '';
	var body = document.getElementById('body');
	var preview = document.getElementById('admin-preview');
	var status = document.querySelector('[data-preview-status]');
	var previewTimer = 0;
	var mathPromise = null;

	function setStatus(text) {
		if (status) status.textContent = text;
	}

	function renderPreview() {
		if (!body || !preview) return;
		setStatus('Renderizando…');
		fetch('/admin/preview', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8',
				'X-Admin-CSRF': csrf
			},
			body: new URLSearchParams({ content: body.value })
		}).then(function (response) {
			if (!response.ok) throw new Error(response.statusText);
			return response.text();
		}).then(function (html) {
			preview.innerHTML = html || '<p class="admin-empty">Sin contenido.</p>';
			setStatus('Actualizado');
			renderMath();
		}).catch(function () {
			setStatus('Error al previsualizar');
		});
	}

	function schedulePreview() {
		clearTimeout(previewTimer);
		previewTimer = setTimeout(renderPreview, 220);
	}

	function renderMath() {
		if (!preview) return;
		var text = preview.textContent || '';
		if (text.indexOf('$') === -1 && text.indexOf('\\(') === -1 && text.indexOf('\\[') === -1) return;
		if (window.renderMathInElement) {
			doMath();
			return;
		}
		if (!mathPromise) mathPromise = loadKatex();
		mathPromise.then(doMath).catch(function () {});
	}

	function doMath() {
		if (!window.renderMathInElement || !preview) return;
		window.renderMathInElement(preview, {
			delimiters: [
				{ left: '$$', right: '$$', display: true },
				{ left: '$', right: '$', display: false },
				{ left: '\\(', right: '\\)', display: false },
				{ left: '\\[', right: '\\]', display: true }
			],
			throwOnError: false,
			ignoredTags: ['script', 'noscript', 'style', 'textarea', 'pre', 'code', 'option']
		});
	}

	function loadKatex() {
		return new Promise(function (resolve, reject) {
			var base = '/admin/theme/vendor/katex/';
			var style = document.createElement('link');
			style.rel = 'stylesheet';
			style.href = base + 'katex.min.css';
			document.head.appendChild(style);

			var script = document.createElement('script');
			script.src = base + 'katex.min.js';
			script.onload = function () {
				var extra = document.createElement('script');
				extra.src = base + 'contrib/auto-render.min.js';
				extra.onload = resolve;
				extra.onerror = reject;
				document.head.appendChild(extra);
			};
			script.onerror = reject;
			document.head.appendChild(script);
		});
	}

	function insertAtCursor(el, text) {
		var start = el.selectionStart || 0;
		var end = el.selectionEnd || 0;
		var before = el.value.slice(0, start);
		var after = el.value.slice(end);
		el.value = before + text + after;
		el.selectionStart = el.selectionEnd = start + text.length;
		el.dispatchEvent(new Event('input', { bubbles: true }));
		el.focus();
	}

	if (body && preview) {
		body.addEventListener('input', schedulePreview);
		if (body.value.trim() !== '') {
			schedulePreview();
		}
	}

	var bodyImage = document.querySelector('[data-body-image]');
	if (bodyImage && body) {
		bodyImage.addEventListener('change', function () {
			var file = bodyImage.files && bodyImage.files[0];
			if (!file) return;
			var form = new FormData();
			form.append('file', file);
			setStatus('Subiendo imagen…');
			fetch('/admin/upload', {
				method: 'POST',
				headers: { 'X-Admin-CSRF': csrf },
				body: form
			}).then(function (response) {
				if (!response.ok) throw new Error(response.statusText);
				return response.json();
			}).then(function (data) {
				var alt = file.name.replace(/\.[^.]+$/, '');
				insertAtCursor(body, '![' + alt + '](' + data.url + ')\n');
				setStatus('Imagen insertada');
			}).catch(function () {
				setStatus('Error al subir la imagen');
			});
			bodyImage.value = '';
		});
	}

	var toggle = document.querySelector('[data-theme-toggle]');
	if (toggle) {
		toggle.addEventListener('click', function () {
			var root = document.documentElement;
			var next = root.classList.contains('dark') ? 'light' : 'dark';
			root.classList.remove('light', 'dark');
			root.classList.add(next);
			try { localStorage.setItem('theme', next); } catch (err) {}
		});
	}
})();
