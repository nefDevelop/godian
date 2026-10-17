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

function buildFromSections(el, sections) {
	el.innerHTML = '';
	var nav = document.createElement('div');
	nav.className = 'tabs-nav';
	var panels = sections.map(function (s) {
		var p = document.createElement('div');
		p.className = 'tab-panel';
		s.nodes.forEach(function (n) { p.appendChild(n); });
		return p;
	});
	sections.forEach(function (s, i) {
		var b = document.createElement('button');
		b.type = 'button';
		b.className = 'tab';
		b.textContent = s.title;
		b.addEventListener('click', function () { activate(i); });
		nav.appendChild(b);
	});
	var container = document.createElement('div');
	container.className = 'tabs-panels';
	panels.forEach(function (p) { container.appendChild(p); });
	el.appendChild(nav);
	el.appendChild(container);

	function activate(i) {
		Array.prototype.forEach.call(nav.children, function (b, j) {
			b.classList.toggle('is-active', j === i);
			b.setAttribute('aria-selected', j === i ? 'true' : 'false');
		});
		panels.forEach(function (p, j) { p.hidden = j !== i; });
	}
	activate(0);
}

function buildFromItems(el, props) {
	var items = (props.items || '')
		.split(',')
		.map(function (s) { return s.trim(); })
		.filter(Boolean);
	if (!items.length) return;

	el.innerHTML = '';
	var nav = document.createElement('div');
	nav.className = 'tabs-nav';
	var panel = document.createElement('div');
	panel.className = 'tabs-panel';

	items.forEach(function (label, i) {
		var b = document.createElement('button');
		b.type = 'button';
		b.className = 'tab';
		b.textContent = label;
		b.addEventListener('click', function () { activate(i); });
		nav.appendChild(b);
	});
	el.appendChild(nav);
	el.appendChild(panel);

	function activate(i) {
		Array.prototype.forEach.call(nav.children, function (b, j) {
			b.classList.toggle('is-active', j === i);
			b.setAttribute('aria-selected', j === i ? 'true' : 'false');
		});
		panel.textContent = items[i];
	}
	activate(0);
}
