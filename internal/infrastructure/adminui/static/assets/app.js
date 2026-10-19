import { clear, errorText, h, icon, loading, toast } from './ui.js';
import { hasSession, initState, login, logout, state } from './state.js';
import { renderDashboard } from './dashboard.js';
import { renderEditor } from './editor.js';
import { renderMediaPage } from './media.js';

const app = document.getElementById('app');

function themeToggle() {
	return h('button', {
		type: 'button',
		class: 'admin-theme-toggle',
		'aria-label': 'Cambiar tema',
		title: 'Cambiar tema',
		onclick: () => {
			const root = document.documentElement;
			const next = root.classList.contains('dark') ? 'light' : 'dark';
			root.classList.remove('light', 'dark');
			root.classList.add(next);
			try { localStorage.setItem('theme', next); } catch (err) {}
		},
	}, icon('moon'));
}

function tokenHelpURL() {
	if (state.config.provider === 'github') return 'https://github.com/settings/tokens?type=beta';
	const base = String(state.config.base_url || '').replace(/\/+$/, '');
	return base + '/user/settings/applications';
}

function renderLogin(message = '') {
	clear(app);
	const tokenInput = h('input', {
		type: 'password',
		id: 'admin-token',
		autocomplete: 'off',
		spellcheck: 'false',
		placeholder: state.config.provider === 'github' ? 'github_pat_…' : 'token de Gitea',
	});
	const error = h('div', { class: 'admin-alert admin-alert-error', hidden: true }, message);
	if (message) error.hidden = false;
	const submit = h('button', { type: 'submit', class: 'admin-button admin-button-primary' }, 'Conectar');
	const form = h('form', {
		onsubmit: async (event) => {
			event.preventDefault();
			const token = tokenInput.value.trim();
			if (!token) return;
			submit.disabled = true;
			submit.textContent = 'Conectando…';
			error.hidden = true;
			try {
				await login(token);
				renderApp();
			} catch (err) {
				error.textContent = errorText(err);
				error.hidden = false;
				submit.disabled = false;
				submit.textContent = 'Conectar';
			}
		},
	},
		h('div', { class: 'admin-field' },
			h('label', { for: 'admin-token' }, 'Token de acceso'),
			tokenInput,
			h('p', { class: 'admin-hint' },
				'Se guarda solo en esta pestaña (sessionStorage). ',
				h('a', { class: 'admin-link', href: tokenHelpURL(), target: '_blank', rel: 'noopener' }, 'Crear token'),
			),
		),
		error,
		submit,
	);
	const card = h('div', { class: 'admin-login' },
		h('div', { class: 'admin-login-card' },
			h('p', { class: 'admin-eyebrow' }, state.config.provider === 'github' ? 'GitHub' : 'Gitea'),
			h('h1', {}, state.config.site_title || 'Admin'),
			h('p', {}, 'Repositorio ' + state.config.owner + '/' + state.config.repo + ' · rama ' + state.config.branch),
			form,
		),
	);
	app.appendChild(card);
}

function renderApp() {
	clear(app);
	const navigate = (hash) => {
		if (location.hash === hash) route();
		else location.hash = hash;
	};
	const brand = h('a', { class: 'admin-brand', href: '#/', onclick: (event) => { event.preventDefault(); navigate('#/'); } },
		h('span', { class: 'admin-brand-mark' }, icon('dashboard')),
		h('span', { class: 'admin-brand-title' }, state.config.site_title || 'godian'),
		h('span', { class: 'admin-brand-sub' }, 'Admin'),
	);
	const repoLink = h('a', { href: state.provider.repoURL(), target: '_blank', rel: 'noopener' }, state.config.owner + '/' + state.config.repo);
	const nav = h('nav', { class: 'admin-nav', 'aria-label': 'Admin' },
		h('a', { href: '#/', onclick: (event) => { event.preventDefault(); navigate('#/'); } }, 'Contenido'),
		h('a', { href: '#/media', onclick: (event) => { event.preventDefault(); navigate('#/media'); } }, 'Medios'),
		h('a', { href: '../', target: '_blank', rel: 'noopener' }, 'Ver sitio'),
		repoLink,
		h('button', {
			type: 'button',
			class: 'admin-icon-button',
			title: 'Cerrar sesión',
			'aria-label': 'Cerrar sesión',
			onclick: () => { logout(); renderLogin(); },
		}, icon('logout')),
		themeToggle(),
	);
	const header = h('header', { class: 'admin-header' }, brand, nav);
	const main = h('main', { class: 'admin-main' });
	app.append(header, main);

	const route = async () => {
		if (!hasSession()) {
			renderLogin();
			return;
		}
		const raw = location.hash.replace(/^#/, '') || '/';
		const [path, query = ''] = raw.split('?');
		const params = new URLSearchParams(query);
		clear(main);
		main.appendChild(loading('Cargando…'));
		try {
			if (path === '/' || path === '') {
				await renderDashboard(main, state, navigate);
			} else if (path === '/edit') {
				await renderEditor(main, state, params, navigate, false);
			} else if (path === '/new') {
				await renderEditor(main, state, params, navigate, true);
			} else if (path === '/media') {
				await renderMediaPage(main, state);
			} else {
				await renderDashboard(main, state, navigate);
			}
		} catch (err) {
			clear(main);
			main.appendChild(h('div', { class: 'admin-alert admin-alert-error' }, errorText(err)));
		}
	};

	window.addEventListener('hashchange', route);
	route();
}

async function boot() {
	clear(app);
	app.appendChild(loading('Iniciando admin…'));
	try {
		await initState();
		if (hasSession()) renderApp();
		else renderLogin();
	} catch (err) {
		clear(app);
		app.appendChild(h('div', { class: 'admin-alert admin-alert-error' },
			'No se pudo iniciar el admin: ' + errorText(err) + '. Revisa admin/config.json y el proxy /admin/api.',
		));
	}
}

boot();
