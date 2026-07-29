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
	<div class="onboarding-overlay">
		<div class="onboarding-card">
			<Sparkles size={28} />
			<h2>{steps[step].title}</h2>
			<p>{steps[step].body}</p>
			<div class="dots">
				{#each steps as _, i}
					<span class="dot" class:active={i === step}></span>
				{/each}
			</div>
			<div class="actions">
				<Button variant="ghost" size="sm" onclick={finish}>Skip</Button>
				<Button size="sm" onclick={next}
					>{step === steps.length - 1 ? 'Get Started' : 'Next'}</Button
				>
			</div>
		</div>
	</div>
{/if}

<style>
	.onboarding-overlay {
		position: fixed;
		inset: 0;
		z-index: 100;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(15, 23, 42, 0.5);
		backdrop-filter: blur(6px);
	}
	.onboarding-card {
		width: 100%;
		max-width: 26rem;
		margin: 1rem;
		padding: 2rem;
		border-radius: 1.5rem;
		border: 1px solid rgba(255, 255, 255, 0.6);
		background: rgba(255, 255, 255, 0.95);
		box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
		text-align: center;
		color: #1e293b;
	}
	.onboarding-card :global(svg) {
		margin: 0 auto 1rem;
		color: #6366f1;
	}
	h2 {
		font-size: 1.25rem;
		font-weight: 700;
		margin-bottom: 0.5rem;
	}
	p {
		font-size: 0.9rem;
		color: #64748b;
		line-height: 1.5;
		min-height: 3.5rem;
	}
	.dots {
		display: flex;
		justify-content: center;
		gap: 0.4rem;
		margin: 1.25rem 0;
	}
	.dot {
		width: 0.5rem;
		height: 0.5rem;
		border-radius: 9999px;
		background: #cbd5e1;
		transition: background 0.2s;
	}
	.dot.active {
		background: #6366f1;
	}
	.actions {
		display: flex;
		justify-content: space-between;
		gap: 0.5rem;
	}
</style>
