// Isla reloj: muestra la hora local actualizándose cada segundo.
export function mount(el, props) {
	el.classList.add('clock');
	function tick() {
		el.textContent = new Date().toLocaleTimeString();
	}
	tick();
	setInterval(tick, 1000);
}
