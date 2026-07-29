import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import type { AxiosError, AxiosInstance } from 'axios';

export interface QueuedMutation {
	id: string;
	method: 'POST' | 'PUT' | 'PATCH' | 'DELETE';
	url: string;
	data?: unknown;
	headers?: Record<string, string>;
	timestamp: number;
	retries: number;
}

const DB_NAME = 'quantify-offline';
const STORE = 'mutations';

export const pendingCount = writable(0);

function openDB(): Promise<IDBDatabase> {
	return new Promise((resolve, reject) => {
		const req = indexedDB.open(DB_NAME, 1);
		req.onupgradeneeded = () => {
			req.result.createObjectStore(STORE, { keyPath: 'id' });
		};
		req.onsuccess = () => resolve(req.result);
		req.onerror = () => reject(req.error);
	});
}

export async function enqueue(
	mutation: Omit<QueuedMutation, 'id' | 'timestamp' | 'retries'>
): Promise<void> {
	const db = await openDB();
	const entry: QueuedMutation = {
		...mutation,
		id: crypto.randomUUID(),
		timestamp: Date.now(),
		retries: 0
	};
	return new Promise((resolve, reject) => {
		const tx = db.transaction(STORE, 'readwrite');
		tx.objectStore(STORE).add(entry);
		tx.oncomplete = () => {
			updateCount();
			resolve();
		};
		tx.onerror = () => reject(tx.error);
	});
}

export async function getAll(): Promise<QueuedMutation[]> {
	const db = await openDB();
	return new Promise((resolve, reject) => {
		const req = db.transaction(STORE).objectStore(STORE).getAll();
		req.onsuccess = () =>
			resolve(req.result.sort((a: QueuedMutation, b: QueuedMutation) => a.timestamp - b.timestamp));
		req.onerror = () => reject(req.error);
	});
}

export async function remove(id: string): Promise<void> {
	const db = await openDB();
	return new Promise((resolve) => {
		const tx = db.transaction(STORE, 'readwrite');
		tx.objectStore(STORE).delete(id);
		tx.oncomplete = () => {
			updateCount();
			resolve();
		};
	});
}

export async function updateCount(): Promise<void> {
	if (!browser) return;
	const all = await getAll();
	pendingCount.set(all.length);
}

export async function syncQueue(
	axiosInstance: AxiosInstance
): Promise<{ synced: number; failed: number }> {
	const queue = await getAll();
	let synced = 0;
	let failed = 0;
	for (const mutation of queue) {
		try {
			await axiosInstance.request({
				method: mutation.method,
				url: mutation.url,
				data: mutation.data,
				headers: mutation.headers
			});
			await remove(mutation.id);
			synced++;
		} catch (e) {
			const status = (e as AxiosError)?.response?.status;
			if (status && status >= 400 && status < 500 && status !== 408 && status !== 429) {
				await remove(mutation.id);
				failed++;
			} else {
				failed++;
			}
		}
	}
	await updateCount();
	return { synced, failed };
}

if (browser) updateCount();
