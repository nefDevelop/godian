// Proxy same-origin del admin estático hacia Gitea/GitHub.
// No guarda ni inyecta tokens: reenvía el Authorization del navegador.
// Variables de entorno de Cloudflare Pages: GIT_PROVIDER, GIT_BASE_URL,
// GIT_OWNER, GIT_REPO.
const ALLOWED = new Set(['GET', 'PUT', 'DELETE', 'OPTIONS']);

export async function onRequest(context) {
	const { request, env } = context;
	const method = request.method.toUpperCase();
	if (!ALLOWED.has(method)) {
		return new Response('Method Not Allowed', { status: 405 });
	}
	if (method === 'OPTIONS') {
		return new Response(null, { status: 204, headers: { Allow: 'GET, PUT, DELETE, OPTIONS' } });
	}

	const provider = String(env.GIT_PROVIDER || 'gitea').toLowerCase();
	const owner = String(env.GIT_OWNER || '');
	const repo = String(env.GIT_REPO || '');
	const baseURL = String(env.GIT_BASE_URL || '').replace(/\/+$/, '');
	if (!owner || !repo || (provider === 'gitea' && !baseURL)) {
		return new Response('Admin proxy sin configurar (GIT_OWNER/GIT_REPO/GIT_BASE_URL)', { status: 500 });
	}
	const base = provider === 'github' ? 'https://api.github.com' : baseURL + '/api/v1';

	const url = new URL(request.url);
	const path = url.pathname.replace(/^\/admin\/api/, '');
	if (!path.startsWith('/repos/')) {
		return new Response('Not Found', { status: 404 });
	}
	const rest = path.slice('/repos/'.length);
	const parts = rest.split('/');
	if (parts.length < 2) {
		return new Response('Not Found', { status: 404 });
	}
	let requestOwner;
	let requestRepo;
	try {
		requestOwner = decodeURIComponent(parts[0]);
		requestRepo = decodeURIComponent(parts[1]);
	} catch (err) {
		return new Response('Bad Request', { status: 400 });
	}
	if (requestOwner !== owner || requestRepo !== repo) {
		return new Response('Forbidden', { status: 403 });
	}
	const tail = parts.slice(2).join('/');
	if (tail && tail !== 'contents' && !tail.startsWith('contents/')) {
		return new Response('Forbidden', { status: 403 });
	}
	if (tail.split('/').some((segment) => segment === '..' || segment === '.')) {
		return new Response('Forbidden', { status: 403 });
	}
	const target = new URL(base + '/repos/' + encodeURIComponent(owner) + '/' + encodeURIComponent(repo) + (tail ? '/' + tail : ''));
	target.search = url.search;

	const headers = new Headers();
	for (const key of ['authorization', 'content-type', 'accept', 'x-github-api-version']) {
		const value = request.headers.get(key);
		if (value) headers.set(key, value);
	}
	headers.set('user-agent', 'godian-admin-pages-proxy');

	const upstream = await fetch(target.toString(), {
		method,
		headers,
		body: method === 'GET' ? undefined : request.body,
		redirect: 'manual',
	});
	const outHeaders = new Headers(upstream.headers);
	outHeaders.delete('content-length');
	outHeaders.delete('connection');
	outHeaders.set('cache-control', 'no-store');
	return new Response(upstream.body, { status: upstream.status, headers: outHeaders });
}
