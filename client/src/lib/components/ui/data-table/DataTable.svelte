<script lang="ts" generics="T">
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import * as Pagination from '$lib/components/ui/pagination';
	import { ArrowUpDown, Loader2, FileQuestion } from 'lucide-svelte';
	import type { Snippet } from 'svelte';
	import { fade, fly } from 'svelte/transition';

	type Props = {
		data: T[];
		columns: {
			accessorKey?: keyof T | string;
			header: string;
			sortable?: boolean;
			class?: string;
		}[];
		totalItems?: number;
		pageSize?: number;
		currentPage?: number;
		onPageChange?: (page: number) => void;
		loading?: boolean;
		onRowClick?: (row: T) => void;
		children?: Snippet<[T]>;
	};

	let {
		data,
		columns,
		totalItems = 0,
		pageSize = 10,
		currentPage = 1,
		onPageChange,
		loading = false,
		onRowClick,
		children
	} = $props<Props>();

	let totalPages = $derived(Math.ceil(totalItems / pageSize));

	const handlePageChange = (page: number) => {
		if (onPageChange && page >= 1 && page <= totalPages) {
			onPageChange(page);
		}
	};
</script>

<div class="w-full space-y-6">
	<!-- Main Table Container -->
	<div
		class="relative overflow-hidden rounded-3xl border border-white/40 bg-white/60 shadow-2xl shadow-indigo-100/50 backdrop-blur-xl transition-all duration-500 hover:shadow-indigo-200/50"
		in:fly={{ y: 20, duration: 600, delay: 200 }}
	>
		<!-- Decorative Gradient Blob -->
		<div
			class="pointer-events-none absolute -right-20 -top-20 h-64 w-64 rounded-full bg-gradient-to-br from-indigo-100/40 to-purple-100/40 blur-3xl transition-transform duration-1000"
		></div>

		<Table>
			<TableHeader class="bg-gradient-to-r from-indigo-50/50 to-white/50">
				<TableRow class="border-b border-indigo-100/50 hover:bg-transparent">
					{#each columns as col}
						<TableHead
							class="h-14 px-6 text-xs font-bold uppercase tracking-widest text-indigo-900/60 {col.class ??
								''}"
						>
							<div class="flex items-center gap-2">
								{#if col.sortable}
									<Button
										variant="ghost"
										size="sm"
										class="-ml-3 h-8 text-xs font-bold uppercase tracking-widest text-indigo-900/60 transition-colors hover:bg-indigo-100/50 hover:text-indigo-700"
									>
										{col.header}
										<ArrowUpDown class="ml-2 h-3.5 w-3.5 opacity-50" />
									</Button>
								{:else}
									{col.header}
								{/if}
							</div>
						</TableHead>
					{/each}
				</TableRow>
			</TableHeader>
			<TableBody>
				{#if loading}
					{#each Array(pageSize) as _, i}
						<TableRow class="border-b border-indigo-50/30 hover:bg-transparent">
							{#each columns as _}
								<TableCell class="px-6 py-4">
									<div
										class="h-8 w-full animate-pulse rounded-full bg-indigo-50/50"
										style="animation-delay: {i * 100}ms"
									></div>
								</TableCell>
							{/each}
						</TableRow>
					{/each}
				{:else if data.length === 0}
					<TableRow>
						<TableCell colspan={columns.length} class="h-[400px] text-center">
							<div class="flex flex-col items-center justify-center gap-4 text-indigo-300">
								<div class="rounded-full bg-indigo-50 p-6 shadow-inner">
									<FileQuestion class="h-10 w-10 text-indigo-400" />
								</div>
								<div class="space-y-1">
									<p class="text-lg font-semibold text-indigo-900/70">No records found</p>
									<p class="text-sm text-indigo-400">Try adjusting your filters</p>
								</div>
							</div>
						</TableCell>
					</TableRow>
				{:else}
					{#each data as row, i (i)}
						<TableRow
							class="group border-b border-indigo-50/30 transition-all duration-300 hover:bg-indigo-50/40 {onRowClick
								? 'cursor-pointer active:scale-[0.995] active:bg-indigo-100/50'
								: ''}"
							onclick={() => onRowClick?.(row)}
						>
							{#if children}
								{@render children(row)}
							{:else}
								{#each columns as col}
									<TableCell
										class="px-6 py-4 text-sm font-medium text-slate-600 transition-colors group-hover:text-slate-900 {col.class ??
											''}"
									>
										{#if col.accessorKey}
											{@const value = row[col.accessorKey as keyof T]}
											<span in:fade={{ duration: 400, delay: i * 50 }}>
												{value}
											</span>
										{/if}
									</TableCell>
								{/each}
							{/if}
						</TableRow>
					{/each}
				{/if}
			</TableBody>
		</Table>
	</div>

	{#if totalPages > 1}
		<div class="flex items-center justify-end px-2" in:fade={{ duration: 400, delay: 400 }}>
			<Pagination.Root
				count={totalItems}
				perPage={pageSize}
				page={currentPage}
				onPageChange={(page) => handlePageChange(page)}
			>
				{#snippet children({ pages, currentPage })}
					<Pagination.Content class="gap-2">
						<Pagination.Item>
							<Pagination.PrevButton
								class="h-9 w-9 rounded-xl border border-white/60 bg-white/50 shadow-sm backdrop-blur transition-all hover:border-indigo-200 hover:bg-white hover:text-indigo-600 hover:shadow-md disabled:opacity-30"
							/>
						</Pagination.Item>
						{#each pages as page (page.key)}
							{#if page.type === 'ellipsis'}
								<Pagination.Item>
									<Pagination.Ellipsis class="text-indigo-300" />
								</Pagination.Item>
							{:else}
								<Pagination.Item>
									<Pagination.Link
										{page}
										isActive={currentPage === page.value}
										class="h-9 w-9 rounded-xl border text-sm font-medium transition-all duration-300 {currentPage ===
										page.value
											? 'scale-105 border-indigo-500 bg-indigo-500 text-white shadow-lg shadow-indigo-500/30'
											: 'border-white/60 bg-white/50 text-slate-600 hover:border-indigo-200 hover:bg-white hover:text-indigo-600 hover:shadow-md'}"
									>
										{page.value}
									</Pagination.Link>
								</Pagination.Item>
							{/if}
						{/each}
						<Pagination.Item>
							<Pagination.NextButton
								class="h-9 w-9 rounded-xl border border-white/60 bg-white/50 shadow-sm backdrop-blur transition-all hover:border-indigo-200 hover:bg-white hover:text-indigo-600 hover:shadow-md disabled:opacity-30"
							/>
						</Pagination.Item>
					</Pagination.Content>
				{/snippet}
			</Pagination.Root>
		</div>
	{/if}
</div>
