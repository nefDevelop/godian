import { dump, load } from './yaml.js';

export function splitDocument(text) {
	const source = String(text || '').replace(/^\uFEFF/, '');
	if (!/^---\r?\n/.test(source)) {
		return { raw: '', body: source };
	}
	const rest = source.slice(source.indexOf('\n') + 1);
	const lines = rest.split(/\r?\n/);
	const raw = [];
	let index = 0;
	for (; index < lines.length; index += 1) {
		if (lines[index].trim() === '---') break;
		raw.push(lines[index]);
	}
	if (index >= lines.length) {
		return { raw: '', body: source };
	}
	return { raw: raw.join('\n'), body: lines.slice(index + 1).join('\n') };
}

export function parseDocument(collection, text) {
	if (collection?.format === 'frontmatter') {
		const { raw, body } = splitDocument(text);
		return { data: raw.trim() ? load(raw) : {}, body };
	}
	return { data: load(text), body: '' };
}

export function stringifyDocument(collection, data, body) {
	if (collection?.format === 'frontmatter') {
		const text = String(body || '');
		const separator = /^\r?\n/.test(text) ? '\n' : '\n\n';
		return '---\n' + dump(data) + '---' + separator + text;
	}
	return dump(data);
}

export function titleFrom(data, body, fallback = '') {
	const value = data?.title || data?.name;
	if (value !== undefined && value !== null && String(value).trim() !== '') return String(value);
	for (const line of String(body || '').split(/\r?\n/)) {
		const trimmed = line.trim();
		if (trimmed.startsWith('# ')) return trimmed.slice(2).trim();
	}
	return String(fallback || '').replace(/\.[^.]+$/, '');
}

export function dateFrom(data) {
	for (const key of ['published', 'pubDate', 'date', 'updatedDate', 'updated']) {
		const value = data?.[key];
		if (!value) continue;
		if (value instanceof Date) return value.toISOString().slice(0, 10);
		const text = String(value);
		return text.slice(0, 10);
	}
	return '';
}

export function isMarkdown(collection) {
	return collection?.format === 'frontmatter';
}
