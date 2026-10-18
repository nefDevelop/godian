// Isla vídeo: carga el iframe de YouTube al pulsar (privacidad + rendimiento).
export function mount(el, props) {
	var id = el.getAttribute('data-video-id') || props.id;
	// El wrapper de la isla no está posicionado: el iframe debe ir dentro de ".video"
	// (que sí tiene position: relative), o se posicionaría respecto al viewport (arriba de todo).
	var box = el.querySelector('.video') || el;
	var btn = el.querySelector('.video-play');
	if (!btn || !id) return;
	btn.addEventListener('click', function (e) {
		if (e) { e.preventDefault(); e.stopPropagation(); }
		var iframe = document.createElement('iframe');
		iframe.src = 'https://www.youtube.com/embed/' + id + '?autoplay=1&mute=1&rel=0&playsinline=1';
		iframe.title = props.title || 'Video';
		// Sin "autoplay" en allow (Firefox no lo reconoce y genera warnings de Feature Policy).
		iframe.allow = 'fullscreen; picture-in-picture';
		iframe.allowFullscreen = true;
		iframe.className = 'video-iframe';
		iframe.setAttribute('tabindex', '-1');
		box.appendChild(iframe);
		btn.style.display = 'none';
	});
}
