<script lang="ts">
	import { Check, ChevronsUpDown, Loader2, Search } from 'lucide-svelte';
	import { cn } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import {
		Command,
		CommandEmpty,
		CommandGroup,
		CommandInput,
		CommandItem,
		CommandList
	} from '$lib/components/ui/command';
	import { Popover, PopoverContent, PopoverTrigger } from '$lib/components/ui/popover';
	import { productsApi } from '$lib/api/resources';
	import type { Product } from '$lib/types';
	import { onMount } from 'svelte';

	let {
		value = $bindable(''),
		placeholder = 'Select product...',
		className = '',
		onSelect
	} = $props<{
		value?: string;
		placeholder?: string;
		className?: string;
		onSelect?: (product: Product) => void;
	}>();

	let open = $state(false);
	let products = $state<Product[]>([]);
	let loading = $state(false);
	let selectedProduct = $state<Product | null>(null);
	let searchQuery = $state('');
	let debounceTimer: NodeJS.Timeout;

	async function fetchProducts(search: string = '') {
		loading = true;
		try {
			const res = await productsApi.list(1, 50, search);
			products = res.products || [];
		} catch (e) {
			console.error('Failed to fetch products', e);
			products = [];
		} finally {
			loading = false;
		}
	}

	function handleSearch(val: string) {
		searchQuery = val;
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => {
			fetchProducts(val);
		}, 300);
	}

	// Initial fetch
	onMount(() => {
		fetchProducts();
	});

	// Update selected label when value changes externally
	$effect(() => {
		if (value && !selectedProduct) {
			// If we have a value but no object, we might need to fetch it or find it
			// For now, we rely on the list. If not in list, label might be missing.
			// Ideally we fetch the specific product if missing.
			const found = products.find((p) => String(p.ID) === String(value));
			if (found) selectedProduct = found;
			else if (value) {
				// Fetch single product to get label
				productsApi
					.get(Number(value))
					.then((p) => (selectedProduct = p))
					.catch(() => {});
			}
		}
	});
</script>

<Popover bind:open>
	<PopoverTrigger>
		{#snippet child({ props })}
			<Button
				variant="outline"
				role="combobox"
				aria-expanded={open}
				class={cn(
					'h-12 w-full justify-between rounded-2xl border-white/60 bg-white/50 shadow-sm backdrop-blur-md transition-all hover:bg-white/80',
					className
				)}
				{...props}
			>
				{#if selectedProduct}
					<span class="truncate">{selectedProduct.Name}</span>
				{:else}
					<span class="text-muted-foreground">{placeholder}</span>
				{/if}
				<ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
			</Button>
		{/snippet}
	</PopoverTrigger>
	<PopoverContent
		class="z-50 w-[300px] rounded-2xl border border-white/60 bg-white/90 p-0 shadow-[0_40px_100px_-20px_rgba(50,60,90,0.2),inset_0_1px_0_0_rgba(255,255,255,0.8)] outline-none backdrop-blur-3xl backdrop-saturate-[180%]"
		align="start"
	>
		<Command shouldFilter={false}>
			<div class="flex items-center border-b border-white/20 px-4 py-2" data-cmdk-input-wrapper="">
				<Search class="mr-2 h-4 w-4 shrink-0 opacity-50" />
				<input
					class="flex h-12 w-full rounded-md bg-transparent py-3 text-base font-medium text-slate-900 outline-none placeholder:text-slate-500/80 disabled:cursor-not-allowed disabled:opacity-50"
					placeholder="Search by name or SKU..."
					value={searchQuery}
					oninput={(e) => handleSearch(e.currentTarget.value)}
				/>
			</div>
			<CommandList>
				{#if loading}
					<div class="text-muted-foreground flex justify-center py-6 text-center text-sm">
						<Loader2 class="mr-2 h-4 w-4 animate-spin" /> Loading...
					</div>
				{:else if products.length === 0}
					<CommandEmpty>No product found.</CommandEmpty>
				{:else}
					<CommandGroup>
						{#each products as product}
							<CommandItem
								value={String(product.ID)}
								onSelect={() => {
									value = String(product.ID);
									selectedProduct = product;
									open = false;
									onSelect?.(product);
								}}
							>
								<Check
									class={cn(
										'mr-2 h-4 w-4',
										value === String(product.ID) ? 'opacity-100' : 'opacity-0'
									)}
								/>
								<div class="flex flex-col">
									<span class="font-medium text-slate-900">{product.Name}</span>
									<span class="text-xs text-slate-500">SKU: {product.SKU}</span>
								</div>
							</CommandItem>
						{/each}
					</CommandGroup>
				{/if}
			</CommandList>
		</Command>
	</PopoverContent>
</Popover>
