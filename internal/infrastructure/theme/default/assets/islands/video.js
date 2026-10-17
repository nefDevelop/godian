// Isla vídeo: carga el iframe de YouTube al pulsar (privacidad + rendimiento).
export function mount(el, props) {
	var id = el.getAttribute('data-video-id') || props.id;
	var btn = el.querySelector('.video-play');
	if (!btn || !id) return;
	btn.addEventListener('click', function () {
		var iframe = document.createElement('iframe');
		iframe.src = 'https://www.youtube.com/embed/' + id + '?autoplay=1';
		iframe.title = props.title || 'Video';
		iframe.allow = 'accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture';
		iframe.allowFullscreen = true;
		iframe.className = 'video-iframe';
		el.innerHTML = '';
		el.appendChild(iframe);
	});
}
