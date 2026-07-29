import { writable } from 'svelte/store';
import { browser } from '$app/environment';

function createNetworkStore() {
	const online = writable<boolean>(browser ? navigator.onLine : true);

	if (browser) {
		window.addEventListener('online', () => online.set(true));
		window.addEventListener('offline', () => online.set(false));
	}

	return {
		subscribe: online.subscribe,
		isOnline: online
	};
}

export const network = createNetworkStore();
