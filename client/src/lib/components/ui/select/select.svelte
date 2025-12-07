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
				'absolute left-3 text-gray-600 transition-all duration-200 ease-in-out',
				value || isOpen ? '-top-2 bg-white px-1 text-xs' : 'top-1/2 -translate-y-1/2 text-base',
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
			'flex h-12 w-full items-center justify-between rounded-md border bg-white px-4 py-2 text-left',
			'focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500',
			'transition-all duration-200 ease-in-out',
			'shadow-sm', // Material UI-like shadow
			'appearance-none', // Reset default button styles for cross-browser consistency
			'-webkit-appearance: none;', // For Safari
			'background-color: transparent;', // Ensure background is controlled by Tailwind
			'border: none;', // Ensure border is controlled by Tailwind
			'cursor: pointer;', // Explicitly set cursor
			disabled && 'cursor-not-allowed bg-gray-100 opacity-50',
			value && 'pt-4' // Adjust padding when value is selected and label floats
		)}
		onclick={toggleDropdown}
		aria-haspopup="listbox"
		aria-expanded={isOpen}
		{disabled}
	>
		<span class={cn(value ? 'text-gray-800' : 'text-gray-500')}>
			{selectedLabel}
		</span>
		<svg
			class={cn('h-4 w-4 transition-transform duration-200', isOpen && 'rotate-180')}
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
				'absolute z-10 max-h-60 w-full overflow-auto rounded-md border border-gray-300 bg-white shadow-lg',
				dropdownPosition === 'top' ? 'bottom-full mb-1' : 'top-full mt-1'
			)}
			transitionfly={{ y: dropdownPosition === 'top' ? 10 : -10, duration: 150 }}
			role="listbox"
		>
			{#each options as option (option.value)}
				<li
					class={cn(
						'cursor-pointer px-4 py-2 hover:bg-blue-50 hover:text-blue-700',
						value === option.value && 'bg-blue-100 font-medium text-blue-800'
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
