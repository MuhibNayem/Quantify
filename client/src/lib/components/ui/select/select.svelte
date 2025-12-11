<!-- client/src/lib/components/ui/select/select.svelte -->
<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { fly } from 'svelte/transition';
	import { cn } from '$lib/utils';
	import { browser } from '$app/environment';

	const dispatch = createEventDispatcher();

	let {
		value = $bindable(''),
		options = [],
		placeholder = 'Select an option',
		label = '',
		disabled = false,
		style = ''
	} = $props<{
		value?: string;
		options?: { value: string; label: string }[];
		placeholder?: string;
		label?: string;
		disabled?: boolean;
		style?: string;
	}>();

	let isOpen = $state(false);
	let selectRef: HTMLElement | null = null;
	let triggerRef: HTMLButtonElement | null = null;
	let dropdownPosition = $state('bottom');

	const toggleDropdown = () => {
		if (!disabled) {
			isOpen = !isOpen;
		}
	};

	const selectOption = (optionValue: string) => {
		value = optionValue;
		isOpen = false;
		dispatch('change', value);
	};

	const selectedLabel = $derived(options.find((o) => o.value === value)?.label || placeholder);

	const checkPosition = () => {
		if (triggerRef && isOpen) {
			const rect = triggerRef.getBoundingClientRect();
			const spaceBelow = window.innerHeight - rect.bottom;
			// If less than 250px below, and more space above, flip to top
			if (spaceBelow < 250 && rect.top > 250) {
				dropdownPosition = 'top';
			} else {
				dropdownPosition = 'bottom';
			}
		}
	};

	$effect(() => {
		if (!browser || !isOpen) {
			return;
		}

		checkPosition();

		const handleClickOutside = (event: MouseEvent) => {
			if (selectRef && !selectRef.contains(event.target as Node)) {
				isOpen = false;
			}
		};

		const handleResizeScroll = () => {
			checkPosition();
		};

		document.addEventListener('click', handleClickOutside);
		window.addEventListener('resize', handleResizeScroll);
		window.addEventListener('scroll', handleResizeScroll, true);

		return () => {
			document.removeEventListener('click', handleClickOutside);
			window.removeEventListener('resize', handleResizeScroll);
			window.removeEventListener('scroll', handleResizeScroll, true);
		};
	});
</script>

<div class={cn('relative w-full', style)} bind:this={selectRef}>
	<!-- Floating Label -->
	{#if label}
		<label
			class={cn(
				'absolute left-3 z-10 font-bold text-slate-800 transition-all duration-200 ease-in-out',
				value || isOpen
					? '-top-2 rounded-full bg-white/80 px-1 text-xs backdrop-blur-sm'
					: 'top-1/2 -translate-y-1/2 text-lg',
				disabled && 'opacity-50'
			)}
		>
			{label}
		</label>
	{/if}

	<!-- Select Trigger -->
	<button
		bind:this={triggerRef}
		type="button"
		class={cn(
			'flex h-12 w-full items-center justify-between rounded-2xl border border-white/60 bg-gradient-to-br from-white/90 to-white/60 px-5 text-left text-lg font-bold text-slate-900 shadow-[inset_0_2px_4px_rgba(255,255,255,0.9),0_4px_10px_rgba(0,0,0,0.05)] outline-none ring-1 ring-white/40 backdrop-blur-xl backdrop-saturate-[180%] transition-all',
			'focus:scale-[1.01] focus:border-blue-500/30 focus:ring-4 focus:ring-blue-500/10',
			'appearance-none',
			'-webkit-appearance: none;',
			'cursor: pointer;',
			disabled && 'cursor-not-allowed bg-gray-100 opacity-50',
			value && label && 'pt-2'
		)}
		onclick={toggleDropdown}
		aria-haspopup="listbox"
		aria-expanded={isOpen}
		{disabled}
	>
		<span class={cn(value ? 'text-slate-900' : 'text-slate-500/80')}>
			{selectedLabel}
		</span>
		<svg
			class={cn('h-5 w-5 text-slate-600 transition-transform duration-300', isOpen && 'rotate-180')}
			fill="none"
			stroke="currentColor"
			viewBox="0 0 24 24"
			xmlns="http://www.w3.org/2000/svg"
		>
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"
			></path>
		</svg>
	</button>

	<!-- Dropdown Options -->
	{#if isOpen}
		<ul
			class={cn(
				'absolute z-10 max-h-60 w-full overflow-auto rounded-xl border border-white/60 bg-white/90 shadow-[0_20px_40px_-10px_rgba(0,0,0,0.1),inset_0_1px_0_0_rgba(255,255,255,0.8)] ring-1 ring-white/40 backdrop-blur-2xl',
				dropdownPosition === 'top' ? 'bottom-full mb-2' : 'top-full mt-2'
			)}
			transition:fly={{ y: dropdownPosition === 'top' ? 10 : -10, duration: 150 }}
			role="listbox"
		>
			{#each options as option (option.value)}
				<li
					class={cn(
						'cursor-pointer px-4 py-3 font-bold text-slate-700 transition-colors hover:bg-emerald-500/10 hover:text-emerald-700',
						value === option.value && 'bg-emerald-500/10 text-emerald-800'
					)}
					onclick={() => selectOption(option.value)}
					role="option"
					aria-selected={value === option.value}
				>
					{option.label}
				</li>
			{/each}
		</ul>
	{/if}
</div>
