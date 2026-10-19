const DATE_LINE = /^(\s*[A-Za-z0-9_.-]+:\s*)["'](\d{4}-\d{2}-\d{2})["']\s*$/gm;

function lib() {
	const yaml = globalThis.jsyaml;
	if (!yaml) throw new Error('js-yaml no está disponible');
	return yaml;
}

export function load(text) {
	if (!String(text || '').trim()) return {};
	const yaml = lib();
	const value = yaml.load(String(text), { schema: yaml.CORE_SCHEMA });
	if (value === null || value === undefined) return {};
	if (typeof value !== 'object' || Array.isArray(value)) {
		throw new Error('el YAML debe ser un objeto');
	}
	return value;
}

export function dump(value) {
	const yaml = lib();
	const text = yaml.dump(value || {}, {
		schema: yaml.CORE_SCHEMA,
		lineWidth: -1,
		noRefs: true,
		sortKeys: false,
	});
	return text.replace(DATE_LINE, '$1$2');
}

export function isEmpty(value) {
	if (value === null || value === undefined) return true;
	if (typeof value === 'string') return value.trim() === '';
	if (Array.isArray(value)) return value.length === 0;
	if (typeof value === 'object') return Object.keys(value).length === 0;
	return false;
}

export function clone(value) {
	if (value === undefined) return undefined;
	return JSON.parse(JSON.stringify(value));
}
