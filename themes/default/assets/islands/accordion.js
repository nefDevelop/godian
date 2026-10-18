// Isla acordeón. Con contenido hijo usa los encabezados `##` como secciones;
// sin contenido, usa props.items = "Título=Cuerpo;Otro=Cuerpo".
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
	sections.forEach(function (s) {
		var d = document.createElement('details');
		d.className = 'accordion-item';
		var sum = document.createElement('summary');
		sum.textContent = s.title;
		var body = document.createElement('div');
		body.className = 'accordion-body';
		s.nodes.forEach(function (n) { body.appendChild(n); });
		d.appendChild(sum);
		d.appendChild(body);
		el.appendChild(d);
	});
}

function buildFromItems(el, props) {
	var items = (props.items || '')
		.split(';')
		.map(function (s) {
			var i = s.indexOf('=');
			if (i < 0) return { title: s.trim(), body: '' };
			return { title: s.slice(0, i).trim(), body: s.slice(i + 1).trim() };
		})
		.filter(function (x) { return x.title; });
	if (!items.length) return;

	el.innerHTML = '';
	items.forEach(function (it) {
		var d = document.createElement('details');
		d.className = 'accordion-item';
		var s = document.createElement('summary');
		s.textContent = it.title;
		var p = document.createElement('div');
		p.className = 'accordion-body';
		p.textContent = it.body;
		d.appendChild(s);
		d.appendChild(p);
		el.appendChild(d);
	});
}
