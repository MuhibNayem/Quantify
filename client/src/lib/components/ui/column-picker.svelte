<script lang="ts">
	import { Settings2 } from 'lucide-svelte';
	import { Popover, PopoverContent, PopoverTrigger } from '$lib/components/ui/popover';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import type { ColumnDef } from '$lib/column-prefs';

	let {
		columns,
		visible = $bindable(new Set<string>()),
		onToggle
	}: {
		columns: ColumnDef[];
		visible: Set<string>;
		onToggle: (key: string) => void;
	} = $props();
</script>

<Popover>
	<PopoverTrigger>
		{#snippet child({ props })}
			<button {...props} class="col-picker" aria-label="Customize columns">
				<Settings2 size={16} />
			</button>
		{/snippet}
	</PopoverTrigger>
	<PopoverContent align="end" class="w-48">
		<p class="mb-2 text-xs font-semibold opacity-60">Columns</p>
		{#each columns as col}
			<label class="flex cursor-pointer items-center gap-2 py-1 text-sm">
				<Checkbox checked={visible.has(col.key)} onCheckedChange={() => onToggle(col.key)} />
				<span>{col.label}</span>
			</label>
		{/each}
	</PopoverContent>
</Popover>

<style>
	.col-picker {
		display: inline-flex;
		padding: 0.35rem;
		border-radius: 6px;
		opacity: 0.6;
		transition: opacity 0.15s;
	}
	.col-picker:hover {
		opacity: 1;
		background: rgba(0, 0, 0, 0.05);
	}
</style>
