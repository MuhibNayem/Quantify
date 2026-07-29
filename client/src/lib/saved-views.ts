const VIEWS_KEY = 'quantify:savedViews';

export function saveView(page: string, filters: Record<string, unknown>) {
	const all = JSON.parse(localStorage.getItem(VIEWS_KEY) || '{}');
	all[page] = filters;
	localStorage.setItem(VIEWS_KEY, JSON.stringify(all));
}

export function loadView<T>(page: string, defaults: T): T {
	const all = JSON.parse(localStorage.getItem(VIEWS_KEY) || '{}');
	return { ...defaults, ...(all[page] || {}) };
}
