const PREFS_KEY = 'quantify:columnPrefs';

export interface ColumnDef {
	key: string;
	label: string;
	defaultVisible?: boolean;
}

export function loadColumnPrefs(tableId: string, columns: ColumnDef[]): Set<string> {
	try {
		const all = JSON.parse(localStorage.getItem(PREFS_KEY) || '{}');
		const saved = all[tableId];
		if (saved) return new Set(saved);
	} catch {}
	return new Set(columns.filter((c) => c.defaultVisible !== false).map((c) => c.key));
}

export function saveColumnPrefs(tableId: string, visible: Set<string>): void {
	try {
		const all = JSON.parse(localStorage.getItem(PREFS_KEY) || '{}');
		all[tableId] = [...visible];
		localStorage.setItem(PREFS_KEY, JSON.stringify(all));
	} catch {}
}
