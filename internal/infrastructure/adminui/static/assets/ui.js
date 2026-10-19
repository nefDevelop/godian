const ICONS = {
	dashboard: '<rect x="3" y="3" width="7" height="9" rx="1"/><rect x="14" y="3" width="7" height="5" rx="1"/><rect x="14" y="12" width="7" height="9" rx="1"/><rect x="3" y="16" width="7" height="5" rx="1"/>',
	plus: '<path d="M12 5v14M5 12h14"/>',
	search: '<circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/>',
	edit: '<path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z"/>',
	trash: '<path d="M3 6h18M8 6V4h8v2M19 6l-1 14H6L5 6M10 11v6M14 11v6"/>',
	copy: '<rect x="9" y="9" width="11" height="11" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>',
	external: '<path d="M15 3h6v6M10 14 21 3M21 14v6a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1h6"/>',
	image: '<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="m21 15-5-5L5 21"/>',
	save: '<path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2Z"/><path d="M7 3v6h10V3M7 21v-7h10v7"/>',
	logout: '<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9"/>',
	refresh: '<path d="M20 11a8 8 0 1 0-2.3 5.7M20 5v6h-6"/>',
	x: '<path d="M18 6 6 18M6 6l12 12"/>',
	upload: '<path d="M12 16V4M7 9l5-5 5 5M5 20h14"/>',
	file: '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8Z"/><path d="M14 2v6h6"/>',
	folder: '<path d="M3 7a2 2 0 0 1 2-2h5l2 2h7a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2Z"/>',
	link: '<path d="M10 13a5 5 0 0 0 7 0l3-3a5 5 0 0 0-7-7l-1 1"/><path d="M14 11a5 5 0 0 0-7 0l-3 3a5 5 0 0 0 7 7l1-1"/>',
	bold: '<path d="M6 4h8a4 4 0 0 1 0 8H6Z"/><path d="M6 12h9a4 4 0 0 1 0 8H6Z"/>',
	italic: '<path d="M19 4h-9M14 20H5M15 4 9 20"/>',
	code: '<path d="m8 6-6 6 6 6M16 6l6 6-6 6"/>',
	quote: '<path d="M3 21c3-2 4-4 4-7V7H3v7h4M14 21c3-2 4-4 4-7V7h-4v7h4"/>',
	list: '<path d="M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01"/>',
	heading: '<path d="M6 4v16M18 4v16M6 12h12"/>',
	duplicate: '<rect x="9" y="9" width="11" height="11" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>',
	rename: '<path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z"/>',
	moon: '<path d="M21 12.8A9 9 0 1 1 11.2 3 7 7 0 0 0 21 12.8Z"/>',
	chevron: '<path d="m9 18 6-6-6-6"/>',
	arrow: '<path d="M5 12h14M13 6l6 6-6 6"/>',
	menu: '<path d="M4 6h16M4 12h16M4 18h16"/>',
	check: '<path d="M20 6 9 17l-5-5"/>',
};

export function icon(name, label) {
	const span = document.createElement('span');
	span.className = 'admin-icon';
	span.setAttribute('aria-hidden', 'true');
	span.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">' + (ICONS[name] || ICONS.file) + '</svg>';
	if (label) span.setAttribute('title', label);
	return span;
}

export function h(tag, attrs = {}, ...children) {
	const node = document.createElement(tag);
	for (const [key, value] of Object.entries(attrs || {})) {
		if (value === undefined || value === null || value === false) continue;
		if (key === 'class' || key === 'className') {
			node.className = value;
		} else if (key === 'dataset') {
			for (const [k, v] of Object.entries(value)) node.dataset[k] = v;
		} else if (key === 'style' && typeof value === 'object') {
			for (const [k, v] of Object.entries(value)) node.style.setProperty(k, v);
		} else if (key === 'html') {
			node.innerHTML = value;
		} else if (key === 'text') {
			node.textContent = value;
		} else if (key.startsWith('on') && typeof value === 'function') {
			node.addEventListener(key.slice(2).toLowerCase(), value);
		} else if (key === 'value') {
			node.value = value;
		} else if (key === 'checked') {
			node.checked = Boolean(value);
		} else if (key === 'selected') {
			node.selected = Boolean(value);
		} else {
			node.setAttribute(key, String(value));
		}
	}
	append(node, children);
	return node;
}

function append(node, children) {
	for (const child of children) {
		if (child === undefined || child === null || child === false) continue;
		if (Array.isArray(child)) {
			append(node, child);
		} else if (child instanceof Node) {
			node.appendChild(child);
		} else {
			node.appendChild(document.createTextNode(String(child)));
		}
	}
}

export function clear(node) {
	node.replaceChildren();
}

export function toast(message, kind = 'info', timeout = 3800) {
	let box = document.querySelector('.admin-toasts');
	if (!box) {
		box = h('div', { class: 'admin-toasts', role: 'status', 'aria-live': 'polite' });
		document.body.appendChild(box);
	}
	const item = h('div', { class: 'admin-toast ' + kind }, message);
	box.appendChild(item);
	setTimeout(() => item.remove(), timeout);
}

export function dialog({ title, body, actions = [], width }) {
	const dlg = h('dialog', { class: 'admin-modal' });
	if (width) dlg.style.width = width;
	const closeBtn = h('button', { type: 'button', class: 'admin-icon-button', 'aria-label': 'Cerrar', onclick: () => dlg.close() }, icon('x'));
	const head = h('div', { class: 'admin-modal-head' }, h('h2', {}, title || ''), closeBtn);
	const bodyBox = h('div', { class: 'admin-modal-body' });
	if (body instanceof Node) bodyBox.appendChild(body);
	else if (typeof body === 'string') bodyBox.textContent = body;
	const foot = h('div', { class: 'admin-modal-actions' });
	for (const action of actions) {
		const btn = h('button', {
			type: 'button',
			class: 'admin-button' + (action.kind === 'primary' ? ' admin-button-primary' : '') + (action.kind === 'danger' ? ' admin-button-danger' : ''),
			onclick: async () => {
				const keep = action.onClick ? await action.onClick() : false;
				if (!keep) dlg.close();
			},
		}, action.label);
		foot.appendChild(btn);
	}
	dlg.append(head, bodyBox);
	if (actions.length) dlg.appendChild(foot);
	document.body.appendChild(dlg);
	dlg.addEventListener('close', () => dlg.remove(), { once: true });
	dlg.showModal();
	return { dialog: dlg, body: bodyBox, close: () => dlg.close() };
}

export function confirmDialog(message, options = {}) {
	const title = options.title || 'Confirmar';
	const confirmLabel = options.confirm || 'Aceptar';
	const cancelLabel = options.cancel || 'Cancelar';
	return new Promise((resolve) => {
		const dlg = h('dialog', { class: 'admin-modal' });
		const closeBtn = h('button', { type: 'button', class: 'admin-icon-button', 'aria-label': 'Cerrar', onclick: () => dlg.close() }, icon('x'));
		const head = h('div', { class: 'admin-modal-head' }, h('h2', {}, title), closeBtn);
		const body = h('div', { class: 'admin-modal-body' }, message);
		const foot = h('div', { class: 'admin-modal-actions' });
		let result = false;
		foot.append(
			h('button', { type: 'button', class: 'admin-button', onclick: () => dlg.close() }, cancelLabel),
			h('button', {
				type: 'button',
				class: 'admin-button admin-button-primary' + (options.danger ? ' admin-button-danger' : ''),
				onclick: () => { result = true; dlg.close(); },
			}, confirmLabel),
		);
		dlg.append(head, body, foot);
		document.body.appendChild(dlg);
		dlg.addEventListener('close', () => { dlg.remove(); resolve(result); }, { once: true });
		dlg.showModal();
	});
}

export function promptDialog(title, label, value = '', options = {}) {
	return new Promise((resolve) => {
		const input = h('input', { type: 'text', value, placeholder: options.placeholder || '' });
		const body = h('div', { class: 'admin-field' }, h('label', {}, label || ''), input);
		const dlg = h('dialog', { class: 'admin-modal' });
		const closeBtn = h('button', { type: 'button', class: 'admin-icon-button', 'aria-label': 'Cerrar', onclick: () => dlg.close() }, icon('x'));
		const foot = h('div', { class: 'admin-modal-actions' });
		let result = null;
		const accept = () => { result = input.value.trim(); dlg.close(); };
		foot.append(
			h('button', { type: 'button', class: 'admin-button', onclick: () => dlg.close() }, 'Cancelar'),
			h('button', { type: 'button', class: 'admin-button admin-button-primary', onclick: accept }, options.confirm || 'Aceptar'),
		);
		input.addEventListener('keydown', (event) => { if (event.key === 'Enter') { event.preventDefault(); accept(); } });
		dlg.append(
			h('div', { class: 'admin-modal-head' }, h('h2', {}, title || 'Editar'), closeBtn),
			h('div', { class: 'admin-modal-body' }, body),
			foot,
		);
		document.body.appendChild(dlg);
		dlg.addEventListener('close', () => { dlg.remove(); resolve(result); }, { once: true });
		dlg.showModal();
		input.focus();
		input.select();
	});
}

export function loading(label = 'Cargando…') {
	return h('div', { class: 'admin-loading' }, h('span', { class: 'admin-spinner', 'aria-hidden': 'true' }), label);
}

export function formatDate(value) {
	if (!value) return '';
	const text = String(value).slice(0, 10);
	return /^\d{4}-\d{2}-\d{2}$/.test(text) ? text : String(value);
}

export function debounce(fn, wait = 220) {
	let timer = 0;
	return (...args) => {
		clearTimeout(timer);
		timer = setTimeout(() => fn(...args), wait);
	};
}

export function copyText(text) {
	if (navigator.clipboard && window.isSecureContext) return navigator.clipboard.writeText(text);
	const area = document.createElement('textarea');
	area.value = text;
	area.style.position = 'fixed';
	area.style.opacity = '0';
	document.body.appendChild(area);
	area.select();
	document.execCommand('copy');
	area.remove();
	return Promise.resolve();
}

export function errorText(err) {
	if (!err) return 'Error desconocido';
	return err.message || String(err);
}
