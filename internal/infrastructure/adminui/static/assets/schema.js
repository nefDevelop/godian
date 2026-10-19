import { h } from './ui.js';
import { isEmpty } from './yaml.js';
import { today } from './paths.js';

export function findCollection(schema, name) {
	return (schema?.collections || []).find((collection) => collection.name === name) || null;
}

export function defaultValues(collection) {
	const data = {};
	for (const field of collection?.fields || []) {
		if (field.default !== undefined) data[field.name] = field.default;
		if (field.type === 'datetime' && (field.name === 'published' || field.name === 'pubDate')) {
			if (data[field.name] === undefined) data[field.name] = today();
		}
	}
	if (collection?.name === 'blog') data.draft = true;
	return data;
}

export function sourceOptions(field, schema, options) {
	const source = field.source || '';
	if (source.startsWith('taxonomy:')) {
		return options?.taxonomies?.[source.slice('taxonomy:'.length)] || [];
	}
	if (source === 'authors') return options?.authors || [];
	if (source === 'series') return options?.series || [];
	if (field.options) return field.options;
	return [];
}

function inputId(path) {
	return 'f-' + String(path).replace(/[^A-Za-z0-9_-]+/g, '-');
}

function valueString(value) {
	if (value === null || value === undefined) return '';
	if (value instanceof Date) return value.toISOString().slice(0, 10);
	return String(value);
}

function listValue(value) {
	if (Array.isArray(value)) return value.map(valueString).join(', ');
	return valueString(value);
}

function splitList(value) {
	const seen = new Set();
	const out = [];
	for (const part of String(value || '').split(/[,\n\r]+/)) {
		const item = part.trim();
		if (!item || seen.has(item)) continue;
		seen.add(item);
		out.push(item);
	}
	return out;
}

function emptyValue(value) {
	return isEmpty(value);
}

export function validateFields(collection, data) {
	const errors = [];
	for (const field of collection?.fields || []) {
		const value = data?.[field.name];
		if (field.required && emptyValue(value)) {
			errors.push('El campo “' + (field.label || field.name) + '” es obligatorio.');
			continue;
		}
		if (value === undefined || value === null || value === '') continue;
		if (field.type === 'datetime' && !/^\d{4}-\d{2}-\d{2}$/.test(String(value).slice(0, 10))) {
			errors.push('“' + (field.label || field.name) + '” debe ser una fecha YYYY-MM-DD.');
		}
		if (field.type === 'number' && Number.isNaN(Number(value))) {
			errors.push('“' + (field.label || field.name) + '” debe ser un número.');
		}
	}
	return errors;
}

export function renderFields(collection, data, ctx) {
	return renderFieldSet(collection?.fields || [], data || {}, ctx, 'field');
}

function renderFieldSet(fields, data, ctx, prefix) {
	const container = h('div', { class: 'admin-form-grid' });
	const editors = [];
	for (const field of fields) {
		const editor = renderField(field, data?.[field.name], ctx, prefix + '.' + field.name);
		editors.push(editor);
		if (editor.el) container.appendChild(editor.el);
	}
	const get = () => {
		const out = {};
		for (const editor of editors) {
			const value = editor.get();
			if (!emptyValue(value)) out[editor.name] = value;
		}
		return out;
	};
	return { el: container, get };
}

function labelFor(field) {
	return (field.label || field.name) + (field.required ? ' *' : '');
}

function renderField(field, value, ctx, path) {
	const id = inputId(path);
	const initial = value === undefined ? field.default : value;

	if (field.type === 'object') {
		const child = renderFieldSet(field.fields || [], initial || {}, ctx, path);
		const box = h('fieldset', { class: 'admin-fieldset' },
			h('legend', {}, labelFor(field)),
			child.el,
		);
		return { name: field.name, el: box, get: () => child.get() };
	}

	if (field.type === 'list' && Array.isArray(field.fields) && field.fields.length) {
		const rows = h('div', { class: 'admin-row-list' });
		const rowEditors = [];
		const addRow = (rowValue = {}) => {
			const rowIndex = rowEditors.length;
			const child = renderFieldSet(field.fields, rowValue, ctx, path + '.' + rowIndex);
			const remove = h('button', {
				type: 'button',
				class: 'admin-icon-button danger',
				title: 'Quitar',
				'aria-label': 'Quitar',
				onclick: () => {
					row.remove();
					const index = rowEditors.indexOf(editor);
					if (index >= 0) rowEditors.splice(index, 1);
				},
			}, '×');
			const row = h('div', { class: 'admin-row-card' },
				h('div', { class: 'admin-row-head' },
					h('strong', {}, (field.label || field.name) + ' ' + (rowIndex + 1)),
					remove,
				),
				child.el,
			);
			const editor = { el: row, get: child.get };
			rowEditors.push(editor);
			rows.appendChild(row);
		};
		const box = h('fieldset', { class: 'admin-fieldset' },
			h('legend', {}, (field.label || field.name)),
			rows,
			h('button', { type: 'button', class: 'admin-button', onclick: () => addRow({}) }, 'Añadir'),
		);
		if (Array.isArray(initial)) initial.forEach((item) => addRow(item || {}));
		return {
			name: field.name,
			el: box,
			get: () => rowEditors.map((editor) => editor.get()).filter((item) => !emptyValue(item)),
		};
	}

	const wrap = h('div', { class: 'admin-field' });
	const label = h('label', { for: id }, labelFor(field));
	let getter;

	if (field.type === 'boolean') {
		const input = h('input', { type: 'checkbox', id, checked: Boolean(initial) });
		wrap.className = 'admin-field admin-check';
		wrap.append(input, h('label', { for: id }, field.label || field.name));
		getter = () => input.checked;
		wrap.setAttribute('data-required', field.required ? '1' : '0');
		if (field.required) wrap.setAttribute('aria-required', 'true');
		return { name: field.name, el: wrap, get: getter };
	}

	if (field.type === 'text') {
		wrap.classList.add('admin-field-wide');
		const input = h('textarea', { id, rows: 3 }, valueString(initial));
		wrap.append(label, input);
		getter = () => input.value.trim();
	} else if (field.type === 'datetime') {
		const input = h('input', { type: 'date', id, value: valueString(initial).slice(0, 10) });
		wrap.append(label, input);
		getter = () => input.value;
	} else if (field.type === 'number') {
		const input = h('input', { type: 'number', id, step: '1', value: valueString(initial) });
		wrap.append(label, input);
		getter = () => (input.value === '' ? undefined : Number(input.value));
	} else if (field.type === 'select') {
		const options = sourceOptions(field, ctx.schema, ctx.options);
		if (Array.isArray(field.options) && field.options.length) {
			const select = h('select', { id });
			select.appendChild(h('option', { value: '' }, '—'));
			for (const option of field.options) {
				select.appendChild(h('option', { value: option, selected: option === valueString(initial) }, option));
			}
			wrap.append(label, select);
			getter = () => select.value;
		} else {
			const listId = id + '-list';
			const input = h('input', { id, type: 'text', value: valueString(initial), list: options.length ? listId : null });
			wrap.append(label, input);
			if (options.length) {
				const datalist = h('datalist', { id: listId });
				for (const option of options) datalist.appendChild(h('option', { value: option }));
				wrap.appendChild(datalist);
			}
			getter = () => input.value.trim();
		}
	} else if (field.type === 'list') {
		wrap.classList.add('admin-field-wide');
		const options = sourceOptions(field, ctx.schema, ctx.options);
		const listId = id + '-list';
		const input = h('input', { id, type: 'text', value: listValue(initial), list: options.length ? listId : null });
		wrap.append(label, input);
		if (options.length) {
			const datalist = h('datalist', { id: listId });
			for (const option of options) datalist.appendChild(h('option', { value: option }));
			wrap.appendChild(datalist);
		}
		wrap.appendChild(h('p', { class: 'admin-hint' }, 'Separa por comas.'));
		getter = () => splitList(input.value);
	} else if (field.type === 'image') {
		const input = h('input', { id, type: 'text', value: valueString(initial), placeholder: '/uploads/…' });
		const preview = h('img', { class: 'admin-thumb', src: valueString(initial), alt: '' });
		const update = () => {
			preview.src = input.value.trim();
			preview.style.display = input.value.trim() ? 'block' : 'none';
		};
		const picker = h('button', {
			type: 'button',
			class: 'admin-button',
			onclick: async () => {
				if (!ctx.pickImage) return;
				const url = await ctx.pickImage();
				if (url) {
					input.value = url;
					update();
				}
			},
		}, 'Elegir imagen');
		input.addEventListener('input', update);
		update();
		wrap.append(label, input, h('div', { class: 'admin-editor-tools' }, picker), preview);
		getter = () => input.value.trim();
	} else {
		const type = field.type === 'url' ? 'url' : 'text';
		const input = h('input', { id, type, value: valueString(initial) });
		wrap.append(label, input);
		getter = () => input.value.trim();
	}

	if (field.required) {
		const input = wrap.querySelector('input, textarea, select');
		if (input) input.setAttribute('required', 'required');
	}
	return { name: field.name, el: wrap, get: getter };
}
