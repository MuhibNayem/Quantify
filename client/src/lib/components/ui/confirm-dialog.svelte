<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as Dialog from '$lib/components/ui/dialog';
	import { AlertTriangle } from 'lucide-svelte';

	let {
		open = $bindable(false),
		title,
		description,
		confirmText = '',
		confirmLabel = 'Delete',
		onConfirm
	}: {
		open?: boolean;
		title: string;
		description: string;
		confirmText?: string;
		confirmLabel?: string;
		onConfirm?: () => void;
	} = $props();

	let typed = $state('');
	const canConfirm = $derived(!confirmText || typed === confirmText);

	function handleConfirm() {
		if (!canConfirm) return;
		onConfirm?.();
		open = false;
		typed = '';
	}

	function handleOpenChange(v: boolean) {
		open = v;
		if (!v) typed = '';
	}
</script>

<Dialog.Root {open} onOpenChange={handleOpenChange}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<div class="flex items-start gap-4">
				<div
					class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-rose-100"
				>
					<AlertTriangle class="h-5 w-5 text-rose-600" />
				</div>
				<div class="space-y-1">
					<Dialog.Title class="text-lg font-semibold text-slate-900">{title}</Dialog.Title>
					<Dialog.Description class="text-sm text-slate-500">
						{description}
					</Dialog.Description>
				</div>
			</div>
		</Dialog.Header>

		{#if confirmText}
			<div class="space-y-2 px-6">
				<p class="text-sm text-slate-600">
					Type <span class="font-mono font-semibold text-slate-900">{confirmText}</span> to confirm
				</p>
				<Input
					bind:value={typed}
					placeholder={confirmText}
					class="font-mono"
					onkeydown={(e) => e.key === 'Enter' && handleConfirm()}
				/>
			</div>
		{/if}

		<Dialog.Footer class="px-6 pb-6">
			<Button variant="outline" onclick={() => handleOpenChange(false)}>Cancel</Button>
			<Button variant="destructive" disabled={!canConfirm} onclick={handleConfirm}>
				{confirmLabel}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
