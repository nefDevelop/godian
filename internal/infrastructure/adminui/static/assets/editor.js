import { confirmDialog, debounce, errorText, h, icon, loading, promptDialog, toast } from './ui.js';
import { defaultValues, findCollection, renderFields, validateFields } from './schema.js';
import { isMarkdown, parseDocument, stringifyDocument, titleFrom } from './frontmatter.js';
import { applyToolbar, bindScrollSync, installShortcuts, renderInto } from './markdown.js';
import { openMediaPicker } from './media.js';
import { basename, filenameFor, joinPath, publicURL, relativePath, today } from './paths.js';

function insertAtCursor(textarea, text) {
	const start = textarea.selectionStart || 0;
	const end = textarea.selectionEnd || 0;
	textarea.value = textarea.value.slice(0, start) + text + textarea.value.slice(end);
	textarea.selectionStart = textarea.selectionEnd = start + text.length;
	textarea.focus();
	textarea.dispatchEvent(new Event('input', { bubbles: true }));
}

function toolbarButton(label, action, title, textarea, onImage) {
	return h('button', {
		type: 'button',
		title: title || label,
		'aria-label': title || label,
		onclick: () => {
			if (action === 'image' && onImage) onImage();
			else applyToolbar(textarea, action);
		},
	}, label);
}

export async function renderEditor(main, state, params, navigate, forceNew = false) {
	const collection = findCollection(state.schema, params.get('collection'));
	if (!collection) {
		main.replaceChildren(h('div', { class: 'admin-alert admin-alert-error' }, 'Colección desconocida.'));
		return;
	}

	let currentPath = params.get('path') || '';
	let currentIsNew = forceNew || !currentPath;
	let data = defaultValues(collection);
	let body = '';
	let sha = null;
	let initialTitle = '';

	main.replaceChildren(loading('Cargando editor…'));
	try {
		if (!currentIsNew) {
			const file = await state.provider.readFile(currentPath);
			const parsed = parseDocument(collection, file.content);
			data = parsed.data;
			body = parsed.body;
			sha = file.sha;
			initialTitle = titleFrom(data, body, basename(currentPath));
		}
	} catch (err) {
		main.replaceChildren(h('div', { class: 'admin-alert admin-alert-error' }, errorText(err)));
		return;
	}

	let title = initialTitle || (currentIsNew ? 'Nuevo ' + (collection.label || collection.name) : basename(currentPath));
	const ctx = {
		schema: state.schema,
		options: state.options,
		pickImage: () => openMediaPicker(state),
	};
	const fields = renderFields(collection, data, ctx);

	const titleEl = h('h1', {}, title);
	const pathEl = h('p', { class: 'admin-muted' }, currentPath || collection.folder + '/…');
	const publicLink = currentIsNew ? '' : publicURL(state.config, collection, currentPath);
	const viewBtn = h('a', {
		class: 'admin-button',
		href: publicLink || '#',
		target: '_blank',
		rel: 'noopener',
		hidden: !publicLink,
	}, 'Ver en el sitio');

	const saveBtn = h('button', { type: 'button', class: 'admin-button admin-button-primary' }, 'Guardar');
	const publishBtn = h('button', { type: 'button', class: 'admin-button' }, 'Publicar');
	const duplicateBtn = h('button', { type: 'button', class: 'admin-button', hidden: currentIsNew }, 'Duplicar');
	const renameBtn = h('button', { type: 'button', class: 'admin-button', hidden: currentIsNew }, 'Renombrar');
	const deleteBtn = h('button', { type: 'button', class: 'admin-button admin-button-danger', hidden: currentIsNew }, 'Borrar');

	const head = h('section', { class: 'admin-edit-head' },
		h('div', {},
			h('nav', { class: 'admin-breadcrumb' },
				h('a', { href: '#/', onclick: (event) => { event.preventDefault(); navigate('#/'); } }, 'Contenido'),
				h('span', {}, '/'),
				h('span', {}, collection.label || collection.name),
			),
			h('p', { class: 'admin-eyebrow' }, collection.label || collection.name),
			titleEl,
			pathEl,
		),
		h('div', { class: 'admin-actions' },
			viewBtn,
			duplicateBtn,
			renameBtn,
			deleteBtn,
			publishBtn,
			saveBtn,
		),
	);
	const commitBox = h('div', { class: 'admin-commit', hidden: true });
	const form = h('div', { class: 'admin-form' });

	let filenameInput = null;
	if (currentIsNew) {
		filenameInput = h('input', { id: 'filename', type: 'text', value: '', placeholder: 'mi-entrada' });
		form.appendChild(h('div', { class: 'admin-field' },
			h('label', { for: 'filename' }, 'Nombre del archivo'),
			filenameInput,
			h('p', { class: 'admin-hint' }, 'Sin extensión; se añade .' + collection.extension),
		));
	}
	form.appendChild(fields.el);

	let bodyInput = null;
	let preview = null;
	if (isMarkdown(collection)) {
		bodyInput = h('textarea', { id: 'body', class: 'admin-editor-textarea', spellcheck: 'false' }, body);
		preview = h('div', { id: 'admin-preview', class: 'admin-preview prose' });
		const pickBodyImage = async () => {
			const url = await openMediaPicker(state);
			if (url) insertAtCursor(bodyInput, '![imagen](' + url + ')\n');
		};
		const toolbar = h('div', { class: 'admin-md-toolbar' },
			toolbarButton('H1', 'h1', 'Encabezado 1', bodyInput),
			toolbarButton('H2', 'h2', 'Encabezado 2', bodyInput),
			toolbarButton('H3', 'h3', 'Encabezado 3', bodyInput),
			h('span', { class: 'sep' }),
			toolbarButton('B', 'bold', 'Negrita', bodyInput),
			toolbarButton('I', 'italic', 'Cursiva', bodyInput),
			toolbarButton('S', 'strike', 'Tachado', bodyInput),
			toolbarButton('==', 'highlight', 'Resaltado', bodyInput),
			h('span', { class: 'sep' }),
			toolbarButton('Link', 'link', 'Enlace', bodyInput),
			toolbarButton('Imagen', 'image', 'Imagen', bodyInput, pickBodyImage),
			toolbarButton('Código', 'code', 'Bloque de código', bodyInput),
			toolbarButton('Inline', 'code-inline', 'Código inline', bodyInput),
			h('span', { class: 'sep' }),
			toolbarButton('Cita', 'quote', 'Cita', bodyInput),
			toolbarButton('Lista', 'list', 'Lista', bodyInput),
			toolbarButton('1.', 'ordered', 'Lista numerada', bodyInput),
		);
		const editor = h('div', { class: 'admin-split' },
			h('div', { class: 'admin-editor-pane' },
				h('label', { for: 'body' }, 'Cuerpo Markdown'),
				toolbar,
				bodyInput,
			),
			h('div', { class: 'admin-editor-pane' },
				h('div', { class: 'admin-editor-preview-head' },
					h('span', {}, 'Vista previa'),
					h('span', { class: 'admin-muted' }, 'Markdown + KaTeX'),
				),
				preview,
			),
		);
		form.appendChild(editor);
		form.appendChild(h('div', { class: 'admin-editor-tools' },
			h('button', { type: 'button', class: 'admin-button', onclick: pickBodyImage }, 'Insertar imagen'),
			h('span', { class: 'admin-muted' }, 'Se sube a ' + (state.config.uploads_dir || 'assets/uploads') + ' y se inserta como Markdown.'),
		));
		renderInto(preview, body, state.config);
		const updatePreview = debounce(() => renderInto(preview, bodyInput.value, state.config), 250);
		bodyInput.addEventListener('input', updatePreview);
		installShortcuts(bodyInput);
		bindScrollSync(bodyInput, preview);
	}

	form.appendChild(h('div', { class: 'admin-form-actions' },
		h('a', { class: 'admin-button', href: '#/', onclick: (event) => { event.preventDefault(); navigate('#/'); } }, 'Cancelar'),
		saveBtn,
	));
	main.replaceChildren(head, commitBox, form);

	const showCommit = (result) => {
		const url = state.provider.commitURL(result);
		const commitSha = state.provider.commitSHA(result) || sha || '';
		commitBox.hidden = false;
		commitBox.replaceChildren(
			h('span', {}, 'Commit ' + (commitSha ? commitSha.slice(0, 8) : 'creado') + '.'),
			url ? h('a', { href: url, target: '_blank', rel: 'noopener' }, 'Ver en ' + state.provider.displayName) : '',
		);
	};

	const save = async (overrides = {}) => {
		const values = { ...fields.get(), ...overrides };
		const errors = validateFields(collection, values);
		if (errors.length) {
			toast(errors.join(' '), 'error');
			return false;
		}
		let target = currentPath;
		if (currentIsNew) {
			const filename = filenameInput?.value || values.title || values.name || '';
			target = joinPath(collection.folder, filenameFor(collection, filename));
		}
		const content = stringifyDocument(collection, values, bodyInput ? bodyInput.value : '');
		const commitTitle = titleFrom(values, bodyInput ? bodyInput.value : '', basename(target));
		const message = currentIsNew
			? 'content(' + collection.name + '): crear ' + basename(target)
			: 'content(' + collection.name + '): actualizar ' + commitTitle;
		saveBtn.disabled = true;
		publishBtn.disabled = true;
		saveBtn.textContent = 'Guardando…';
		try {
			const result = await state.provider.writeFile(target, content, sha, message);
			sha = result?.content?.sha || result?.commit?.sha || sha;
			if (currentIsNew) {
				currentPath = target;
				currentIsNew = false;
				filenameInput?.closest('.admin-field')?.remove();
				window.history.replaceState(null, '', '#/edit?collection=' + encodeURIComponent(collection.name) + '&path=' + encodeURIComponent(target));
				pathEl.textContent = target;
				viewBtn.hidden = false;
				viewBtn.href = publicURL(state.config, collection, target) || '#';
				duplicateBtn.hidden = false;
				renameBtn.hidden = false;
				deleteBtn.hidden = false;
			}
			title = commitTitle;
			titleEl.textContent = title;
			showCommit(result);
			toast('Guardado y commiteado', 'success');
			return true;
		} catch (err) {
			toast(errorText(err), 'error');
			return false;
		} finally {
			saveBtn.disabled = false;
			publishBtn.disabled = false;
			saveBtn.textContent = 'Guardar';
		}
	};

	saveBtn.addEventListener('click', () => save());
	publishBtn.addEventListener('click', async () => {
		const values = fields.get();
		if (collection.name === 'blog') {
			await save({ draft: !values.draft, published: values.published || today() });
		} else if (collection.name === 'docs') {
			await save({ pubDate: values.pubDate || today() });
		} else {
			await save();
		}
	});

	duplicateBtn.addEventListener('click', async () => {
		const value = await promptDialog('Duplicar', 'Nombre del nuevo archivo', 'copia-' + basename(currentPath), { confirm: 'Duplicar' });
		if (!value) return;
		const target = joinPath(collection.folder, filenameFor(collection, value));
		const values = { ...fields.get(), title: 'Copia de ' + (fields.get().title || fields.get().name || '') };
		try {
			await state.provider.writeFile(target, stringifyDocument(collection, values, bodyInput ? bodyInput.value : ''), null, 'content(' + collection.name + '): duplicar ' + basename(currentPath));
			toast('Archivo duplicado', 'success');
			navigate('#/edit?collection=' + encodeURIComponent(collection.name) + '&path=' + encodeURIComponent(target));
		} catch (err) {
			toast(errorText(err), 'error');
		}
	});

	renameBtn.addEventListener('click', async () => {
		const value = await promptDialog('Renombrar', 'Nuevo nombre de archivo', basename(currentPath), { confirm: 'Renombrar' });
		if (!value) return;
		const target = joinPath(collection.folder, filenameFor(collection, value));
		if (target === currentPath) return;
		const values = fields.get();
		try {
			await state.provider.writeFile(target, stringifyDocument(collection, values, bodyInput ? bodyInput.value : ''), null, 'content(' + collection.name + '): renombrar ' + basename(currentPath));
			await state.provider.deleteFile(currentPath, sha, 'content(' + collection.name + '): borrar ' + basename(currentPath));
			toast('Archivo renombrado', 'success');
			navigate('#/edit?collection=' + encodeURIComponent(collection.name) + '&path=' + encodeURIComponent(target));
		} catch (err) {
			toast(errorText(err), 'error');
		}
	});

	deleteBtn.addEventListener('click', async () => {
		const ok = await confirmDialog('¿Borrar “' + title + '” del repositorio?', { danger: true, confirm: 'Borrar' });
		if (!ok) return;
		try {
			await state.provider.deleteFile(currentPath, sha, 'content(' + collection.name + '): borrar ' + basename(currentPath));
			toast('Archivo borrado', 'success');
			navigate('#/');
		} catch (err) {
			toast(errorText(err), 'error');
		}
	});
}
