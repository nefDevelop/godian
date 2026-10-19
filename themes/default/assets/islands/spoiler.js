import { prefersReducedMotion, motionMs, motionValue } from '/assets/js/motion.js';

// Isla spoiler: oculta el texto hasta pulsar el botón.
export function mount(el, props) {
	var btn = el.querySelector('.spoiler-btn');
	var body = el.querySelector('.spoiler-body');
	if (!btn || !body) return;
	var anim = null;

	btn.addEventListener('click', function () {
		if (anim) return;

		var opening = body.hidden;
		btn.setAttribute('aria-expanded', opening ? 'true' : 'false');
		el.classList.toggle('is-open', opening);

		if (prefersReducedMotion() || typeof body.animate !== 'function') {
			body.hidden = !opening;
			return;
		}

		var height;
		body.style.overflow = 'hidden';
		if (opening) {
			body.hidden = false;
			height = body.scrollHeight;
			anim = body.animate(
				[{ height: '0px', opacity: 0 }, { height: height + 'px', opacity: 1 }],
				{
					duration: motionMs('--motion-base', 220),
					easing: motionValue('--motion-ease-enter', 'cubic-bezier(0, 0, 0.2, 1)')
				}
			);
		} else {
			height = body.scrollHeight;
			anim = body.animate(
				[{ height: height + 'px', opacity: 1 }, { height: '0px', opacity: 0 }],
				{
					duration: motionMs('--motion-fast', 140),
					easing: motionValue('--motion-ease-exit', 'cubic-bezier(0.4, 0, 1, 1)')
				}
			);
		}

		anim.onfinish = function () {
			if (!opening) body.hidden = true;
			body.style.overflow = '';
			body.style.height = '';
			anim = null;
		};
		anim.oncancel = function () {
			body.style.overflow = '';
			body.style.height = '';
			anim = null;
		};
	});
}
