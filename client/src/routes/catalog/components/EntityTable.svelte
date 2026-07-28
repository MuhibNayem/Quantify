<script lang="ts" generics="T">
	import DataTable from '$lib/components/ui/data-table/DataTable.svelte';
	import type { Snippet } from 'svelte';

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
		children
	}: Props = $props();
</script>

<DataTable
	{data}
	{columns}
	{loading}
	{totalItems}
	{pageSize}
	{currentPage}
	{onPageChange}
	{onRowClick}
	{children}
/>
