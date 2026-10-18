// Isla compartir: props.url (por defecto, la URL actual), props.title.
export function mount(el, props) {
	var url = props.url || location.href;
	var title = props.title || document.title;
	var enc = encodeURIComponent;

	var links = [
		['X', 'https://twitter.com/intent/tweet?url=' + enc(url) + '&text=' + enc(title)],
		['Telegram', 'https://t.me/share/url?url=' + enc(url) + '&text=' + enc(title)],
		['LinkedIn', 'https://www.linkedin.com/sharing/share-offsite/?url=' + enc(url)],
		['Email', 'mailto:?subject=' + enc(title) + '&body=' + enc(url)]
	];

	el.classList.add('share-island');
	el.innerHTML = '';

	links.forEach(function (l) {
		var a = document.createElement('a');
		a.href = l[1];
		a.textContent = l[0];
		a.rel = 'noopener';
		a.target = '_blank';
		el.appendChild(a);
	});

	var btn = document.createElement('button');
	btn.type = 'button';
	btn.className = 'share-copy';
	btn.textContent = (window.__i18n && window.__i18n.copy_link) || 'Copy link';
	btn.addEventListener('click', function () {
		var done = function () {
			var old = btn.textContent;
			btn.textContent = (window.__i18n && window.__i18n.copied) || 'Copied!';
			setTimeout(function () { btn.textContent = old; }, 1500);
		};
		if (navigator.clipboard) navigator.clipboard.writeText(url).then(done, done);
	});
	el.appendChild(btn);
}
