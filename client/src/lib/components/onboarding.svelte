<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Sparkles } from 'lucide-svelte';

	const ONBOARDED_KEY = 'quantify:onboarded';

	let { open = $bindable(false) } = $props();
	let step = $state(0);

	const steps = [
		{
			title: 'Welcome to Quantify',
			body: "Your inventory, sales, and insights — all in one place. Let's take a 30-second tour."
		},
		{
			title: 'Sell with the POS',
			body: 'Scan barcodes or search products to build a cart, then check out with cash, card, or bKash.'
		},
		{
			title: 'Track everything',
			body: 'Stock levels, alerts, and returns update in real time. Watch the bell for notifications.'
		},
		{
			title: 'Understand your business',
			body: 'Reports and the Intelligence hub show sales trends, forecasts, and churn risk.'
		}
	];

	function next() {
		if (step < steps.length - 1) step++;
		else finish();
	}

	function finish() {
		localStorage.setItem(ONBOARDED_KEY, '1');
		open = false;
	}

	export function shouldShow() {
		return !localStorage.getItem(ONBOARDED_KEY);
	}
</script>

{#if open}
	<div
		class="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/30 p-4 backdrop-blur-md"
	>
		<div class="glass-float w-full max-w-md rounded-2xl p-8 text-center">
			<Sparkles class="mx-auto mb-4 h-7 w-7 text-primary" />
			<h2 class="text-xl font-semibold tracking-tight text-foreground">{steps[step].title}</h2>
			<p class="mx-auto mt-2 min-h-14 max-w-sm text-sm leading-relaxed text-muted-foreground">
				{steps[step].body}
			</p>
			<div class="my-5 flex justify-center gap-1.5">
				{#each steps as _, i}
					<span
						class="h-1.5 rounded-full transition-all duration-200 {i === step
							? 'w-5 bg-primary'
							: 'w-1.5 bg-border'}"
					></span>
				{/each}
			</div>
			<div class="flex items-center justify-between gap-2">
				<Button variant="ghost" size="sm" onclick={finish}>Skip</Button>
				<Button size="sm" onclick={next}>
					{step === steps.length - 1 ? 'Get Started' : 'Next'}
				</Button>
			</div>
		</div>
	</div>
{/if}
