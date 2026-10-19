import { confirmDialog, errorText, h, icon, loading, promptDialog, toast } from './ui.js';
import { dateFrom, parseDocument, stringifyDocument, titleFrom } from './frontmatter.js';
import { basename, filenameFor, joinPath, publicURL, relativePath } from './paths.js';

async function mapLimit(items, limit, fn) {
	let index = 0;
	const workers = Array.from({ length: Math.min(limit, items.length) }, async () => {
		while (index < items.length) {
			const current = items[index];
			index += 1;
			await fn(current);
		}
	});
	await Promise.all(workers);
}

async function loadCollection(state, collection) {
	if (state.files.has(collection.name)) return state.files.get(collection.name);
	const list = await state.provider.listFiles(collection.folder, collection.extension);
	const files = new Array(list.length);
	await mapLimit(list, 4, async (entry, i) => {
		const index = list.indexOf(entry);
		try {
			const file = await state.provider.readFile(entry.path);
			const parsed = parseDocument(collection, file.content);
			files[index] = {
				path: entry.path,
				name: entry.name,
				sha: file.sha,
				title: titleFrom(parsed.data, parsed.body, basename(entry.path)),
				date: dateFrom(parsed.data),
			};
		} catch (err) {
			files[index] = {
				path: entry.path,
				name: entry.name,
				sha: entry.sha,
				title: basename(entry.path).replace(/\.[^.]+$/, ''),
				date: '',
			};
		}
	});
	state.files.set(collection.name, files);
	return files;
}

function sortFiles(files, mode) {
	const copy = files.slice();
	if (mode === 'title') {
		copy.sort((a, b) => a.title.localeCompare(b.title, 'es'));
	} else if (mode === 'oldest') {
		copy.sort((a, b) => String(a.date || '9999').localeCompare(String(b.date || '9999')));
	} else {
		copy.sort((a, b) => String(b.date || '').localeCompare(String(a.date || '')) || a.title.localeCompare(b.title, 'es'));
	}
	return copy;
}

async function duplicateEntry(state, collection, file, navigate) {
	const suggested = 'copia-' + basename(file.path);
	const value = await promptDialog('Duplicar archivo', 'Nombre del nuevo archivo', suggested, { confirm: 'Duplicar' });
	if (!value) return;
	const target = joinPath(collection.folder, filenameFor(collection, value));
	if (target === file.path) {
		toast('El nombre es el mismo', 'error');
		return;
	}
	try {
		const source = await state.provider.readFile(file.path);
		const parsed = parseDocument(collection, source.content);
		if (collection.format === 'frontmatter') {
			parsed.data.title = 'Copia de ' + titleFrom(parsed.data, parsed.body, basename(file.path));
		} else if (parsed.data.title !== undefined) {
			parsed.data.title = 'Copia de ' + (parsed.data.title || '');
		} else if (parsed.data.name !== undefined) {
			parsed.data.name = 'Copia de ' + (parsed.data.name || '');
		}
		const content = stringifyDocument(collection, parsed.data, parsed.body);
		await state.provider.writeFile(target, content, null, 'content(' + collection.name + '): duplicar ' + basename(file.path));
		state.files.delete(collection.name);
		toast('Archivo duplicado', 'success');
		navigate('#/edit?collection=' + encodeURIComponent(collection.name) + '&path=' + encodeURIComponent(target));
	} catch (err) {
		toast(errorText(err), 'error');
	}
}

async function renameEntry(state, collection, file, navigate) {
	const value = await promptDialog('Renombrar archivo', 'Nuevo nombre de archivo', basename(file.path), { confirm: 'Renombrar' });
	if (!value) return;
	const target = joinPath(collection.folder, filenameFor(collection, value));
	if (target === file.path) return;
	try {
		const source = await state.provider.readFile(file.path);
		await state.provider.writeFile(target, source.content, null, 'content(' + collection.name + '): renombrar ' + basename(file.path));
		try {
			await state.provider.deleteFile(file.path, source.sha, 'content(' + collection.name + '): borrar ' + basename(file.path));
		} catch (err) {
			toast('Se creó el nuevo archivo, pero no se pudo borrar el anterior: ' + errorText(err), 'error');
		}
		state.files.delete(collection.name);
		toast('Archivo renombrado', 'success');
		navigate('#/edit?collection=' + encodeURIComponent(collection.name) + '&path=' + encodeURIComponent(target));
	} catch (err) {
		toast(errorText(err), 'error');
	}
}

async function deleteEntry(state, collection, file) {
	const ok = await confirmDialog('¿Borrar “' + file.title + '” del repositorio?', { danger: true, confirm: 'Borrar' });
	if (!ok) return;
	try {
		const source = await state.provider.readFile(file.path);
		await state.provider.deleteFile(file.path, source.sha, 'content(' + collection.name + '): borrar ' + basename(file.path));
		state.files.delete(collection.name);
		toast('Archivo borrado', 'success');
	} catch (err) {
		toast(errorText(err), 'error');
	}
}

function fileRow(state, collection, file, navigate, refresh) {
	const publicLink = publicURL(state.config, collection, file.path);
	const actions = h('div', { class: 'admin-row-actions' });
	actions.append(
		h('button', {
			type: 'button',
			class: 'admin-icon-button',
			title: 'Editar',
			'aria-label': 'Editar',
			onclick: () => navigate('#/edit?collection=' + encodeURIComponent(collection.name) + '&path=' + encodeURIComponent(file.path)),
		}, icon('edit')),
	);
	if (publicLink) {
		actions.appendChild(h('a', {
			class: 'admin-icon-button',
			href: publicLink,
			target: '_blank',
			rel: 'noopener',
			title: 'Ver en el sitio',
			'aria-label': 'Ver en el sitio',
		}, icon('external')));
	}
	actions.append(
		h('button', {
			type: 'button',
			class: 'admin-icon-button',
			title: 'Duplicar',
			'aria-label': 'Duplicar',
			onclick: () => duplicateEntry(state, collection, file, navigate),
		}, icon('duplicate')),
		h('button', {
			type: 'button',
			class: 'admin-icon-button',
			title: 'Renombrar',
			'aria-label': 'Renombrar',
			onclick: () => renameEntry(state, collection, file, navigate),
		}, icon('rename')),
		h('button', {
			type: 'button',
			class: 'admin-icon-button danger',
			title: 'Borrar',
			'aria-label': 'Borrar',
			onclick: async () => { await deleteEntry(state, collection, file); refresh(); },
		}, icon('trash')),
	);
	return h('li', { class: 'admin-file-row' },
		h('a', {
			class: 'admin-file',
			href: '#/edit?collection=' + encodeURIComponent(collection.name) + '&path=' + encodeURIComponent(file.path),
			onclick: (event) => { event.preventDefault(); navigate('#/edit?collection=' + encodeURIComponent(collection.name) + '&path=' + encodeURIComponent(file.path)); },
		},
			h('span', { class: 'admin-file-title' }, file.title),
			h('span', { class: 'admin-file-meta' }, relativePath(collection, file.path) + (file.date ? ' · ' + file.date : '')),
		),
		actions,
	);
}

export async function renderDashboard(main, state, navigate) {
	let search = '';
	let sort = 'newest';
	const head = h('section', { class: 'admin-hero' },
		h('div', {},
			h('p', { class: 'admin-eyebrow' }, 'Contenido'),
			h('h1', {}, 'Panel de edición'),
			h('p', { class: 'admin-lead' }, 'Edita el contenido del repositorio y publica con un commit. El build del sitio valida el resultado.'),
		),
	);
	const searchInput = h('input', { class: 'admin-search', type: 'search', placeholder: 'Buscar por título o ruta…', 'aria-label': 'Buscar contenido' });
	const sortSelect = h('select', { class: 'admin-select', 'aria-label': 'Ordenar' },
		h('option', { value: 'newest' }, 'Más recientes'),
		h('option', { value: 'oldest' }, 'Más antiguos'),
		h('option', { value: 'title' }, 'Título'),
	);
	const toolbar = h('div', { class: 'admin-toolbar' },
		searchInput,
		sortSelect,
		h('button', {
			type: 'button',
			class: 'admin-icon-button',
			title: 'Recargar',
			'aria-label': 'Recargar',
			onclick: () => { state.files.clear(); load(); },
		}, icon('refresh')),
	);
	const content = h('div', { class: 'admin-collections' }, loading('Cargando contenido…'));
	main.replaceChildren(head, toolbar, content);

	const refresh = () => renderCollections();
	const renderCollections = () => {
		const term = search.trim().toLowerCase();
		content.replaceChildren();
		let any = false;
		for (const collection of state.schema.collections) {
			const cached = state.files.get(collection.name);
			if (!cached) continue;
			any = true;
			let files = cached.filter((file) => {
				if (!term) return true;
				return file.title.toLowerCase().includes(term) || file.path.toLowerCase().includes(term);
			});
			files = sortFiles(files, sort);
			const list = h('ul', { class: 'admin-file-list' });
			for (const file of files) list.appendChild(fileRow(state, collection, file, navigate, refresh));
			const card = h('section', { class: 'admin-card' },
				h('header', { class: 'admin-card-head' },
					h('div', {},
						h('h2', {}, collection.label || collection.name),
						h('p', { class: 'admin-muted' }, collection.folder + ' · .' + collection.extension),
					),
					h('a', {
						class: 'admin-button admin-button-primary',
						href: '#/new?collection=' + encodeURIComponent(collection.name),
						onclick: (event) => { event.preventDefault(); navigate('#/new?collection=' + encodeURIComponent(collection.name)); },
					}, 'Nuevo'),
				),
				files.length ? list : h('p', { class: 'admin-empty' }, cached.length ? 'Sin resultados.' : 'No hay archivos todavía.'),
			);
			content.appendChild(card);
		}
		if (!any) content.appendChild(loading('Cargando contenido…'));
	};

	const load = async () => {
		content.replaceChildren(loading('Cargando contenido…'));
		try {
			await Promise.all(state.schema.collections.map((collection) => loadCollection(state, collection)));
			renderCollections();
		} catch (err) {
			content.replaceChildren(h('div', { class: 'admin-alert admin-alert-error' }, errorText(err)));
		}
	};

	searchInput.addEventListener('input', () => { search = searchInput.value; renderCollections(); });
	sortSelect.addEventListener('change', () => { sort = sortSelect.value; renderCollections(); });
	await load();
}
