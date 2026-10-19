// Isla selector de color de acento: props.colors = "#a,#b,#c" (opcional).
export function mount(el, props) {
	var colors = (props.colors || '#2563eb,#16a34a,#db2777,#f59e0b,#7c3aed')
		.split(',')
		.map(function (s) { return s.trim(); })
		.filter(Boolean);

	el.classList.add('theme-picker');
	el.innerHTML = '';
	colors.forEach(function (c) {
		var b = document.createElement('button');
		b.type = 'button';
		b.className = 'swatch';
		b.style.background = c;
		b.setAttribute('aria-label', c);
		b.addEventListener('click', function () {
			var apply = function () {
				document.documentElement.style.setProperty('--accent', c);
			};
			if (window.godianThemeTransition) window.godianThemeTransition(apply);
			else apply();
			try { localStorage.setItem('accent', c); } catch (e) {}
		});
		el.appendChild(b);
	});

	try {
		var saved = localStorage.getItem('accent');
		if (saved) document.documentElement.style.setProperty('--accent', saved);
	} catch (e) {}
}
