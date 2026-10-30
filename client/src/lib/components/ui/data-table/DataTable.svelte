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
	import { Checkbox } from '$lib/components/ui/checkbox';
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
		selectable?: boolean;
		allSelected?: boolean;
		someSelected?: boolean;
		onSelectAll?: () => void;
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
		children,
		selectable = false,
		allSelected = false,
		someSelected = false,
		onSelectAll
	}: Props = $props();

	let totalPages = $derived(Math.ceil(totalItems / pageSize));

	const handlePageChange = (page: number) => {
		if (onPageChange && page >= 1 && page <= totalPages) {
			onPageChange(page);
		}
	};
</script>

<div class="w-full space-y-4">
	<!-- Main Table Container -->
	<div class="panel overflow-hidden" in:fly={{ y: 16, duration: 500, delay: 150 }}>
		<Table class="glass-table">
			<TableHeader class="border-b border-border bg-muted/60">
				<TableRow class="border-border hover:bg-transparent">
					{#if selectable}
						<TableHead class="w-12 px-4">
							<Checkbox
								checked={allSelected}
								indeterminate={someSelected && !allSelected}
								onCheckedChange={() => onSelectAll?.()}
							/>
						</TableHead>
					{/if}
					{#each columns as col}
						<TableHead
							class="h-11 px-4 text-[0.6875rem] font-semibold tracking-[0.08em] {col.class ?? ''}"
						>
							<div
								class="flex items-center gap-2 {col.class?.includes('text-right')
									? 'justify-end'
									: 'justify-start'}"
							>
								{#if col.sortable}
									<Button
										variant="ghost"
										size="sm"
										class="-ml-2 h-8 px-2 text-[0.6875rem] font-semibold tracking-[0.08em] text-muted-foreground transition-colors hover:bg-transparent hover:text-foreground"
									>
										{col.header}
										<ArrowUpDown class="ml-1.5 h-3.5 w-3.5 opacity-50" />
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
						<TableRow class="border-b border-border/60 hover:bg-transparent">
							{#if selectable}
								<TableCell class="px-4 py-3.5"></TableCell>
							{/if}
							{#each columns as _}
								<TableCell class="px-4 py-3.5">
									<div
										class="h-6 w-full animate-pulse rounded-md bg-muted"
										style="animation-delay: {i * 100}ms"
									></div>
								</TableCell>
							{/each}
						</TableRow>
					{/each}
				{:else if data.length === 0}
					<TableRow>
						<TableCell
							colspan={columns.length + (selectable ? 1 : 0)}
							class="h-[360px] text-center"
						>
							<div class="flex flex-col items-center justify-center gap-3 text-muted-foreground">
								<div class="rounded-full bg-muted p-5">
									<FileQuestion class="h-7 w-7 text-muted-foreground/70" />
								</div>
								<div class="space-y-1">
									<p class="text-sm font-semibold text-foreground">No records found</p>
									<p class="text-sm text-muted-foreground">Try adjusting your filters</p>
								</div>
							</div>
						</TableCell>
					</TableRow>
				{:else}
					{#each data as row, i (i)}
						<TableRow
							class="group border-b border-border/60 transition-colors hover:bg-primary/[0.035] {onRowClick
								? 'cursor-pointer'
								: ''}"
							onclick={() => onRowClick?.(row)}
						>
							{#if children}
								{@render children(row)}
							{:else}
								{#each columns as col}
									<TableCell
										class="px-4 py-3.5 text-sm font-medium text-foreground {col.class ?? ''}"
									>
										{#if col.accessorKey}
											{@const value = row[col.accessorKey as keyof T]}
											<span in:fade={{ duration: 350, delay: i * 40 }}>
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
		<div class="flex items-center justify-end px-1" in:fade={{ duration: 300, delay: 250 }}>
			<Pagination.Root
				count={totalItems}
				perPage={pageSize}
				page={currentPage}
				onPageChange={(page) => handlePageChange(page)}
			>
				{#snippet children({ pages, currentPage })}
					<Pagination.Content class="gap-1.5">
						<Pagination.Item>
							<Pagination.PrevButton
								class="h-9 w-9 rounded-md border border-border bg-card text-muted-foreground shadow-xs transition-colors hover:bg-muted hover:text-foreground disabled:opacity-30"
							/>
						</Pagination.Item>
						{#each pages as page (page.key)}
							{#if page.type === 'ellipsis'}
								<Pagination.Item>
									<Pagination.Ellipsis class="text-muted-foreground" />
								</Pagination.Item>
							{:else}
								<Pagination.Item>
									<Pagination.Link
										{page}
										isActive={currentPage === page.value}
										class="h-9 w-9 rounded-md border text-sm font-semibold transition-colors {currentPage ===
										page.value
											? 'border-primary bg-primary text-primary-foreground shadow-xs'
											: 'border-border bg-card text-muted-foreground hover:bg-muted hover:text-foreground'}"
									>
										{page.value}
									</Pagination.Link>
								</Pagination.Item>
							{/if}
						{/each}
						<Pagination.Item>
							<Pagination.NextButton
								class="h-9 w-9 rounded-md border border-border bg-card text-muted-foreground shadow-xs transition-colors hover:bg-muted hover:text-foreground disabled:opacity-30"
							/>
						</Pagination.Item>
					</Pagination.Content>
				{/snippet}
			</Pagination.Root>
		</div>
	{/if}
</div>
