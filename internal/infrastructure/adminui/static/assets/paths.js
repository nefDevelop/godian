const ACCENTS = {
	'á': 'a', 'à': 'a', 'ä': 'a', 'â': 'a', 'ã': 'a',
	'é': 'e', 'è': 'e', 'ë': 'e', 'ê': 'e',
	'í': 'i', 'ì': 'i', 'ï': 'i', 'î': 'i',
	'ó': 'o', 'ò': 'o', 'ö': 'o', 'ô': 'o', 'õ': 'o',
	'ú': 'u', 'ù': 'u', 'ü': 'u', 'û': 'u',
	'ñ': 'n', 'ç': 'c',
};

export function slugify(value) {
	let text = String(value || '').toLowerCase().trim();
	text = text.replace(/[áàäâãéèëêíìïîóòöôõúùüûñç]/g, (ch) => ACCENTS[ch] || ch);
	text = text.replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '');
	return text;
}

export function basename(path) {
	const parts = String(path || '').split('/');
	return parts[parts.length - 1] || '';
}

export function extname(path) {
	const name = basename(path);
	const index = name.lastIndexOf('.');
	return index > 0 ? name.slice(index).toLowerCase() : '';
}

export function dirname(path) {
	const parts = String(path || '').split('/');
	parts.pop();
	return parts.join('/');
}

export function joinPath(...parts) {
	return parts
		.filter((part) => part !== undefined && part !== null && String(part) !== '')
		.map((part) => String(part).replace(/^\/+|\/+$/g, ''))
		.join('/');
}

export function collectionFolder(collection) {
	return String(collection?.folder || '').replace(/^\/+|\/+$/g, '');
}

export function relativePath(collection, path) {
	const folder = collectionFolder(collection);
	const clean = String(path || '').replace(/^\/+/, '');
	if (folder && clean.startsWith(folder + '/')) return clean.slice(folder.length + 1);
	return clean;
}

export function slugFromPath(collection, path) {
	const rel = relativePath(collection, path);
	const withoutExt = rel.slice(0, rel.length - extname(rel).length);
	return withoutExt.replace(/\/index$/, '');
}

function escapeSegments(slug) {
	return slug.split('/').map((part) => encodeURIComponent(part)).join('/');
}

export function publicURL(config, collection, path) {
	const name = collection?.name || '';
	const slug = slugFromPath(collection, path);
	if (!slug && name !== 'pages') return '';
	let route = '';
	if (name === 'blog') route = '/blog/' + escapeSegments(slug) + '/';
	else if (name === 'docs') route = '/docs/' + escapeSegments(slug) + '/';
	else if (name === 'series') route = '/series/' + escapeSegments(slug) + '/';
	else if (name === 'pages') route = slug ? '/' + escapeSegments(slug) + '/' : '/';
	else return '';
	const base = String(config?.base_path || '').replace(/\/+$/, '');
	return base + route;
}

export function uploadsURL(config, name) {
	const base = String(config?.uploads_url || '/uploads').replace(/\/+$/, '');
	return base + '/' + encodeURIComponent(name);
}

export function today() {
	const now = new Date();
	const offset = now.getTimezoneOffset();
	const local = new Date(now.getTime() - offset * 60 * 1000);
	return local.toISOString().slice(0, 10);
}

export function filenameFor(collection, value) {
	const ext = collection?.extension || 'md';
	const raw = String(value || '').trim();
	const withoutExt = raw.toLowerCase().endsWith('.' + ext) ? raw.slice(0, -(ext.length + 1)) : raw;
	const slug = slugify(withoutExt) || 'entrada';
	return slug + '.' + ext;
}

export function collectionByFolder(schema, path) {
	const clean = String(path || '').replace(/^\/+/, '');
	for (const collection of schema?.collections || []) {
		const folder = collectionFolder(collection);
		if (clean === folder || clean.startsWith(folder + '/')) return collection;
	}
	return null;
}

export function isInside(path, folder) {
	const clean = String(path || '').replace(/^\/+/, '');
	const base = String(folder || '').replace(/^\/+|\/+$/g, '');
	return clean === base || clean.startsWith(base + '/');
}
