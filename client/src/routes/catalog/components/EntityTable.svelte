<script lang="ts" generics="T">
	import * as Pagination from '$lib/components/ui/pagination';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import type { Snippet } from 'svelte';
	import { fade } from 'svelte/transition';

	type Column = {
		accessorKey?: keyof T | string;
		header: string;
		sortable?: boolean;
		class?: string;
	};

	type Props = {
		data: T[];
		columns: Column[];
		loading?: boolean;
		totalItems?: number;
		pageSize?: number;
		currentPage?: number;
		onPageChange?: (page: number) => void;
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
		loading = false,
		totalItems = 0,
		pageSize = 10,
		currentPage = 1,
		onPageChange,
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
	<div class="panel overflow-hidden" in:fade={{ duration: 250 }}>
		<div class="custom-scrollbar overflow-x-auto">
			<table class="glass-table min-w-[36rem]">
				<thead>
					<tr>
						{#if selectable}
							<th class="w-12">
								<Checkbox
									checked={allSelected}
									indeterminate={someSelected && !allSelected}
									onCheckedChange={() => onSelectAll?.()}
								/>
							</th>
						{/if}
						{#each columns as col}
							<th class={col.class ?? ''}>{col.header}</th>
						{/each}
					</tr>
				</thead>
				<tbody>
					{#if loading}
						{#each Array(pageSize) as _, i}
							<tr>
								{#if selectable}
									<td></td>
								{/if}
								{#each columns as _}
									<td>
										<div
											class="h-3.5 w-full animate-pulse rounded-md bg-muted"
											style="animation-delay: {i * 80}ms"
										></div>
									</td>
								{/each}
							</tr>
						{/each}
					{:else if data.length === 0}
						<tr>
							<td
								colspan={columns.length + (selectable ? 1 : 0)}
								class="px-4 py-14 text-center text-sm text-muted-foreground"
							>
								No records found
							</td>
						</tr>
					{:else}
						{#each data as row, i (i)}
							<tr class={onRowClick ? 'cursor-pointer' : ''} onclick={() => onRowClick?.(row)}>
								{#if children}
									{@render children(row)}
								{:else}
									{#each columns as col}
										<td class={col.class ?? ''}>
											{#if col.accessorKey}
												{@const value = row[col.accessorKey as keyof T]}
												<span in:fade={{ duration: 300, delay: i * 30 }}>{value}</span>
											{/if}
										</td>
									{/each}
								{/if}
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>

	{#if totalPages > 1}
		<div class="flex items-center justify-end px-1" in:fade={{ duration: 300, delay: 150 }}>
			<Pagination.Root
				count={totalItems}
				perPage={pageSize}
				page={currentPage}
				onPageChange={(page) => handlePageChange(page)}
			>
				{#snippet children({ pages, currentPage })}
					<Pagination.Content class="gap-1">
						<Pagination.Item>
							<Pagination.PrevButton
								class="h-9 w-9 rounded-lg text-muted-foreground shadow-none hover:bg-muted hover:text-foreground disabled:opacity-40"
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
										class="num h-9 w-9 rounded-lg border text-sm font-semibold transition-colors {currentPage ===
										page.value
											? 'border-primary/30 bg-accent text-accent-foreground'
											: 'border-border bg-card text-muted-foreground hover:bg-muted hover:text-foreground'}"
									>
										{page.value}
									</Pagination.Link>
								</Pagination.Item>
							{/if}
						{/each}
						<Pagination.Item>
							<Pagination.NextButton
								class="h-9 w-9 rounded-lg text-muted-foreground shadow-none hover:bg-muted hover:text-foreground disabled:opacity-40"
							/>
						</Pagination.Item>
					</Pagination.Content>
				{/snippet}
			</Pagination.Root>
		</div>
	{/if}
</div>
