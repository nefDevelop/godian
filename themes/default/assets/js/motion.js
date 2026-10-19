// Utilidades de movimiento compartidas por las islas del tema.
// La decisión de animar siempre respeta prefers-reduced-motion.

export function prefersReducedMotion() {
	return typeof window !== 'undefined' &&
		typeof window.matchMedia === 'function' &&
		window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

// motionValue lee un token CSS (por ejemplo --motion-base) con fallback.
export function motionValue(name, fallback) {
	if (typeof window === 'undefined' || typeof getComputedStyle !== 'function') {
		return fallback;
	}
	var value = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
	return value || fallback;
}

// motionMs devuelve la duración en milisegundos para la Web Animations API.
export function motionMs(name, fallback) {
	var value = motionValue(name, '');
	var n = parseFloat(value);
	if (!isNaN(n)) {
		return value.indexOf('ms') >= 0 ? n : n * 1000;
	}
	return fallback;
}

// animate es un envoltorio de Element.animate que no hace nada si el usuario
// prefiere movimiento reducido o el navegador no soporta la API.
export function animate(el, keyframes, options) {
	if (!el || typeof el.animate !== 'function' || prefersReducedMotion()) {
		return null;
	}
	return el.animate(keyframes, options);
}
