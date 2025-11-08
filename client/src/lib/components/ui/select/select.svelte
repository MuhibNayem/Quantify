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
		style = '',
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

	const selectedLabel = $derived(() => options.find((o) => o.value === value)?.label || placeholder);

	$effect(() => {
		if (!browser || !isOpen) {
			return;
		}

		const handleClickOutside = (event: MouseEvent) => {
			if (selectRef && !selectRef.contains(event.target as Node)) {
				isOpen = false;
			}
		};

		document.addEventListener('click', handleClickOutside);
		return () => document.removeEventListener('click', handleClickOutside);
	});
</script>

<div
    class={cn(
        "relative w-full",
        style
    )}
    bind:this={selectRef}
>
    <!-- Floating Label -->
    {#if label}
        <label
            class={cn(
                "absolute left-3 transition-all duration-200 ease-in-out text-gray-600",
                value || isOpen ? "-top-2 text-xs bg-white px-1" : "top-1/2 -translate-y-1/2 text-base",
                disabled && "opacity-50"
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
            "flex justify-between items-center w-full h-12 px-4 py-2 border rounded-md bg-white text-left",
            "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500",
            "transition-all duration-200 ease-in-out",
            "shadow-sm", // Material UI-like shadow
            disabled && "opacity-50 cursor-not-allowed bg-gray-100",
            value && "pt-4" // Adjust padding when value is selected and label floats
        )}
        onclick={toggleDropdown}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
        {disabled}
    >
        <span class={cn(value ? "text-gray-800" : "text-gray-500")}>
            {selectedLabel}
        </span>
        <svg
            class={cn("w-4 h-4 transition-transform duration-200", isOpen && "rotate-180")}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            xmlns="http://www.w3.org/2000/svg"
        >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
        </svg>
    </button>

    <!-- Dropdown Options -->
    {#if isOpen}
        <ul
            class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-auto"
            transitionfly={{ y: -10, duration: 150 }}
            role="listbox"
        >
            {#each options as option (option.value)}
                <li
                    class={cn(
                        "px-4 py-2 cursor-pointer hover:bg-blue-50 hover:text-blue-700",
                        value === option.value && "bg-blue-100 text-blue-800 font-medium"
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
