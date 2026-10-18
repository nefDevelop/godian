// Isla lightbox: imagen individual con zoom (PhotoSwipe).
export function mount(el, props) {
  var src = props.src || '';
  var alt = props.alt || '';
  var caption = props.caption || '';
  el.classList.add('lightbox');
  el.innerHTML = '';
  var img = document.createElement('img');
  img.src = src; img.alt = alt; img.loading = 'lazy'; img.decoding = 'async';
  img.style.cursor = 'zoom-in';
  el.appendChild(img);
  if (caption) { var c = document.createElement('p'); c.className = 'lightbox-caption'; c.textContent = caption; el.appendChild(c); }
  img.addEventListener('click', function () {
    import('/assets/vendor/photoswipe/photoswipe.esm.min.js').then(function (mod) {
      var PhotoSwipe = mod.default || mod.PhotoSwipe;
      var link = document.createElement('link');
      link.rel = 'stylesheet';
      link.href = '/assets/vendor/photoswipe/photoswipe.css';
      if (!document.querySelector('link[href="/assets/vendor/photoswipe/photoswipe.css"]')) document.head.appendChild(link);
      var items = [{ src: src, w: img.naturalWidth || 1600, h: img.naturalHeight || 1000, alt: alt }];
      var pswp = new PhotoSwipe({ dataSource: items, index: 0 });
      pswp.init();
    }).catch(function () {});
  });
}
