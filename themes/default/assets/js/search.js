(function () {
	var index = null;

	function esc(s) {
		return (s || '').replace(/[&<>"]/g, function (c) {
			return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[c];
		});
	}

	function load() {
		if (index) return Promise.resolve(index);
		return fetch('/search-index.json')
			.then(function (r) { return r.json(); })
			.then(function (d) { index = d; return d; });
	}

	function score(doc, terms) {
		var hay = (doc.title + ' ' + (doc.description || '') + ' ' + (doc.tags || []).join(' ') + ' ' + (doc.text || '')).toLowerCase();
		var title = (doc.title || '').toLowerCase();
		var s = 0;
		terms.forEach(function (t) {
			if (!t) return;
			if (title.indexOf(t) >= 0) s += 3;
			else if (hay.indexOf(t) >= 0) s += 1;
		});
		return s;
	}

	function renderInto(results, query) {
		load().then(function (docs) {
			var terms = query.toLowerCase().split(/\s+/).filter(Boolean);
			if (!terms.length) { results.innerHTML = ''; return; }
			var hits = docs
				.map(function (d) { return { d: d, s: score(d, terms) }; })
				.filter(function (x) { return x.s > 0; })
				.sort(function (a, b) { return b.s - a.s; })
				.slice(0, 20);
			if (!hits.length) { results.innerHTML = '<p class="muted">' + ((window.__i18n && window.__i18n.no_results) || 'No results.') + '</p>'; return; }
			results.innerHTML = hits.map(function (x) {
				return '<article class="search-hit"><h2><a href="' + x.d.url + '">' +
					esc(x.d.title) + '</a></h2><p class="muted">' + esc(x.d.description || '') + '</p></article>';
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
		input.addEventListener('input', function () { renderInto(results, input.value); });

		var form = input.closest('form');
		if (form) {
			form.addEventListener('submit', function (e) {
				e.preventDefault();
				var v = input.value;
				history.replaceState(null, '', '/search/' + (v ? '?q=' + encodeURIComponent(v) : ''));
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
