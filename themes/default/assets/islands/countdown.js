// Isla cuenta atrás: props.to (fecha ISO), props.label.
export function mount(el, props) {
	var target = new Date(props.to || Date.now());
	if (isNaN(target.getTime())) {
		el.textContent = '—';
		return;
	}
	el.classList.add('countdown');
	function tick() {
		var diff = target - new Date();
		if (diff <= 0) {
			el.textContent = props.label || 'Done';
			return;
		}
		var s = Math.floor(diff / 1000);
		var d = Math.floor(s / 86400);
		var h = Math.floor((s % 86400) / 3600);
		var m = Math.floor((s % 3600) / 60);
		var sec = s % 60;
		el.textContent = (d ? d + 'd ' : '') + h + 'h ' + m + 'm ' + sec + 's';
	}
	tick();
	setInterval(tick, 1000);
}
