let katexPromise = null;

function escapeHTML(value) {
	return String(value).replace(/[&<>"]/g, (ch) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[ch]));
}

function preprocessCallouts(markdown) {
	const lines = String(markdown || '').split(/\r?\n/);
	const out = [];
	for (let i = 0; i < lines.length; i += 1) {
		const match = lines[i].match(/^>\s*\[!(\w+)\]\s*(.*)$/);
		if (!match) {
			out.push(lines[i]);
			continue;
		}
		const type = match[1].toLowerCase();
		const title = match[2] || type;
		const body = [];
		let j = i + 1;
		while (j < lines.length && /^>/.test(lines[j])) {
			body.push(lines[j].replace(/^>\s?/, ''));
			j += 1;
		}
		out.push('<div class="callout callout-' + type + '">');
		out.push('<div class="callout-title">' + escapeHTML(title) + '</div>');
		out.push(window.marked.parse(body.join('\n')));
		out.push('</div>');
		i = j - 1;
	}
	return out.join('\n');
}

export function renderMarkdown(markdown) {
	if (!window.marked || !window.DOMPurify) {
		return '<p class="admin-alert admin-alert-error">Preview no disponible: faltan marked/DOMPurify.</p>';
	}
	window.marked.setOptions({ gfm: true, breaks: false, headerIds: true, mangle: false });
	let source = preprocessCallouts(markdown);
	source = source.replace(/==([^=\n]+)==/g, '<mark>$1</mark>');
	const html = window.marked.parse(source);
	const clean = window.DOMPurify.sanitize(html, {
		ADD_ATTR: ['target', 'rel', 'loading', 'decoding'],
	});
	const doc = new DOMParser().parseFromString(clean, 'text/html');
	doc.querySelectorAll('a[href]').forEach((link) => {
		const href = link.getAttribute('href') || '';
		if (/^https?:\/\//i.test(href)) {
			link.setAttribute('target', '_blank');
			link.setAttribute('rel', 'noopener noreferrer');
		}
	});
	doc.querySelectorAll('img').forEach((img) => {
		img.setAttribute('loading', 'lazy');
		img.setAttribute('decoding', 'async');
	});
	return doc.body.innerHTML;
}

function loadKatex(base) {
	if (window.renderMathInElement) return Promise.resolve();
	if (katexPromise) return katexPromise;
	const vendor = String(base || '/assets/vendor/').replace(/\/?$/, '/');
	katexPromise = new Promise((resolve, reject) => {
		const style = document.createElement('link');
		style.rel = 'stylesheet';
		style.href = vendor + 'katex/katex.min.css';
		document.head.appendChild(style);
		const script = document.createElement('script');
		script.src = vendor + 'katex/katex.min.js';
		script.onload = () => {
			const extra = document.createElement('script');
			extra.src = vendor + 'katex/contrib/auto-render.min.js';
			extra.onload = () => resolve();
			extra.onerror = reject;
			document.head.appendChild(extra);
		};
		script.onerror = reject;
		document.head.appendChild(script);
	});
	return katexPromise;
}

export function renderMath(element, config) {
	if (!element) return;
	const text = element.textContent || '';
	if (!/[$]|\\\(|\\\[/.test(text)) return;
	loadKatex(config?.vendor_base).then(() => {
		if (!window.renderMathInElement) return;
		window.renderMathInElement(element, {
			delimiters: [
				{ left: '$$', right: '$$', display: true },
				{ left: '$', right: '$', display: false },
				{ left: '\\(', right: '\\)', display: false },
				{ left: '\\[', right: '\\]', display: true },
			],
			throwOnError: false,
			ignoredTags: ['script', 'noscript', 'style', 'textarea', 'pre', 'code', 'option'],
		});
	}).catch(() => {});
}

export function renderInto(element, markdown, config) {
	if (!element) return;
	element.innerHTML = renderMarkdown(markdown);
	renderMath(element, config);
}

function replaceSelection(textarea, before, after = '', placeholder = '') {
	const start = textarea.selectionStart || 0;
	const end = textarea.selectionEnd || 0;
	const selected = textarea.value.slice(start, end) || placeholder;
	const text = before + selected + after;
	textarea.value = textarea.value.slice(0, start) + text + textarea.value.slice(end);
	const cursor = start + before.length;
	textarea.selectionStart = cursor;
	textarea.selectionEnd = cursor + selected.length;
	textarea.focus();
	textarea.dispatchEvent(new Event('input', { bubbles: true }));
}

function prefixLines(textarea, prefix) {
	const start = textarea.selectionStart || 0;
	const end = textarea.selectionEnd || 0;
	const value = textarea.value;
	const lineStart = value.lastIndexOf('\n', start - 1) + 1;
	const block = value.slice(lineStart, end) || '';
	const lines = block.split('\n');
	const next = lines.map((line) => prefix + line).join('\n');
	textarea.value = value.slice(0, lineStart) + next + value.slice(end);
	textarea.selectionStart = lineStart;
	textarea.selectionEnd = lineStart + next.length;
	textarea.focus();
	textarea.dispatchEvent(new Event('input', { bubbles: true }));
}

export function applyToolbar(textarea, action) {
	if (!textarea) return;
	switch (action) {
		case 'bold': replaceSelection(textarea, '**', '**', 'texto'); break;
		case 'italic': replaceSelection(textarea, '*', '*', 'texto'); break;
		case 'strike': replaceSelection(textarea, '~~', '~~', 'texto'); break;
		case 'highlight': replaceSelection(textarea, '==', '==', 'texto'); break;
		case 'code-inline': replaceSelection(textarea, '`', '`', 'código'); break;
		case 'link': replaceSelection(textarea, '[', '](https://)', 'texto'); break;
		case 'image': replaceSelection(textarea, '![', '](/uploads/…)', 'alt'); break;
		case 'h1': prefixLines(textarea, '# '); break;
		case 'h2': prefixLines(textarea, '## '); break;
		case 'h3': prefixLines(textarea, '### '); break;
		case 'quote': prefixLines(textarea, '> '); break;
		case 'list': prefixLines(textarea, '- '); break;
		case 'ordered': prefixLines(textarea, '1. '); break;
		case 'code': replaceSelection(textarea, '```\n', '\n```', 'código'); break;
		case 'hr': replaceSelection(textarea, '\n---\n', '', ''); break;
		default: break;
	}
}

export function installShortcuts(textarea, onAction) {
	textarea.addEventListener('keydown', (event) => {
		const mod = event.metaKey || event.ctrlKey;
		if (!mod) return;
		const key = event.key.toLowerCase();
		const map = { b: 'bold', i: 'italic', k: 'link' };
		if (!map[key]) return;
		event.preventDefault();
		if (onAction) onAction(map[key]);
		else applyToolbar(textarea, map[key]);
	});
}

export function bindScrollSync(textarea, preview) {
	if (!textarea || !preview) return;
	let lock = false;
	const sync = (source, target) => {
		if (lock) return;
		lock = true;
		const sourceMax = source.scrollHeight - source.clientHeight;
		const targetMax = target.scrollHeight - target.clientHeight;
		if (sourceMax > 0) target.scrollTop = (source.scrollTop / sourceMax) * targetMax;
		requestAnimationFrame(() => { lock = false; });
	};
	textarea.addEventListener('scroll', () => sync(textarea, preview));
	preview.addEventListener('scroll', () => sync(preview, textarea));
}
