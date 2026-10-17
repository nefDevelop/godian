// Isla de acordeón: props.items = "Título=Contenido;Otro=Cuerpo".
export function mount(el, props) {
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
