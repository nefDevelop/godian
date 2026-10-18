// Isla calendar: calendario mensual con eventos y fechas de contenido.
// - props.events = "5:Reunión,12:Entrega" (eventos manuales del mes mostrado)
// - props.source = "site" (usa /content-dates.json) o URL a un JSON [{date,title,url}]
export function mount(el, props) {
	var now = new Date();
	var year = Number(props.year || now.getFullYear());
	var month = Number(props.month || (now.getMonth() + 1)) - 1;
	el.classList.add('calendar');
	var manual = parseManual(props.events, year, month);

	function draw(byDate) { render(el, year, month, byDate); }

	if (props.source) {
		fetchSource(props.source, function (items) {
			var byDate = {};
			items.forEach(function (it) {
				if (!it || !it.date) return;
				(byDate[it.date] = byDate[it.date] || []).push({ label: it.title || it.date, url: it.url });
			});
			Object.keys(manual).forEach(function (k) {
				byDate[k] = (byDate[k] || []).concat(manual[k]);
			});
			draw(byDate);
		}, function () { draw(manual); });
	} else {
		draw(manual);
	}
}

function pad(n) { return (n < 10 ? '0' : '') + n; }

// parseManual convierte "5:Reunión,12:Entrega" en { "YYYY-MM-DD": [{label}] }.
function parseManual(spec, year, month) {
	var out = {};
	(spec || '').split(',').forEach(function (e) {
		var i = e.indexOf(':');
		if (i < 0) return;
		var day = Number(e.slice(0, i).trim());
		var label = e.slice(i + 1).trim();
		if (!day || !label) return;
		var key = year + '-' + pad(month + 1) + '-' + pad(day);
		(out[key] = out[key] || []).push({ label: label });
	});
	return out;
}

// fetchSource obtiene las fechas de contenido (source="site" → content-dates.json).
function fetchSource(source, ok, fail) {
	var url = source === 'site' ? '/content-dates.json' : source;
	fetch(url).then(function (r) { return r.ok ? r.json() : []; }).then(ok).catch(fail);
}

function render(el, year, month, byDate) {
	el.textContent = '';
	var first = new Date(year, month, 1);
	var last = new Date(year, month + 1, 0);
	var startDay = first.getDay();
	var days = last.getDate();
	var locale = document.documentElement.lang || 'es';
	var dow = ['Dom', 'Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb'];

	var head = document.createElement('div');
	head.className = 'calendar-head';
	head.textContent = first.toLocaleString(locale, { month: 'long', year: 'numeric' });
	el.appendChild(head);

	var grid = document.createElement('div');
	grid.className = 'calendar-grid';
	dow.forEach(function (n) {
		var c = document.createElement('div');
		c.className = 'calendar-dow';
		c.textContent = n;
		grid.appendChild(c);
	});
	for (var i = 0; i < startDay; i++) {
		var empty = document.createElement('div');
		empty.className = 'calendar-cell empty';
		grid.appendChild(empty);
	}

	var panel = document.createElement('div');
	panel.className = 'calendar-panel';
	panel.hidden = true;

	for (var d = 1; d <= days; d++) {
		var key = year + '-' + pad(month + 1) + '-' + pad(d);
		var items = byDate[key] || [];
		var cell = document.createElement(items.length ? 'button' : 'div');
		cell.className = 'calendar-cell' + (items.length ? ' has-event' : '');
		if (items.length) {
			cell.type = 'button';
			cell.setAttribute('aria-label', d + ' — ' + items.map(function (x) { return x.label; }).join(', '));
			cell.addEventListener('click', (function (its) {
				return function () { showPanel(panel, its); };
			})(items));
		}
		cell.appendChild(document.createTextNode(String(d)));
		if (items.length) {
			var dot = document.createElement('span');
			dot.className = 'calendar-dot';
			cell.appendChild(dot);
		}
		grid.appendChild(cell);
	}

	el.appendChild(grid);
	el.appendChild(panel);
}

function showPanel(panel, items) {
	panel.textContent = '';
	panel.hidden = false;
	var ul = document.createElement('ul');
	items.forEach(function (it) {
		var li = document.createElement('li');
		if (it.url) {
			var a = document.createElement('a');
			a.href = it.url;
			a.textContent = it.label;
			li.appendChild(a);
		} else {
			li.textContent = it.label;
		}
		ul.appendChild(li);
	});
	panel.appendChild(ul);
}
