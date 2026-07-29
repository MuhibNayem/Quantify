const RECENT_KEY = 'quantify:recent';
const MAX = 8;

export interface RecentItem {
	type: 'product' | 'order' | 'supplier' | 'customer';
	id: number;
	label: string;
	href: string;
	timestamp: number;
}

export function addRecent(item: Omit<RecentItem, 'timestamp'>) {
	const all = getRecent().filter((r) => !(r.type === item.type && r.id === item.id));
	all.unshift({ ...item, timestamp: Date.now() });
	localStorage.setItem(RECENT_KEY, JSON.stringify(all.slice(0, MAX)));
}

export function getRecent(): RecentItem[] {
	try {
		return JSON.parse(localStorage.getItem(RECENT_KEY) || '[]');
	} catch {
		return [];
	}
}
