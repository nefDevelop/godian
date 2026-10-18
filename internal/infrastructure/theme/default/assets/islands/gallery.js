// Isla galería: rejilla de imágenes con lightbox (PhotoSwipe vendorizado).
// props.images = "url1,url2,…"
export function mount(el, props) {
	var imgs = (props.images || '')
		.split(',')
		.map(function (s) { return s.trim(); })
		.filter(Boolean);
	if (!imgs.length) return;

	el.classList.add('gallery');
	el.innerHTML = '';

	var links = imgs.map(function (src, idx) {
		var a = document.createElement('a');
		a.className = 'gallery-item';
		a.href = src;
		// Nombre accesible del enlace (la imagen es decorativa).
		a.setAttribute('aria-label', 'Ampliar imagen ' + (idx + 1) + ' de ' + imgs.length);
		var img = document.createElement('img');
		img.src = src;
		img.loading = 'lazy';
		img.decoding = 'async';
		img.alt = '';
		a.appendChild(img);
		el.appendChild(a);
		return a;
	});

	import('/assets/vendor/photoswipe/photoswipe.esm.min.js').then(function (mod) {
		var PhotoSwipe = mod.default || mod.PhotoSwipe;
		var link = document.createElement('link');
		link.rel = 'stylesheet';
		link.href = '/assets/vendor/photoswipe/photoswipe.css';
		document.head.appendChild(link);

		el.addEventListener('click', function (e) {
			var a = e.target.closest('.gallery-item');
			if (!a) return;
			e.preventDefault();
			var items = imgs.map(function (src) { return { src: src, w: 1600, h: 1000 }; });
			var idx = links.indexOf(a);
			var pswp = new PhotoSwipe({ dataSource: items, index: idx < 0 ? 0 : idx });
			pswp.init();
		});
	}).catch(function () {});
}
