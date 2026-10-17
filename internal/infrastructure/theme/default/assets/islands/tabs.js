// Isla de pestañas: props.items = "Uno,Dos,Tres".
export function mount(el, props) {
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
