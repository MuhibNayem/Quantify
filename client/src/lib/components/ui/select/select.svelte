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
	}: {
		value?: string;
		options?: { value: string; label: string }[];
		placeholder?: string;
		label?: string;
		disabled?: boolean;
		style?: string;
	} = $props();

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
				'absolute left-3 z-10 font-medium text-foreground transition-all duration-200 ease-in-out',
				value || isOpen
					? '-top-2 rounded-sm bg-card px-1 text-xs'
					: 'top-1/2 -translate-y-1/2 text-base',
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
			'flex h-12 w-full items-center justify-between rounded-md border border-input bg-card px-4 text-left text-sm font-medium text-foreground shadow-xs outline-none transition-all',
			'focus:border-primary focus:ring-[3px] focus:ring-ring/30',
			'appearance-none',
			'-webkit-appearance: none;',
			'cursor: pointer;',
			disabled && 'cursor-not-allowed bg-muted opacity-50',
			value && label && 'pt-2'
		)}
		onclick={toggleDropdown}
		aria-haspopup="listbox"
		aria-expanded={isOpen}
		{disabled}
	>
		<span class={cn(value ? 'text-foreground' : 'text-muted-foreground')}>
			{selectedLabel}
		</span>
		<svg
			class={cn('h-4 w-4 text-muted-foreground transition-transform duration-300', isOpen && 'rotate-180')}
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
				'custom-scrollbar absolute z-10 max-h-60 w-full overflow-auto rounded-md border border-border bg-card p-1 shadow-lg',
				dropdownPosition === 'top' ? 'bottom-full mb-2' : 'top-full mt-2'
			)}
			transition:fly={{ y: dropdownPosition === 'top' ? 10 : -10, duration: 150 }}
			role="listbox"
		>
			{#each options as option (option.value)}
				<li
					class={cn(
						'cursor-pointer rounded-md px-3 py-2.5 text-sm font-medium text-foreground transition-colors hover:bg-muted',
						value === option.value && 'bg-accent text-accent-foreground'
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
