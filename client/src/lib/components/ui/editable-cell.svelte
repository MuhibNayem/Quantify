<script lang="ts">
	import { Check, X, Pencil } from 'lucide-svelte';

	let {
		value = $bindable(),
		type = 'text',
		display,
		onSave,
		prefix = '',
		suffix = ''
	}: {
		value: string | number;
		type?: 'text' | 'number';
		display?: string;
		onSave: (newValue: string | number) => Promise<void> | void;
		prefix?: string;
		suffix?: string;
	} = $props();

	let editing = $state(false);
	let draft = $state<string | number>(value);
	let saving = $state(false);

	function focusSelect(node: HTMLInputElement) {
		node.focus();
		node.select();
	}

	function startEdit() {
		draft = value;
		editing = true;
	}

	async function commit() {
		if (saving) return;
		const next: string | number = type === 'number' ? Number(draft) : draft;
		const invalidNumber = type === 'number' && (Number.isNaN(next) || Number(next) < 0);
		if (next === value || invalidNumber) {
			editing = false;
			return;
		}
		saving = true;
		try {
			await onSave(next);
			value = next;
			editing = false;
		} catch {
			draft = value;
		} finally {
			saving = false;
		}
	}

	function cancel() {
		draft = value;
		editing = false;
	}

	function handleKeydown(e: KeyboardEvent) {
		e.stopPropagation();
		if (e.key === 'Enter') commit();
		if (e.key === 'Escape') cancel();
	}

	function stop(e: Event) {
		e.stopPropagation();
	}
</script>

{#if editing}
	<div
		class="inline-flex items-center gap-0.5 rounded-md border border-primary bg-card py-0.5 pr-0.5 pl-1.5 shadow-sm"
		role="group"
	>
		{#if prefix}
			<span class="text-xs font-semibold text-muted-foreground">{prefix}</span>
		{/if}
		<input
			{type}
			step={type === 'number' ? '0.01' : undefined}
			min={type === 'number' ? 0 : undefined}
			bind:value={draft}
			onkeydown={handleKeydown}
			onclick={stop}
			disabled={saving}
			use:focusSelect
			aria-label="Edit value"
			class="w-20 bg-transparent px-1 py-0.5 text-sm font-semibold text-foreground outline-none disabled:opacity-50"
		/>
		{#if suffix}
			<span class="text-xs font-semibold text-muted-foreground">{suffix}</span>
		{/if}
		<button
			onclick={(e) => {
				stop(e);
				commit();
			}}
			disabled={saving}
			aria-label="Save"
			class="inline-flex rounded-md p-1 transition-colors hover:bg-muted disabled:opacity-40"
		>
			<Check class="h-3.5 w-3.5 text-success" />
		</button>
		<button
			onclick={(e) => {
				stop(e);
				cancel();
			}}
			disabled={saving}
			aria-label="Cancel"
			class="inline-flex rounded-md p-1 transition-colors hover:bg-muted disabled:opacity-40"
		>
			<X class="h-3.5 w-3.5 text-danger" />
		</button>
	</div>
{:else}
	<button
		class="group/cell inline-flex items-center gap-1.5 rounded-md border border-transparent px-1.5 py-1 font-[inherit] transition-colors duration-200 hover:border-border hover:bg-muted hover:shadow-sm"
		onclick={(e) => {
			e.stopPropagation();
			startEdit();
		}}
		title="Click to edit"
	>
		<span class="text-sm font-semibold text-foreground">
			{#if prefix}
				<span class="mr-0.5 text-xs font-medium text-muted-foreground">{prefix}</span>
			{/if}
			{display ?? value}
			{#if suffix}
				<span class="ml-0.5 text-xs font-medium text-muted-foreground">{suffix}</span>
			{/if}
		</span>
		<Pencil
			class="h-3 w-3 text-primary opacity-0 transition-opacity duration-200 group-hover/cell:opacity-60"
		/>
	</button>
{/if}
