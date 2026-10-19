import { confirmDialog, dialog, h, icon, loading, toast } from './ui.js';
import { basename, extname, joinPath, slugify, uploadsURL } from './paths.js';

const IMAGE_EXT = new Set(['.jpg', '.jpeg', '.png', '.webp', '.gif', '.avif', '.svg']);

async function loadUploads(state) {
	try {
		return await state.provider.listUploads();
	} catch (err) {
		toast(err.message || 'No se pudieron listar los medios', 'error');
		return [];
	}
}

function fileToDataURL(file) {
	return new Promise((resolve, reject) => {
		const reader = new FileReader();
		reader.onload = () => resolve(String(reader.result || ''));
		reader.onerror = () => reject(new Error('No se pudo leer el archivo'));
		reader.readAsDataURL(file);
	});
}

async function uploadFile(state, file) {
	const max = Number(state.config.max_upload || 10 * 1024 * 1024);
	if (file.size > max) throw new Error('La imagen supera el límite de ' + Math.round(max / 1024 / 1024) + ' MB');
	const ext = extname(file.name);
	if (!IMAGE_EXT.has(ext)) throw new Error('Formato de imagen no permitido');
	const base = slugify(basename(file.name).slice(0, -ext.length)) || 'imagen';
	const name = base + '-' + Date.now() + ext;
	const dataUrl = await fileToDataURL(file);
	const base64 = dataUrl.split(',')[1] || '';
	const folder = state.config.uploads_dir || 'assets/uploads';
	const path = joinPath(folder, name);
	await state.provider.uploadBinary(path, base64, 'assets: subir ' + name);
	const url = uploadsURL(state.config, name);
	state.localUploads.set(url, dataUrl);
	state.localUploads.set(name, dataUrl);
	return { name, path, url, dataUrl, sha: '' };
}

function imageFor(state, item) {
	return state.localUploads.get(item.url) || state.localUploads.get(item.name) || item.url;
}

function mediaItem(state, item, onSelect, onDelete) {
	const img = h('img', { class: 'admin-media-thumb', src: imageFor(state, item), alt: item.name, loading: 'lazy' });
	img.addEventListener('error', () => {
		img.removeAttribute('src');
		img.alt = item.name;
	});
	const selectBtn = h('button', {
		type: 'button',
		class: 'admin-button',
		onclick: () => onSelect && onSelect(item),
	}, 'Usar');
	const copyBtn = h('button', {
		type: 'button',
		class: 'admin-icon-button',
		title: 'Copiar URL',
		'aria-label': 'Copiar URL',
		onclick: async () => {
			await navigator.clipboard?.writeText(item.url);
			toast('URL copiada', 'success');
		},
	}, icon('copy'));
	const actions = h('div', { class: 'admin-media-actions' }, selectBtn, copyBtn);
	if (onDelete) {
		actions.appendChild(h('button', {
			type: 'button',
			class: 'admin-icon-button danger',
			title: 'Borrar',
			'aria-label': 'Borrar',
			onclick: () => onDelete(item),
		}, icon('trash')));
	}
	return h('div', { class: 'admin-media-item' },
		img,
		h('span', { class: 'admin-media-name' }, item.name),
		actions,
	);
}

function renderGrid(container, state, uploads, onSelect, onDelete) {
	container.replaceChildren();
	if (!uploads.length) {
		container.appendChild(h('p', { class: 'admin-empty' }, 'No hay imágenes en assets/uploads.'));
		return;
	}
	const grid = h('div', { class: 'admin-media-grid' });
	for (const item of uploads) grid.appendChild(mediaItem(state, item, onSelect, onDelete));
	container.appendChild(grid);
}

export function openMediaPicker(state) {
	return new Promise((resolve) => {
		const body = h('div', {});
		const toolbar = h('div', { class: 'admin-toolbar' });
		const uploadInput = h('input', { type: 'file', accept: 'image/*', class: 'admin-file-input' });
		const uploadLabel = h('label', { class: 'admin-button admin-file-button' }, 'Subir imagen', uploadInput);
		const refresh = h('button', {
			type: 'button',
			class: 'admin-icon-button',
			title: 'Recargar',
			'aria-label': 'Recargar',
			onclick: () => load(),
		}, icon('refresh'));
		toolbar.append(uploadLabel, refresh);
		const grid = h('div', {}, loading('Cargando medios…'));
		body.append(toolbar, grid);

		const modal = dialog({
			title: 'Biblioteca de medios',
			body,
			actions: [{ label: 'Cerrar' }],
			width: 'min(980px, calc(100vw - 2rem))',
		});

		let uploads = [];
		let settled = false;
		const finish = (value) => {
			if (settled) return;
			settled = true;
			resolve(value);
			modal.close();
		};
		const rerender = () => renderGrid(grid, state, uploads, (item) => finish(item.url));
		const load = async () => {
			grid.replaceChildren(loading('Cargando medios…'));
			uploads = await loadUploads(state);
			rerender();
		};
		uploadInput.addEventListener('change', async () => {
			const file = uploadInput.files?.[0];
			if (!file) return;
			try {
				const item = await uploadFile(state, file);
				toast('Imagen subida', 'success');
				finish(item.url);
			} catch (err) {
				toast(err.message || 'Error al subir', 'error');
			}
			uploadInput.value = '';
		});
		modal.dialog.addEventListener('close', () => finish(null), { once: true });
		load();
	});
}

export async function renderMediaPage(main, state) {
	main.replaceChildren(loading('Cargando medios…'));
	const head = h('section', { class: 'admin-hero' },
		h('div', {},
			h('p', { class: 'admin-eyebrow' }, 'Medios'),
			h('h1', {}, 'Biblioteca de imágenes'),
			h('p', { class: 'admin-lead' }, 'Imágenes en ' + (state.config.uploads_dir || 'assets/uploads') + '. Se publican en ' + (state.config.uploads_url || '/uploads') + ' tras el build.'),
		),
	);
	const uploadInput = h('input', { type: 'file', accept: 'image/*', class: 'admin-file-input' });
	const toolbar = h('div', { class: 'admin-toolbar' },
		h('label', { class: 'admin-button admin-button-primary admin-file-button' }, 'Subir imagen', uploadInput),
		h('button', {
			type: 'button',
			class: 'admin-icon-button',
			title: 'Recargar',
			'aria-label': 'Recargar',
			onclick: () => load(),
		}, icon('refresh')),
	);
	const grid = h('div', {}, loading('Cargando medios…'));
	main.replaceChildren(head, toolbar, grid);

	let uploads = [];
	const load = async () => {
		grid.replaceChildren(loading('Cargando medios…'));
		uploads = await loadUploads(state);
		renderGrid(grid, state, uploads, (item) => {
			navigator.clipboard?.writeText(item.url);
			toast('URL copiada', 'success');
		}, async (item) => {
			const ok = await confirmDialog('¿Borrar “' + item.name + '” del repositorio?', { danger: true, confirm: 'Borrar' });
			if (!ok) return;
			try {
				await state.provider.deleteFile(item.path, item.sha, 'assets: borrar ' + item.name);
				state.localUploads.delete(item.url);
				toast('Imagen borrada', 'success');
				await load();
			} catch (err) {
				toast(err.message || 'Error al borrar', 'error');
			}
		});
	};
	uploadInput.addEventListener('change', async () => {
		const file = uploadInput.files?.[0];
		if (!file) return;
		try {
			const item = await uploadFile(state, file);
			uploads = [item, ...uploads.filter((entry) => entry.name !== item.name)];
			renderGrid(grid, state, uploads, (entry) => {
				navigator.clipboard?.writeText(entry.url);
				toast('URL copiada', 'success');
			}, null);
			toast('Imagen subida', 'success');
		} catch (err) {
			toast(err.message || 'Error al subir', 'error');
		}
		uploadInput.value = '';
	});
	await load();
}
