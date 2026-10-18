// Isla valoración por estrellas: props.max (por defecto 5), props.value.
export function mount(el, props) {
	var max = Number(props.max || 5);
	var current = Number(props.value || 0);
	el.classList.add('rating');
	el.innerHTML = '';

	var stars = [];
	for (var i = 1; i <= max; i++) {
		(function (n) {
			var s = document.createElement('button');
			s.type = 'button';
			s.className = 'star';
			s.textContent = '★';
			s.setAttribute('aria-label', n + '/' + max);
			s.setAttribute('aria-pressed', 'false');
			s.addEventListener('click', function () { current = n; paint(); });
			stars.push(s);
			el.appendChild(s);
		})(i);
	}

	function paint() {
		stars.forEach(function (s, idx) {
			var on = idx < current;
			s.classList.toggle('is-on', on);
			s.setAttribute('aria-pressed', on ? 'true' : 'false');
		});
	}
	paint();
}
