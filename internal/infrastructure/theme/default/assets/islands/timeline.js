// Isla timeline: línea de tiempo vertical.
// props.items = "2026-01:Lanzamiento,2026-02:Nuevo tema" (fecha|texto o fecha:texto)
export function mount(el, props) {
	var items = (props.items || '').split(',').map(function (s) { return s.trim(); }).filter(Boolean);
	el.classList.add('timeline');
	el.innerHTML = '';
	var ol = document.createElement('ol');
	ol.className = 'timeline-list';
	items.forEach(function (it) {
		var sep = it.indexOf('|') >= 0 ? '|' : ':';
		var i = it.indexOf(sep);
		var when = i >= 0 ? it.slice(0, i).trim() : '';
		var text = i >= 0 ? it.slice(i + 1).trim() : it;
		var li = document.createElement('li');
		li.className = 'timeline-item';
		if (when) {
			var time = document.createElement('time');
			time.className = 'timeline-date';
			time.textContent = when;
			li.appendChild(time);
		}
		var span = document.createElement('span');
		span.className = 'timeline-text';
		span.textContent = text;
		li.appendChild(span);
		ol.appendChild(li);
	});
	el.appendChild(ol);
}
