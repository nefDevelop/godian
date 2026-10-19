try {
	var theme = localStorage.getItem('theme');
	if (theme === 'light' || theme === 'dark') document.documentElement.classList.add(theme);
} catch (err) {}
