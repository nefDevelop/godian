// Isla pestañas. Con contenido hijo usa los encabezados `##` como pestañas;
// sin contenido, usa props.items = "Uno,Dos,Tres".
export function mount(el, props) {
	var sections = sectionsFromHeadings(el);
	if (sections.length) {
		buildFromSections(el, sections);
		return;
	}
	buildFromItems(el, props);
}

function sectionsFromHeadings(el) {
	var headings = Array.prototype.slice.call(el.children).filter(function (n) {
		return n.tagName === 'H2';
	});
	if (!headings.length) return [];
	return headings.map(function (h) {
		var nodes = [];
		var n = h.nextElementSibling;
		while (n && n.tagName !== 'H2') {
			nodes.push(n);
			n = n.nextElementSibling;
		}
		return { title: headingText(h), nodes: nodes };
	});
}

function headingText(h) {
	var clone = h.cloneNode(true);
	var anchor = clone.querySelector('.anchor');
	if (anchor) anchor.remove();
	return clone.textContent.trim();
}

// kbHandler gestiona la navegación por teclado del tablist (ARIA practices).
function kbHandler(buttons, activate) {
	return function (e) {
		var idx = buttons.indexOf(document.activeElement);
		if (idx < 0) return;
		var next = null;
		if (e.key === 'ArrowRight' || e.key === 'ArrowDown') next = (idx + 1) % buttons.length;
		else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') next = (idx - 1 + buttons.length) % buttons.length;
		else if (e.key === 'Home') next = 0;
		else if (e.key === 'End') next = buttons.length - 1;
		if (next === null) return;
		e.preventDefault();
		buttons[next].focus();
		activate(next);
	};
}

function buildTablist(el, titles, panels) {
	el.innerHTML = '';
	var nav = document.createElement('div');
	nav.className = 'tabs-nav';
	nav.setAttribute('role', 'tablist');

	var container = document.createElement('div');
	container.className = 'tabs-panels';

	var buttons = titles.map(function (title, i) {
		var b = document.createElement('button');
		b.type = 'button';
		b.className = 'tab';
		b.setAttribute('role', 'tab');
		b.id = 'tab-' + 'x' + i + '-' + Math.random().toString(36).slice(2, 7);
		b.textContent = title;
		b.addEventListener('click', function () { activate(i); });
		nav.appendChild(b);
		return b;
	});

	var panelEls = panels.map(function (nodes, i) {
		var p = document.createElement('div');
		p.className = 'tab-panel';
		p.setAttribute('role', 'tabpanel');
		p.setAttribute('aria-labelledby', buttons[i].id);
		p.tabIndex = 0;
		nodes.forEach(function (n) { p.appendChild(n); });
		container.appendChild(p);
		return p;
	});

	el.appendChild(nav);
	el.appendChild(container);

	function activate(i) {
		buttons.forEach(function (b, j) {
			var on = j === i;
			b.classList.toggle('is-active', on);
			b.setAttribute('aria-selected', on ? 'true' : 'false');
			b.tabIndex = on ? 0 : -1;
		});
		panelEls.forEach(function (p, j) { p.hidden = j !== i; });
	}
	nav.addEventListener('keydown', kbHandler(buttons, activate));
	activate(0);
}

function buildFromSections(el, sections) {
	buildTablist(el, sections.map(function (s) { return s.title; }),
		sections.map(function (s) { return s.nodes; }));
}

function buildFromItems(el, props) {
	var items = (props.items || '')
		.split(',')
		.map(function (s) { return s.trim(); })
		.filter(Boolean);
	if (!items.length) return;
	// Contenido = la etiqueta (comportamiento previo), como nodo de texto.
	buildTablist(el, items, items.map(function (label) {
		var p = document.createElement('p');
		p.textContent = label;
		return [p];
	}));
}
