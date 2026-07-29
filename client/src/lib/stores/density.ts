import { writable } from 'svelte/store';

export type Density = 'comfortable' | 'compact';

function createDensityStore() {
	const stored = (typeof localStorage !== 'undefined' &&
		localStorage.getItem('quantify:density')) as Density | null;
	let current: Density = stored || 'comfortable';
	const { subscribe, set } = writable<Density>(current);

	if (typeof document !== 'undefined') {
		document.documentElement.dataset.density = current;
	}

	return {
		subscribe,
		toggle: () => {
			current = current === 'comfortable' ? 'compact' : 'comfortable';
			localStorage.setItem('quantify:density', current);
			if (typeof document !== 'undefined') {
				document.documentElement.dataset.density = current;
			}
			set(current);
		}
	};
}

export const density = createDensityStore();
