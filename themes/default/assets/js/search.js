(function () {
	var index = null;
	var BASE = window.__BASE_PATH__ || '';

	function esc(s) {
		return (s || '').replace(/[&<>"]/g, function (c) {
			return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[c];
		});
	}

	function norm(s) {
		s = (s || '').toLowerCase();
		if (s.normalize) s = s.normalize('NFD').replace(/[\u0300-\u036f]/g, '');
		return s;
	}

	function escapeRe(s) {
		return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
	}

	function termsOf(query) {
		return norm(query).split(/\s+/).filter(function (t) { return t.length > 1; });
	}

	function rawTermsOf(query) {
		return (query || '').trim().split(/\s+/).filter(function (t) { return t.length > 1; });
	}

	// hasWord evita coincidencias parciales (p. ej. "go" dentro de "golang").
	function hasWord(hay, term) {
		if (!term) return false;
		var re = new RegExp('(^|[^a-z0-9])' + escapeRe(term) + '([^a-z0-9]|$)', 'i');
		return re.test(hay);
	}

	function score(doc, terms, rawQuery) {
		var title = norm(doc.title);
		var desc = norm(doc.description);
		var tags = norm((doc.tags || []).join(' '));
		var text = norm(doc.text);
		var total = 0;

		for (var i = 0; i < terms.length; i++) {
			var t = terms[i];
			var hit = false;
			if (hasWord(title, t)) { total += 12; hit = true; }
			else if (title.indexOf(t) >= 0) { total += 8; hit = true; }
			if (hasWord(desc, t)) { total += 5; hit = true; }
			if (hasWord(tags, t)) { total += 3; hit = true; }
			if (hasWord(text, t)) { total += 1; hit = true; }
			if (!hit) return 0; // todos los términos deben aparecer
		}

		var q = norm(rawQuery);
		if (q && title === q) total += 20;
		else if (q && title.indexOf(q) === 0) total += 8;
		return total;
	}

	function load() {
		if (index) return Promise.resolve(index);
		return fetch(BASE + '/search-index.json')
			.then(function (r) { return r.json(); })
			.then(function (d) { index = d; return d; });
	}

	function highlight(text, terms) {
		var escaped = esc(text);
		if (!terms.length) return escaped;
		var re = new RegExp('(' + terms.map(escapeRe).join('|') + ')', 'gi');
		return escaped.replace(re, '<mark>$1</mark>');
	}

	function snippet(doc, terms) {
		var text = doc.text || doc.description || '';
		if (!text) return '';
		var lower = text.toLowerCase();
		var idx = -1;
		for (var i = 0; i < terms.length; i++) {
			var p = lower.indexOf(terms[i].toLowerCase());
			if (p >= 0 && (idx < 0 || p < idx)) idx = p;
		}
		if (idx < 0) return '';
		var start = Math.max(0, idx - 60);
		var end = Math.min(text.length, idx + 140);
		var out = (start > 0 ? '…' : '') + text.slice(start, end) + (end < text.length ? '…' : '');
		return highlight(out, terms);
	}

	function renderInto(results, query) {
		load().then(function (docs) {
			var terms = termsOf(query);
			var rawTerms = rawTermsOf(query);
			if (!terms.length) { results.innerHTML = ''; return; }

			var all = docs.map(function (d) {
				return { d: d, s: score(d, terms, query) };
			}).filter(function (x) {
				return x.s > 0;
			}).sort(function (a, b) {
				return b.s - a.s || (a.d.title || '').localeCompare(b.d.title || '');
			});

			// Primero coincidencias en título/descripción/tags; si no hay,
			// se cae al cuerpo para no dejar búsquedas legítimas sin resultados.
			var strong = all.filter(function (x) { return x.s >= 3; });
			var hits = (strong.length ? strong : all).slice(0, 10);

			if (!hits.length) {
				results.innerHTML = '<p class="muted">' + ((window.__i18n && window.__i18n.no_results) || 'No results.') + '</p>';
				return;
			}

			results.innerHTML = hits.map(function (x) {
				var d = x.d;
				var type = d.type === 'doc' ? 'DOC' : 'POST';
				var desc = d.description ? '<p class="muted search-desc">' + esc(d.description) + '</p>' : '';
				var ex = snippet(d, rawTerms);
				var excerpt = ex ? '<p class="search-excerpt">' + ex + '</p>' : '';
				return '<article class="search-hit">' +
					'<div class="search-hit-head"><span class="search-type">' + type + '</span>' +
					'<h2><a href="' + esc(d.url) + '">' + esc(d.title) + '</a></h2></div>' +
					desc + excerpt +
					'</article>';
			}).join('');
		});
	}

	// Enlaza todos los buscadores (página y modal).
	document.querySelectorAll('[data-search-scope]').forEach(function (scope) {
		var input = scope.querySelector('[data-search-input]');
		var results = scope.querySelector('[data-search-results]');
		if (!input || !results) return;

		if (scope.id !== 'search-modal') {
			var q = new URLSearchParams(location.search).get('q');
			if (q) { input.value = q; renderInto(results, q); }
		}

		var timer;
		input.addEventListener('input', function () {
			clearTimeout(timer);
			timer = setTimeout(function () { renderInto(results, input.value); }, 120);
		});

		var form = input.closest('form');
		if (form) {
			form.addEventListener('submit', function (e) {
				e.preventDefault();
				var v = input.value;
				history.replaceState(null, '', BASE + '/search/' + (v ? '?q=' + encodeURIComponent(v) : ''));
				renderInto(results, v);
			});
		}
	});

	// Modal Ctrl+K.
	var modal = document.getElementById('search-modal');
	if (!modal) return;

	function open() {
		modal.hidden = false;
		document.body.classList.add('search-open');
		var i = modal.querySelector('[data-search-input]');
		if (i) i.focus();
	}
	function close() {
		modal.hidden = true;
		document.body.classList.remove('search-open');
	}

	document.querySelectorAll('.nav-search').forEach(function (a) {
		a.addEventListener('click', function (e) { e.preventDefault(); open(); });
	});
	modal.addEventListener('click', function (e) {
		if (e.target.closest('[data-search-close]')) close();
	});
	document.addEventListener('keydown', function (e) {
		if ((e.key === 'k' || e.key === 'K') && (e.metaKey || e.ctrlKey)) {
			e.preventDefault();
			open();
		} else if (e.key === 'Escape' && !modal.hidden) {
			close();
		}
	});
})();
