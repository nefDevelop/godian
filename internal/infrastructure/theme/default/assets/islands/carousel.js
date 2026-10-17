// Isla carrusel: props.images = "url1,url2,url3".
export function mount(el, props) {
	var imgs = (props.images || '')
		.split(',')
		.map(function (s) { return s.trim(); })
		.filter(Boolean);
	if (!imgs.length) return;

	var i = 0;
	el.innerHTML = '';

	var wrap = document.createElement('div');
	wrap.className = 'carousel';

	var prev = document.createElement('button');
	prev.type = 'button';
	prev.className = 'carousel-nav';
	prev.setAttribute('aria-label', 'Anterior');
	prev.textContent = '‹';

	var img = document.createElement('img');
	img.className = 'carousel-img';
	img.loading = 'lazy';
	img.decoding = 'async';

	var next = document.createElement('button');
	next.type = 'button';
	next.className = 'carousel-nav';
	next.setAttribute('aria-label', 'Siguiente');
	next.textContent = '›';

	function render() { img.src = imgs[i]; }
	prev.addEventListener('click', function () { i = (i - 1 + imgs.length) % imgs.length; render(); });
	next.addEventListener('click', function () { i = (i + 1) % imgs.length; render(); });

	wrap.appendChild(prev);
	wrap.appendChild(img);
	wrap.appendChild(next);
	el.appendChild(wrap);
	render();
}
