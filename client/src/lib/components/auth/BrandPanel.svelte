<!-- client/src/lib/components/auth/BrandPanel.svelte
     Calm enterprise brand panel: porcelain wash, indigo/sage
     ambient, and a quiet live-ledger card. -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { Sparkles, Check, ShieldCheck, TrendingUp } from 'lucide-svelte';
	import { t } from '$lib/i18n';

	interface Props {
		title: string;
		subtitle: string;
	}
	let { title, subtitle }: Props = $props();

	const TARGET = 1284402.75;
	let displayed = $state(0);
	const fmt = new Intl.NumberFormat('en-US', {
		minimumFractionDigits: 2,
		maximumFractionDigits: 2
	});

	onMount(() => {
		const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
		if (reduce) {
			displayed = TARGET;
			return;
		}
		const duration = 1400;
		const start = performance.now();
		let raf = 0;
		const tick = (now: number) => {
			const p = Math.min(1, (now - start) / duration);
			const eased = 1 - Math.pow(1 - p, 3);
			displayed = TARGET * eased;
			if (p < 1) raf = requestAnimationFrame(tick);
		};
		raf = requestAnimationFrame(tick);
		return () => cancelAnimationFrame(raf);
	});

	const points = [22, 30, 26, 36, 32, 44, 40, 52, 47, 58, 54, 66, 61, 72];
	const W = 320;
	const H = 64;
	const sparkPath = $derived.by(() => {
		const max = Math.max(...points);
		const min = Math.min(...points);
		const step = W / (points.length - 1);
		return points
			.map((v, i) => {
				const x = i * step;
				const y = H - ((v - min) / (max - min)) * (H - 10) - 5;
				return `${i === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`;
			})
			.join(' ');
	});
	const sparkArea = $derived(`${sparkPath} L${W},${H} L0,${H} Z`);

	const rows = [
		{ account: 'Cash & Bank', code: '1010', debit: '412,806.20', credit: '—' },
		{ account: 'Accounts Receivable', code: '1120', debit: '186,240.00', credit: '—' },
		{ account: 'Inventory', code: '1140', debit: '342,918.55', credit: '—' },
		{ account: 'Accounts Payable', code: '2010', debit: '—', credit: '128,402.10' },
		{ account: 'Sales Revenue', code: '4010', debit: '—', credit: '512,886.40' }
	];
</script>

<aside
	class="bp-root relative hidden w-[48%] flex-col justify-between overflow-hidden border-r border-border lg:flex"
>
	<div class="hero-orbs" aria-hidden="true"></div>

	<!-- Brand row -->
	<div class="relative flex items-center justify-between px-12 pt-12 xl:px-16 xl:pt-14">
		<div class="flex items-center gap-3">
			<div
				class="flex h-10 w-10 items-center justify-center rounded-xl shadow-sm"
				style="background: var(--gradient-primary); box-shadow: 0 8px 24px -10px hsl(var(--primary) / 0.65);"
			>
				<Sparkles class="h-4 w-4 text-white" />
			</div>
			<div class="flex flex-col leading-none">
				<span class="font-display text-[20px] font-semibold tracking-[-0.02em] text-foreground">
					Quantify
				</span>
				<span
					class="mt-1.5 text-[0.6rem] font-semibold tracking-[0.18em] text-muted-foreground uppercase"
				>
					Enterprise Inventory
				</span>
			</div>
		</div>
		<span
			class="glass-pill px-3 py-1 font-mono text-[0.625rem] tracking-wider text-muted-foreground"
		>
			v10.0.0
		</span>
	</div>

	<!-- Headline + live ledger -->
	<div class="relative px-12 xl:px-16">
		<div class="flex items-center gap-2" data-animate="fade-up">
			<span class="live-dot"></span>
			<span class="text-[0.6875rem] font-semibold tracking-[0.16em] text-primary uppercase">
				Live ledger · Real-time
			</span>
		</div>

		<h1
			class="mt-6 max-w-md font-display text-[2.5rem] leading-[1.08] font-semibold tracking-[-0.03em] xl:text-[2.8rem]"
			data-animate="fade-up"
			style="animation-delay: 80ms"
		>
			<span class="text-gradient-hero">{title}</span>
		</h1>
		<p
			class="mt-5 max-w-md text-[1rem] leading-relaxed text-muted-foreground"
			data-animate="fade-up"
			style="animation-delay: 160ms"
		>
			{subtitle}
		</p>

		<!-- Live trial-balance card -->
		<div
			class="surface-2 mt-10 max-w-md p-6"
			aria-hidden="true"
			data-animate="fade-up"
			style="animation-delay: 240ms"
		>
			<div class="flex items-center justify-between border-b border-border/70 pb-3">
				<span
					class="text-[0.625rem] font-semibold tracking-[0.16em] text-muted-foreground uppercase"
				>
					Trial Balance
				</span>
				<span class="font-mono text-[0.625rem] tracking-wider text-muted-foreground">
					Q3 · FY2026
				</span>
			</div>

			<div class="pt-4">
				<div class="flex items-end justify-between gap-4">
					<div>
						<span
							class="text-[0.625rem] font-semibold tracking-[0.14em] text-muted-foreground uppercase"
						>
							Net position
						</span>
						<div class="kpi-num mt-1.5 text-[1.9rem] text-foreground">
							${fmt.format(displayed)}
						</div>
					</div>
					<span class="badge-success">
						<TrendingUp class="h-3 w-3" />
						+12.4%
					</span>
				</div>

				<svg viewBox="0 0 {W} {H}" class="mt-4 h-16 w-full" preserveAspectRatio="none">
					<defs>
						<linearGradient id="bp-spark-grad" x1="0" y1="0" x2="0" y2="1">
							<stop offset="0%" stop-color="hsl(var(--primary))" stop-opacity="0.28" />
							<stop offset="100%" stop-color="hsl(var(--primary))" stop-opacity="0" />
						</linearGradient>
					</defs>
					<path d={sparkArea} class="bp-spark-fill" fill="url(#bp-spark-grad)" />
					<path
						d={sparkPath}
						class="bp-spark-line"
						fill="none"
						stroke="hsl(var(--primary))"
						stroke-width="1.6"
						stroke-linecap="round"
						stroke-linejoin="round"
						vector-effect="non-scaling-stroke"
					/>
				</svg>
			</div>

			<div class="mt-4 space-y-0">
				<div class="flex items-center justify-between gap-4 border-b border-border/60 py-1.5">
					<span
						class="font-mono text-[0.5625rem] tracking-[0.14em] text-muted-foreground uppercase"
					>
						Account
					</span>
					<div class="flex gap-6">
						<span
							class="w-20 text-right font-mono text-[0.5625rem] tracking-[0.14em] text-muted-foreground uppercase"
						>
							Debit
						</span>
						<span
							class="w-20 text-right font-mono text-[0.5625rem] tracking-[0.14em] text-muted-foreground uppercase"
						>
							Credit
						</span>
					</div>
				</div>

				{#each rows as row, i}
					<div
						class="bp-row group flex items-center justify-between gap-4 rounded-lg px-1 py-2 hover:bg-muted/60"
						style="animation-delay: {420 + i * 70}ms"
					>
						<div class="flex min-w-0 items-baseline gap-2.5">
							<span
								class="truncate text-[0.8125rem] text-foreground/90 transition-colors group-hover:text-foreground"
							>
								{row.account}
							</span>
							<span class="font-mono text-[0.625rem] text-muted-foreground">
								{row.code}
							</span>
						</div>
						<div class="flex shrink-0 gap-6">
							<span
								class="w-20 text-right font-mono text-[0.75rem] text-foreground/80 tabular-nums"
							>
								{row.debit}
							</span>
							<span class="w-20 text-right font-mono text-[0.75rem] text-primary tabular-nums">
								{row.credit}
							</span>
						</div>
					</div>
				{/each}
			</div>

			<div class="mt-3 flex items-center justify-between border-t border-border/70 pt-3">
				<span class="inline-flex items-center gap-1.5 text-[0.6875rem] font-semibold text-success">
					<Check class="h-3.5 w-3.5" />
					Reconciled · just now
				</span>
				<span class="font-mono text-[0.625rem] text-muted-foreground">42 accounts</span>
			</div>
		</div>
	</div>

	<!-- Trust footer -->
	<div
		class="relative flex items-center justify-between gap-4 px-12 pb-12 xl:px-16 xl:pb-14"
		data-animate="fade-up"
		style="animation-delay: 340ms"
	>
		<p class="flex items-center gap-2 text-xs text-muted-foreground">
			<ShieldCheck class="h-3.5 w-3.5 shrink-0 text-primary" />
			{$t('auth.login.securedBy')}
		</p>
		<div class="hidden items-center gap-1.5 xl:flex">
			{#each ['SOC 2', 'AES-256', '99.99%'] as chip}
				<span
					class="glass-pill px-2.5 py-0.5 font-mono text-[0.5625rem] tracking-wider text-muted-foreground"
				>
					{chip}
				</span>
			{/each}
		</div>
	</div>
</aside>

<style>
	.bp-root {
		background:
			radial-gradient(900px 520px at 0% 0%, hsl(224 85% 94% / 0.9), transparent 62%),
			radial-gradient(760px 520px at 100% 100%, hsl(168 55% 92% / 0.75), transparent 60%),
			hsl(var(--card));
	}
	:root.dark .bp-root {
		background:
			radial-gradient(900px 520px at 0% 0%, hsl(224 70% 24% / 0.55), transparent 62%),
			radial-gradient(760px 520px at 100% 100%, hsl(168 45% 20% / 0.45), transparent 60%),
			hsl(var(--card));
	}

	.bp-spark-line {
		stroke-dasharray: 520;
		stroke-dashoffset: 520;
		animation: bp-draw 1.6s cubic-bezier(0.3, 0, 0.2, 1) forwards 0.4s;
	}
	.bp-spark-fill {
		opacity: 0;
		animation: bp-fade 1s ease-out forwards 0.9s;
	}

	.bp-row {
		opacity: 0;
		animation: bp-rise 0.5s cubic-bezier(0.2, 0, 0, 1) forwards;
		transition: background 0.15s ease;
	}

	@keyframes bp-draw {
		to {
			stroke-dashoffset: 0;
		}
	}
	@keyframes bp-fade {
		to {
			opacity: 1;
		}
	}
	@keyframes bp-rise {
		from {
			opacity: 0;
			transform: translateY(6px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.bp-spark-line,
		.bp-spark-fill,
		.bp-row {
			animation: none;
			opacity: 1;
			stroke-dashoffset: 0;
		}
	}
</style>
