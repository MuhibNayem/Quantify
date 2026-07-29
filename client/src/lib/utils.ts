import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import type { SvelteHTMLElements } from 'svelte/elements';

import { globalConfig } from '$lib/stores/settings';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

// WithElementRef...
export type WithElementRef<T extends SvelteHTMLElements[keyof SvelteHTMLElements]> = T & {
	ref?: T['ref'];
};

export type WithoutChildrenOrChild<T> = Omit<T, 'children' | 'child'>;
export type WithoutChildren<T> = Omit<T, 'children' | 'child'>;
export type WithoutChild<T> = Omit<T, 'children' | 'child'>;

// Format date to readable string
export function formatDate(dateString: string | Date): string {
	if (!dateString) return '';
	const date = new Date(dateString);
	return date.toLocaleDateString(globalConfig.locale, {
		year: 'numeric',
		month: 'short',
		day: 'numeric',
		timeZone: globalConfig.timezone
	});
}

// Format number as currency
export function formatCurrency(amount: number): string {
	// We can use Intl.NumberFormat if we had a proper currency Code (USD, EUR).
	// The settings give us a Symbol ($).
	// If backend returns 'currency_symbol' like '$', we can just prepend it or use a mapping.
	// For now, let's assume 'USD' style formatting but replace symbol if needed,
	// OR if we store Currency CODE in backend? Backend seeded 'currency_symbol' = '$'.
	// Custom formatting:
	return new Intl.NumberFormat(globalConfig.locale, {
		style: 'currency',
		currency: globalConfig.currency_code || 'USD'
	}).format(amount);
}

export function formatDateTime(dateString?: string | null): string {
	if (!dateString) return '';
	const date = new Date(dateString);
	return date.toLocaleString(globalConfig.locale, {
		year: 'numeric',
		month: 'short',
		day: 'numeric',
		hour: 'numeric',
		minute: 'numeric',
		timeZone: globalConfig.timezone
	});
}

export function formatPercent(value: number | null | undefined): string {
	if (value === null || value === undefined || Number.isNaN(value)) return '';
	const normalized = value > 1 ? value / 100 : value;
	return new Intl.NumberFormat(globalConfig.locale, {
		style: 'percent',
		maximumFractionDigits: 1
	}).format(normalized);
}

export function debounce<T extends (...args: any[]) => void>(
	func: T,
	wait: number
): (...args: Parameters<T>) => void {
	let timeout: NodeJS.Timeout;
	return (...args: Parameters<T>) => {
		clearTimeout(timeout);
		timeout = setTimeout(() => func(...args), wait);
	};
}

// Human-readable identifiers — users remember names, not IDs
export function orderNumber(order: { OrderNumber?: string; ID: number }): string {
	return order.OrderNumber || `ORD-${String(order.ID).padStart(6, '0')}`;
}

export function returnNumber(id: number): string {
	return `RET-${String(id).padStart(6, '0')}`;
}

export function poNumber(id: number): string {
	return `PO-${String(id).padStart(6, '0')}`;
}

// Relative time — "2h ago" instead of raw ISO
export function timeAgo(dateStr: string | Date): string {
	const date = typeof dateStr === 'string' ? new Date(dateStr) : dateStr;
	const seconds = Math.floor((Date.now() - date.getTime()) / 1000);
	if (seconds < 60) return 'just now';
	const minutes = Math.floor(seconds / 60);
	if (minutes < 60) return `${minutes}m ago`;
	const hours = Math.floor(minutes / 60);
	if (hours < 24) return `${hours}h ago`;
	const days = Math.floor(hours / 24);
	if (days < 7) return `${days}d ago`;
	const weeks = Math.floor(days / 7);
	if (weeks < 5) return `${weeks}w ago`;
	return formatDate(dateStr); // fall back to full date
}

// Truncate long names with ellipsis for tight layouts
export function truncate(str: string, max: number): string {
	return str.length > max ? str.slice(0, max - 1) + '…' : str;
}
