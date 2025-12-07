import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import type { SvelteHTMLElements } from 'svelte/elements';

import { globalConfig } from '$lib/stores/settings';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

// ... (Types remain same) ...

// WithElementRef...
export type WithElementRef<T extends SvelteHTMLElements[keyof SvelteHTMLElements]> = T & {
	ref?: T['ref'];
};

export type WithoutChildrenOrChild<T> = Omit<T, 'children' | 'child'>;
export type WithoutChildren<T> = Omit<T, 'children' | 'child'>;
export type WithoutChild<T> = Omit<T, 'children' | 'child'>;

// Format date to readable string
export function formatDate(dateString: string): string {
	if (!dateString) return '';
	const date = new Date(dateString);
	return date.toLocaleDateString(globalConfig.locale, {
		year: 'numeric',
		month: 'short',
		day: 'numeric',
		// timeZone: globalConfig.timezone // Optional: Enforce server timezone or let it be user browser local
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
		currency: 'USD', // Fallback as we don't have code in settings yet
	}).format(amount).replace('$', globalConfig.currency_symbol);
}

export function formatDateTime(dateString: string): string {
	if (!dateString) return '';
	const date = new Date(dateString);
	return date.toLocaleString(globalConfig.locale, {
		year: 'numeric',
		month: 'short',
		day: 'numeric',
		hour: 'numeric',
		minute: 'numeric',
		// timeZone: globalConfig.timezone // Optional
	});
}

export function formatPercent(value: number): string {
	if (value === null || value === undefined || Number.isNaN(value)) return '';
	const normalized = value > 1 ? value / 100 : value;
	return new Intl.NumberFormat(globalConfig.locale, {
		style: 'percent',
		maximumFractionDigits: 1
	}).format(normalized);
}
