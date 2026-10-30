<script lang="ts">
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { cashDrawerApi } from '$lib/api/resources';
	import { formatCurrency } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { ArrowDownToLine, LockKeyhole, Unlock } from 'lucide-svelte';

	let session = $state<any>(null);
	let startingCash = $state(0);
	let endingCash = $state(0);
	let dropAmount = $state(0);
	let busy = $state(false);

	async function load() {
		try {
			const data = await cashDrawerApi.current();
			session = data.session;
			if (session) endingCash = session.SystemCash ?? 0;
		} catch {
			session = null;
		}
	}

	async function openDrawer() {
		busy = true;
		try {
			session = await cashDrawerApi.open(startingCash);
			endingCash = session.SystemCash ?? startingCash;
			toast.success('Cash drawer opened', {
				description: `Starting float ${formatCurrency(startingCash)}`
			});
		} catch (error: any) {
			toast.error('Could not open drawer', {
				description: error?.response?.data?.error || error?.message
			});
		} finally {
			busy = false;
		}
	}

	async function dropCash() {
		if (dropAmount <= 0) return;
		busy = true;
		try {
			session = await cashDrawerApi.drop(dropAmount);
			dropAmount = 0;
			toast.success('Safe drop recorded');
		} catch (error: any) {
			toast.error('Drop failed', { description: error?.response?.data?.error });
		} finally {
			busy = false;
		}
	}

	async function closeDrawer() {
		busy = true;
		try {
			session = await cashDrawerApi.close(endingCash);
			toast.success('Cash drawer closed', {
				description: `Variance ${formatCurrency(session.Variance)}`
			});
			session = null;
		} catch (error: any) {
			toast.error('Could not close drawer', { description: error?.response?.data?.error });
		} finally {
			busy = false;
		}
	}

	onMount(load);
</script>

{#if session}
	<div class="panel flex flex-wrap items-center gap-x-4 gap-y-2 px-4 py-2.5">
		<span class="live-dot" aria-hidden="true"></span>
		<span class="text-xs font-semibold text-foreground">Drawer open</span>
		<span class="text-xs text-muted-foreground">
			Expected
			<span class="num-mono ml-1 font-semibold text-foreground">
				{formatCurrency(session.SystemCash ?? 0)}
			</span>
		</span>
		<span class="hidden text-xs text-muted-foreground sm:inline">
			Sales <span class="num-mono text-foreground">{formatCurrency(session.TotalSales ?? 0)}</span>
			· Refunds
			<span class="num-mono text-foreground">{formatCurrency(session.TotalRefunds ?? 0)}</span>
			· Drops
			<span class="num-mono text-foreground">{formatCurrency(session.TotalDrops ?? 0)}</span>
		</span>

		<div class="ml-auto flex flex-wrap items-center gap-2">
			<Input
				type="number"
				min="0"
				step="0.01"
				bind:value={dropAmount}
				class="glass-input h-8 w-24 text-xs"
				placeholder="Drop"
				aria-label="Cash drop amount"
			/>
			<Button variant="outline" size="sm" class="h-8" onclick={dropCash} disabled={busy}>
				<ArrowDownToLine class="mr-1.5 h-3.5 w-3.5" />
				Drop
			</Button>
			<Input
				type="number"
				min="0"
				step="0.01"
				bind:value={endingCash}
				class="glass-input h-8 w-28 text-xs"
				placeholder="Counted"
				aria-label="Counted ending cash"
			/>
			<Button size="sm" class="h-8" onclick={closeDrawer} disabled={busy}>
				<LockKeyhole class="mr-1.5 h-3.5 w-3.5" />
				Close
			</Button>
		</div>
	</div>
{:else}
	<div class="panel flex flex-wrap items-center gap-3 px-4 py-2.5">
		<Unlock class="h-4 w-4 text-muted-foreground" />
		<span class="text-xs font-semibold text-foreground">Cash drawer closed</span>
		<span class="text-xs text-muted-foreground">
			Open a shift to reconcile cash sales and refunds.
		</span>
		<div class="ml-auto flex items-center gap-2">
			<Input
				type="number"
				min="0"
				step="0.01"
				bind:value={startingCash}
				class="glass-input h-8 w-28 text-xs"
				placeholder="Starting float"
				aria-label="Starting cash float"
			/>
			<Button size="sm" class="h-8" onclick={openDrawer} disabled={busy}>Open drawer</Button>
		</div>
	</div>
{/if}
