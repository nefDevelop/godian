import { createProvider } from './provider.js';

const TOKEN_KEY = 'godian-admin-token';

export const state = {
	config: null,
	schema: null,
	options: null,
	provider: null,
	repo: null,
	token: '',
	files: new Map(),
	localUploads: new Map(),
};

async function loadJSON(path) {
	const response = await fetch(path, { cache: 'no-store', credentials: 'omit' });
	if (!response.ok) throw new Error('No se pudo cargar ' + path + ' (' + response.status + ')');
	return response.json();
}

export async function initState() {
	state.config = await loadJSON('./config.json');
	state.schema = await loadJSON('./schema.json');
	state.options = await loadJSON('./options.json');
	if (state.schema?.uploads_dir) state.config.uploads_dir = state.schema.uploads_dir;
	if (state.schema?.content_dir) state.config.content_dir = state.schema.content_dir;
	if (state.schema?.assets_dir) state.config.assets_dir = state.schema.assets_dir;
	const token = sessionStorage.getItem(TOKEN_KEY) || '';
	if (token) {
		state.token = token;
		state.provider = createProvider(state.config, token);
		try {
			state.repo = await state.provider.getRepo();
		} catch (err) {
			state.token = '';
			state.provider = null;
			state.repo = null;
			sessionStorage.removeItem(TOKEN_KEY);
		}
	}
	return state;
}

export async function login(token) {
	const provider = createProvider(state.config, token);
	const repo = await provider.getRepo();
	state.token = token;
	state.provider = provider;
	state.repo = repo;
	sessionStorage.setItem(TOKEN_KEY, token);
	return repo;
}

export function logout() {
	state.token = '';
	state.provider = null;
	state.repo = null;
	state.files.clear();
	state.localUploads.clear();
	sessionStorage.removeItem(TOKEN_KEY);
}

export function hasSession() {
	return Boolean(state.token && state.provider);
}
