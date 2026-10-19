import { basename, extname, uploadsURL } from './paths.js';

export class ProviderError extends Error {
	constructor(status, message, body) {
		super(message);
		this.name = 'ProviderError';
		this.status = status;
		this.body = body;
	}
}

function b64encode(text) {
	const bytes = new TextEncoder().encode(text);
	let binary = '';
	const chunk = 0x8000;
	for (let i = 0; i < bytes.length; i += chunk) {
		binary += String.fromCharCode(...bytes.subarray(i, i + chunk));
	}
	return btoa(binary);
}

function b64decode(value) {
	const clean = String(value || '').replace(/\s+/g, '');
	const binary = atob(clean);
	const bytes = new Uint8Array(binary.length);
	for (let i = 0; i < binary.length; i += 1) bytes[i] = binary.charCodeAt(i);
	return new TextDecoder().decode(bytes);
}

function encodePath(path) {
	return String(path || '')
		.split('/')
		.filter((part) => part !== '')
		.map((part) => encodeURIComponent(part))
		.join('/');
}

export function createProvider(config, token) {
	if (!token) throw new ProviderError(401, 'Falta el token');
	if (config.provider === 'github') return new GitHubProvider(config, token);
	if (config.provider === 'gitea') return new GiteaProvider(config, token);
	throw new ProviderError(400, 'Proveedor no soportado: ' + config.provider);
}

class GitProvider {
	constructor(config, token, kind) {
		this.config = config;
		this.token = token;
		this.kind = kind;
		this.owner = config.owner;
		this.repo = config.repo;
		this.branch = config.branch || 'main';
	}

	get apiBase() {
		let base = this.config.api_base || (this.config.api_mode === 'proxy' ? 'api' : '');
		if (!base) base = this.kind === 'github' ? 'https://api.github.com' : (this.config.base_url || '') + '/api/v1';
		return String(base).replace(/\/+$/, '');
	}

	get displayName() {
		return this.config.provider === 'github' ? 'GitHub' : 'Gitea';
	}

	repoURL() {
		if (this.kind === 'github') return 'https://github.com/' + encodeURIComponent(this.owner) + '/' + encodeURIComponent(this.repo);
		return String(this.config.base_url || '').replace(/\/+$/, '') + '/' + encodeURIComponent(this.owner) + '/' + encodeURIComponent(this.repo);
	}

	headers(extra = {}) {
		const auth = this.kind === 'github' ? 'Bearer ' + this.token : 'token ' + this.token;
		const headers = {
			Authorization: auth,
			Accept: this.kind === 'github' ? 'application/vnd.github+json' : 'application/json',
			...extra,
		};
		if (this.kind === 'github') headers['X-GitHub-Api-Version'] = '2022-11-28';
		return headers;
	}

	async request(path, options = {}) {
		const controller = new AbortController();
		const timer = setTimeout(() => controller.abort(), 30000);
		const method = options.method || 'GET';
		const headers = this.headers(options.headers || {});
		let body;
		if (options.body !== undefined && options.body !== null) {
			headers['Content-Type'] = 'application/json';
			body = JSON.stringify(options.body);
		}
		let response;
		try {
			response = await fetch(this.apiBase + path, {
				method,
				headers,
				body,
				credentials: 'omit',
				cache: 'no-store',
				signal: controller.signal,
			});
		} catch (err) {
			throw new ProviderError(0, 'No se pudo conectar con ' + this.displayName + '. Revisa la red, el proxy /admin/api o CORS.', null);
		} finally {
			clearTimeout(timer);
		}

		let data = null;
		const type = response.headers.get('content-type') || '';
		try {
			data = type.includes('json') ? await response.json() : await response.text();
		} catch (err) {
			data = null;
		}
		if (!response.ok) {
			let message = data?.message || data?.error || response.statusText || 'Error del proveedor';
			if (response.status === 401) message = 'Token inválido o sin permisos.';
			else if (response.status === 403) message = 'Sin permisos para escribir en el repositorio. Revisa el scope del token.';
			else if (response.status === 409) message = 'El archivo cambió remotamente. Recarga antes de guardar.';
			else if (response.status === 404) message = 'No encontrado en el repositorio.';
			throw new ProviderError(response.status, message, data);
		}
		return data;
	}

	repoPath(suffix = '') {
		return '/repos/' + encodeURIComponent(this.owner) + '/' + encodeURIComponent(this.repo) + suffix;
	}

	async getRepo() {
		return this.request(this.repoPath());
	}

	async listDir(path, branch = this.branch) {
		const ref = '?ref=' + encodeURIComponent(branch);
		try {
			const data = await this.request(this.repoPath('/contents/' + encodePath(path) + ref));
			if (Array.isArray(data)) return data;
			if (data && data.type) return [data];
			return [];
		} catch (err) {
			if (err.status === 404) return [];
			throw err;
		}
	}

	async listFiles(folder, extension, branch = this.branch) {
		const out = [];
		const walk = async (dir) => {
			const entries = await this.listDir(dir, branch);
			for (const entry of entries) {
				const type = entry.type === 'dir' || entry.type === 'directory';
				if (type) {
					await walk(entry.path);
				} else if (!extension || extname(entry.path) === '.' + extension) {
					out.push({
						path: entry.path,
						name: entry.name || basename(entry.path),
						sha: entry.sha,
						size: entry.size,
					});
				}
			}
		};
		await walk(folder);
		return out;
	}

	async readFile(path, branch = this.branch) {
		const data = await this.request(this.repoPath('/contents/' + encodePath(path) + '?ref=' + encodeURIComponent(branch)));
		if (Array.isArray(data)) throw new ProviderError(400, 'La ruta es un directorio');
		const content = b64decode(data.content || '');
		return { path: data.path || path, content, sha: data.sha, name: data.name || basename(path) };
	}

	async writeFile(path, content, sha, message, branch = this.branch) {
		const body = {
			message,
			content: b64encode(content),
			branch,
		};
		if (sha) body.sha = sha;
		return this.request(this.repoPath('/contents/' + encodePath(path)), {
			method: 'PUT',
			body,
		});
	}

	async deleteFile(path, sha, message, branch = this.branch) {
		return this.request(this.repoPath('/contents/' + encodePath(path)), {
			method: 'DELETE',
			body: { message, sha, branch },
		});
	}

	async uploadBinary(path, base64, message, branch = this.branch) {
		return this.request(this.repoPath('/contents/' + encodePath(path)), {
			method: 'PUT',
			body: { message, content: base64, branch },
		});
	}

	async listUploads(branch = this.branch) {
		const folder = this.config.uploads_dir || 'assets/uploads';
		const entries = await this.listDir(folder, branch);
		return entries
			.filter((entry) => entry.type !== 'dir' && entry.type !== 'directory')
			.map((entry) => ({
				name: entry.name || basename(entry.path),
				path: entry.path,
				sha: entry.sha,
				size: entry.size,
				url: uploadsURL(this.config, entry.name || basename(entry.path)),
			}));
	}

	commitURL(data) {
		return data?.commit?.html_url || data?.commit?.url || data?.commit_url || '';
	}

	commitSHA(data) {
		return data?.commit?.sha || data?.commit?.id || '';
	}
}

class GitHubProvider extends GitProvider {
	constructor(config, token) {
		super(config, token, 'github');
	}
}

class GiteaProvider extends GitProvider {
	constructor(config, token) {
		super(config, token, 'gitea');
	}
}
