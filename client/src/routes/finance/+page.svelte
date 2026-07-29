<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { financeApi } from '$lib/api/resources';
	import { auth } from '$lib/stores/auth';
	import { cn, formatCurrency, formatDate, formatPercent } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import EmptyState from '$lib/components/ui/empty-state.svelte';
	import GlassTabs from '$lib/components/ui/glass-tabs.svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import { toast } from 'svelte-sonner';
	import { fly } from 'svelte/transition';
	import {
		Landmark,
		Scale,
		BookOpen,
		ReceiptText,
		Plus,
		Loader2,
		Search,
		AlertTriangle,
		CheckCircle2,
		HandCoins,
		Banknote,
		CreditCard,
		Wallet
	} from 'lucide-svelte';
	import type {
		AccountsReceivable,
		ARAgingBucket,
		TrialBalanceRow,
		IncomeStatement,
		BalanceSheet
	} from '$lib/types';

	const inputClass =
		'liquid-input h-10 w-full rounded-md bg-card bg-none px-3 text-sm font-normal text-foreground shadow-none ring-0 backdrop-blur-none backdrop-saturate-100 from-card to-card placeholder:text-muted-foreground focus:scale-100 focus:border-primary focus:ring-0 md:text-sm';
	const btnSecondary =
		'rounded-md border border-border bg-card bg-none text-foreground shadow-none backdrop-blur-none hover:bg-muted hover:bg-none';

	type Tone = 'success' | 'info' | 'warning' | 'danger';
	const toneText: Record<Tone, string> = {
		success: 'text-success',
		info: 'text-info',
		warning: 'text-warning',
		danger: 'text-danger'
	};
	const toneBorder: Record<Tone, string> = {
		success: 'border-l-success',
		info: 'border-l-info',
		warning: 'border-l-warning',
		danger: 'border-l-danger'
	};
	const statusBadge: Record<string, string> = {
		UNPAID: 'badge-danger',
		PARTIAL: 'badge-warning',
		PAID: 'badge-success'
	};
	const methods = [
		{ value: 'cash', label: 'Cash', icon: Banknote },
		{ value: 'card', label: 'Card', icon: CreditCard },
		{ value: 'bkash', label: 'bKash', icon: Wallet },
		{ value: 'bank', label: 'Bank', icon: Landmark }
	];

	let activeTab = $state('receivables');

	let arLoading = $state(true);
	let buckets = $state<ARAgingBucket[]>([]);
	let receivables = $state<AccountsReceivable[]>([]);
	let statusFilter = $state('');

	let stmtTab = $state('income');
	let from = $state('');
	let to = $state('');
	let stmtLoading = $state(false);
	let income = $state<IncomeStatement | null>(null);
	let sheet = $state<BalanceSheet | null>(null);
	let trialRows = $state<TrialBalanceRow[] | null>(null);

	let payOpen = $state(false);
	let paying = $state(false);
	let payTarget = $state<AccountsReceivable | null>(null);
	let payCustomerId = $state('');
	let payAmount = $state('');
	let payMethod = $state('cash');
	let payNotes = $state('');

	function num(v: string | number | null | undefined): number {
		const n = typeof v === 'number' ? v : parseFloat(v ?? '0');
		return Number.isFinite(n) ? n : 0;
	}

	function money(v: string | number | null | undefined): string {
		return formatCurrency(num(v));
	}

	function outstandingOf(ar: AccountsReceivable): number {
		return num(ar.Amount) - num(ar.PaidAmount);
	}

	function isOverdue(ar: AccountsReceivable): boolean {
		return !!ar.DueDate && ar.Status !== 'PAID' && new Date(ar.DueDate).getTime() < Date.now();
	}

	function customerName(ar: AccountsReceivable): string {
		return ar.Customer?.Username || `Customer #${ar.CustomerID}`;
	}

	function methodLabel(v: string): string {
		return methods.find((m) => m.value === v)?.label ?? v;
	}

	function bucketTone(label: string, i: number): Tone {
		const l = label.toLowerCase();
		if (l.includes('current')) return 'success';
		if (l.startsWith('90')) return 'danger';
		if (l.startsWith('61')) return 'warning';
		if (l.startsWith('31')) return 'warning';
		if (l.startsWith('1')) return 'info';
		return (['success', 'info', 'warning', 'warning', 'danger'] as Tone[])[Math.min(i, 4)];
	}

	function accountBadge(t: string): string {
		switch (t) {
			case 'ASSET':
				return 'badge-info';
			case 'LIABILITY':
				return 'badge-warning';
			case 'EQUITY':
			case 'REVENUE':
				return 'badge-success';
			case 'EXPENSE':
				return 'badge-danger';
			default:
				return 'badge-neutral';
		}
	}

	const filteredReceivables = $derived(
		statusFilter ? receivables.filter((r) => r.Status === statusFilter) : receivables
	);

	const totalOutstanding = $derived(
		receivables.reduce((s, r) => s + Math.max(0, outstandingOf(r)), 0)
	);

	const customerOptions = $derived.by(() => {
		const opts: { value: string; label: string }[] = [];
		for (const r of receivables) {
			const value = String(r.CustomerID);
			if (!opts.some((o) => o.value === value)) opts.push({ value, label: customerName(r) });
		}
		return opts;
	});

	const tbTotals = $derived.by(() =>
		(trialRows ?? []).reduce(
			(acc, r) => {
				acc.debit += num(r.TotalDebit);
				acc.credit += num(r.TotalCredit);
				return acc;
			},
			{ debit: 0, credit: 0 }
		)
	);
	const tbBalanced = $derived(Math.abs(tbTotals.debit - tbTotals.credit) < 0.005);

	const grossMargin = $derived(
		income && num(income.Revenue) !== 0 ? num(income.GrossProfit) / num(income.Revenue) : null
	);
	const netMargin = $derived(
		income && num(income.Revenue) !== 0 ? num(income.NetIncome) / num(income.Revenue) : null
	);
	const netPositive = $derived(!!income && num(income.NetIncome) >= 0);

	const periodLabel = $derived(
		from || to
			? `${from ? formatDate(from) : 'Beginning'} — ${to ? formatDate(to) : 'Today'}`
			: 'All periods'
	);

	$effect(() => {
		if (!auth.hasPermission('reports.financial')) {
			toast.error('Access Denied', {
				description: 'You do not have permission to view finance.'
			});
			goto(resolve('/'));
		}
	});

	$effect(() => {
		const tab = stmtTab;
		if (auth.hasPermission('reports.financial')) loadStatement(tab);
	});

	onMount(() => {
		loadReceivables();
	});

	async function loadReceivables() {
		arLoading = true;
		try {
			const [aging, recv] = await Promise.all([financeApi.arAging(), financeApi.receivables()]);
			buckets = aging.buckets || [];
			receivables = recv.receivables || [];
		} catch (e) {
			console.error(e);
			toast.error('Failed to load receivables', { description: errMessage(e) });
		} finally {
			arLoading = false;
		}
	}

	async function loadStatement(tabOverride?: string) {
		const tab = tabOverride ?? untrack(() => stmtTab);
		const { f, t } = untrack(() => ({ f: from, t: to }));
		stmtLoading = true;
		try {
			const params: { from?: string; to?: string } = {};
			if (f) params.from = f;
			if (t) params.to = t;
			if (tab === 'income') income = await financeApi.incomeStatement(params);
			else if (tab === 'balance') sheet = await financeApi.balanceSheet(params);
			else trialRows = (await financeApi.trialBalance(params)).rows ?? [];
		} catch (e) {
			console.error(e);
			toast.error('Failed to load statement', { description: errMessage(e) });
		} finally {
			stmtLoading = false;
		}
	}

	function openPayment(ar?: AccountsReceivable) {
		payTarget = ar ?? null;
		payCustomerId = ar ? String(ar.CustomerID) : (customerOptions[0]?.value ?? '');
		payAmount = ar && outstandingOf(ar) > 0 ? outstandingOf(ar).toFixed(2) : '';
		payMethod = 'cash';
		payNotes = '';
		payOpen = true;
	}

	async function submitPayment() {
		const customerId = Number(payCustomerId);
		const amount = parseFloat(payAmount);
		if (!customerId) {
			toast.error('Select a customer');
			return;
		}
		if (!Number.isFinite(amount) || amount <= 0) {
			toast.error('Enter a valid payment amount');
			return;
		}
		paying = true;
		try {
			await financeApi.recordPayment({
				customerId,
				amount,
				method: payMethod,
				notes: payNotes.trim() || undefined,
				allocations: payTarget
					? [{ AccountsReceivableID: payTarget.ID, Amount: amount }]
					: undefined
			});
			toast.success('Payment recorded', {
				description: `${money(amount)} received via ${methodLabel(payMethod)}`
			});
			payOpen = false;
			await loadReceivables();
		} catch (e) {
			toast.error('Failed to record payment', { description: errMessage(e) });
		} finally {
			paying = false;
		}
	}

	function errMessage(e: unknown): string {
		const err = e as { response?: { data?: { error?: string } }; message?: string };
		return err?.response?.data?.error || err?.message || 'Unknown error';
	}
</script>

<div class="bg-background relative min-h-screen overflow-hidden p-6 lg:p-10">
	<div class="mx-auto max-w-7xl space-y-8">
		<div
			class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between"
			data-animate="fade-up"
		>
			<div class="space-y-1">
				<p class="flex items-center gap-2 overline">
					<span class="bg-primary/10 text-primary inline-flex rounded-md p-1.5">
						<Landmark class="h-4 w-4" />
					</span>
					Control · Double-Entry Ledger
				</p>
				<h1 class="font-display text-foreground text-3xl font-bold tracking-tight">Finance</h1>
				<p class="text-muted-foreground">
					Financial statements, trial balance & accounts receivable
				</p>
			</div>
			<Button
				onclick={() => openPayment()}
				class="bg-primary text-primary-foreground hover:bg-primary/90 gap-2 rounded-md shadow-md"
			>
				<Plus class="h-4 w-4" />
				Record Payment
			</Button>
		</div>

		<GlassTabs
			bind:value={activeTab}
			tabs={[
				{ value: 'receivables', label: 'Receivables' },
				{ value: 'statements', label: 'Financial Statements' }
			]}
			class="overflow-x-auto"
		/>

		{#if activeTab === 'receivables'}
			<div class="space-y-6">
				{#if arLoading}
					<div class="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-5">
						{#each [0, 80, 160, 240, 320] as delay (delay)}
							<div class="liquid-panel space-y-3 rounded-lg p-4">
								<Skeleton class="h-3 w-16" />
								<Skeleton class="h-7 w-24" style="animation-delay: {delay}ms;" />
								<Skeleton class="h-3 w-12" />
							</div>
						{/each}
					</div>
				{:else if buckets.length > 0}
					<div class="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-5" data-animate="fade-up">
						{#each buckets as b, i (b.Label)}
							{@const tone = bucketTone(b.Label, i)}
							<div
								class="liquid-panel group rounded-lg border-l-2 p-4 transition-all hover:-translate-y-0.5 hover:shadow-md {toneBorder[
									tone
								]}"
							>
								<p class="overline">{b.Label}</p>
								<p class="kpi-num mt-2 text-xl md:text-2xl {toneText[tone]}">
									{money(b.Amount)}
								</p>
								<p class="text-muted-foreground num-mono mt-1 text-xs">
									{b.Count} invoice{b.Count === 1 ? '' : 's'}
								</p>
							</div>
						{/each}
					</div>
				{/if}

				<div
					class="liquid-panel overflow-hidden rounded-lg"
					data-animate="fade-up"
					style="animation-delay: 80ms;"
				>
					<div
						class="border-border flex flex-col gap-3 border-b p-4 sm:flex-row sm:items-center sm:justify-between"
					>
						<div>
							<p class="overline">Accounts Receivable</p>
							<p class="text-foreground mt-0.5 text-sm font-medium">
								{filteredReceivables.length} receivable{filteredReceivables.length === 1 ? '' : 's'}
								·
								<span class="num-mono font-semibold">{money(totalOutstanding)}</span> outstanding
							</p>
						</div>
						<div class="flex items-center gap-2">
							<div class="w-40">
								<div class="select-wrapper">
									<select bind:value={statusFilter}>
										<option value="">All Statuses</option>
										<option value="UNPAID">Unpaid</option>
										<option value="PARTIAL">Partial</option>
										<option value="PAID">Paid</option>
									</select>
								</div>
							</div>
							<Button
								variant="outline"
								size="sm"
								class="{btnSecondary} gap-1"
								onclick={() => openPayment()}
							>
								<Plus class="h-3 w-3" />
								Payment
							</Button>
						</div>
					</div>

					{#if arLoading}
						<div class="divide-border/70 divide-y">
							{#each [0, 60, 120, 180, 240, 300] as delay (delay)}
								<div class="flex items-center gap-4 px-4 py-3.5">
									<Skeleton class="h-4 w-1/5" style="animation-delay: {delay}ms;" />
									<Skeleton class="h-4 w-1/6" />
									<Skeleton class="ml-auto h-4 w-16" />
									<Skeleton class="h-4 w-16" />
									<Skeleton class="h-4 w-16" />
									<Skeleton class="h-5 w-16 rounded-md" />
								</div>
							{/each}
						</div>
					{:else if receivables.length === 0}
						<EmptyState
							icon={ReceiptText}
							title="No receivables yet"
							description="Credit sales will appear here for collection."
							actionLabel="Record Payment"
							onAction={() => openPayment()}
						/>
					{:else if filteredReceivables.length === 0}
						<EmptyState
							icon={Search}
							title="No matching receivables"
							description="Nothing with this status — try another filter."
							actionLabel="Clear Filter"
							onAction={() => (statusFilter = '')}
						/>
					{:else}
						<div class="overflow-x-auto">
							<table class="ledger-table min-w-[56rem]">
								<thead>
									<tr>
										<th>Customer</th>
										<th>Order #</th>
										<th class="text-right">Amount</th>
										<th class="text-right">Paid</th>
										<th class="text-right">Outstanding</th>
										<th>Status</th>
										<th>Due Date</th>
										<th class="text-right">Actions</th>
									</tr>
								</thead>
								<tbody>
									{#each filteredReceivables as ar (ar.ID)}
										{@const out = outstandingOf(ar)}
										{@const overdue = isOverdue(ar)}
										<tr>
											<td>
												<div class="flex flex-col">
													<span class="text-foreground font-semibold">{customerName(ar)}</span>
													{#if ar.Customer?.Email}
														<span class="text-muted-foreground text-xs">{ar.Customer.Email}</span>
													{/if}
												</div>
											</td>
											<td class="num-mono text-muted-foreground">
												{ar.Order?.OrderNumber || `#${ar.OrderID}`}
											</td>
											<td class="col-num text-foreground">{money(ar.Amount)}</td>
											<td class="col-num text-muted-foreground">{money(ar.PaidAmount)}</td>
											<td
												class="col-num font-semibold {out > 0 ? 'text-foreground' : 'text-success'}"
											>
												{money(out)}
											</td>
											<td>
												<span
													class={cn(
														statusBadge[ar.Status] ?? 'badge-neutral',
														'tracking-wide uppercase'
													)}
												>
													{ar.Status}
												</span>
											</td>
											<td class="whitespace-nowrap">
												{#if ar.DueDate}
													<div class="flex flex-col">
														<span
															class={cn(
																'text-sm',
																overdue ? 'text-danger font-semibold' : 'text-muted-foreground'
															)}
														>
															{formatDate(ar.DueDate)}
														</span>
														{#if overdue}
															<span class="text-danger flex items-center gap-1 text-xs">
																<AlertTriangle class="h-3 w-3" />
																Overdue
															</span>
														{/if}
													</div>
												{:else}
													<span class="text-muted-foreground">—</span>
												{/if}
											</td>
											<td>
												<div class="flex justify-end">
													{#if ar.Status !== 'PAID'}
														<Button
															variant="outline"
															size="sm"
															class="{btnSecondary} h-7 gap-1 px-2 text-xs"
															onclick={() => openPayment(ar)}
														>
															<HandCoins class="h-3.5 w-3.5" />
															Payment
														</Button>
													{:else}
														<span class="text-muted-foreground text-xs">Settled</span>
													{/if}
												</div>
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
				</div>
			</div>
		{:else}
			<div class="space-y-6">
				<div
					class="liquid-panel flex flex-col gap-4 rounded-lg p-4 lg:flex-row lg:items-center lg:justify-between"
					data-animate="fade-up"
				>
					<GlassTabs
						bind:value={stmtTab}
						tabs={[
							{ value: 'income', label: 'Income Statement' },
							{ value: 'balance', label: 'Balance Sheet' },
							{ value: 'trial', label: 'Trial Balance' }
						]}
						class="overflow-x-auto"
					/>
					<div class="flex flex-wrap items-center gap-2">
						<Input
							type="date"
							bind:value={from}
							class="{inputClass} num-mono w-40"
							aria-label="From date"
						/>
						<span class="text-muted-foreground text-sm">→</span>
						<Input
							type="date"
							bind:value={to}
							class="{inputClass} num-mono w-40"
							aria-label="To date"
						/>
						<Button
							variant="outline"
							size="sm"
							class="{btnSecondary} gap-1"
							disabled={stmtLoading}
							onclick={() => loadStatement()}
						>
							{#if stmtLoading}
								<Loader2 class="h-3 w-3 animate-spin" />
							{:else}
								<Search class="h-3 w-3" />
							{/if}
							Apply
						</Button>
					</div>
				</div>

				{#key stmtTab}
					<div in:fly={{ y: 16, duration: 250 }}>
						{#if stmtLoading}
							<div class="liquid-panel space-y-4 rounded-lg p-6">
								<Skeleton class="h-4 w-44" />
								{#each [0, 70, 140, 210, 280, 350] as delay (delay)}
									<Skeleton class="h-9 w-full" style="animation-delay: {delay}ms;" />
								{/each}
							</div>
						{:else if stmtTab === 'income'}
							{#if income}
								<div class="liquid-panel overflow-hidden rounded-lg">
									<div class="border-border flex items-center justify-between border-b p-5">
										<div>
											<p class="flex items-center gap-1.5 overline">
												<Scale class="h-3.5 w-3.5" />
												Income Statement
											</p>
											<p class="text-muted-foreground mt-1 text-sm">{periodLabel}</p>
										</div>
									</div>
									<div class="mx-auto max-w-2xl px-6 py-6">
										<dl>
											<div class="flex items-baseline justify-between py-2">
												<dt class="text-foreground text-sm font-medium">Revenue</dt>
												<dd class="num-mono text-foreground text-sm font-semibold">
													{money(income.Revenue)}
												</dd>
											</div>
											<div class="flex items-baseline justify-between py-2 pl-6">
												<dt class="text-muted-foreground text-sm">− Cost of Goods Sold</dt>
												<dd class="num-mono text-muted-foreground text-sm">{money(income.COGS)}</dd>
											</div>
											<div
												class="border-border flex items-baseline justify-between border-t py-2.5"
											>
												<dt class="text-foreground text-sm font-semibold">
													Gross Profit
													{#if grossMargin !== null}
														<span class="text-muted-foreground num-mono text-xs font-normal">
															· {formatPercent(grossMargin)} margin
														</span>
													{/if}
												</dt>
												<dd class="num-mono text-foreground text-sm font-semibold">
													{money(income.GrossProfit)}
												</dd>
											</div>
											<div class="flex items-baseline justify-between py-2 pl-6">
												<dt class="text-muted-foreground text-sm">− Operating Expenses</dt>
												<dd class="num-mono text-muted-foreground text-sm">
													{money(income.Expenses)}
												</dd>
											</div>
											<div
												class="mt-3 flex items-center justify-between rounded-md border px-4 py-3 {netPositive
													? 'border-success/30 bg-success/10'
													: 'border-danger/30 bg-danger/10'}"
											>
												<dt
													class="font-display text-sm font-bold {netPositive
														? 'text-success'
														: 'text-danger'}"
												>
													Net Income
													{#if netMargin !== null}
														<span class="num-mono text-xs font-medium"
															>· {formatPercent(netMargin)}</span
														>
													{/if}
												</dt>
												<dd class="kpi-num text-xl {netPositive ? 'text-success' : 'text-danger'}">
													{money(income.NetIncome)}
												</dd>
											</div>
										</dl>
									</div>
								</div>
							{:else}
								<div class="liquid-panel rounded-lg">
									<EmptyState
										icon={Scale}
										title="No statement data"
										description="Post transactions, then run the report again."
									/>
								</div>
							{/if}
						{:else if stmtTab === 'balance'}
							{#if sheet}
								<div class="liquid-panel overflow-hidden rounded-lg">
									<div
										class="border-border flex flex-col gap-3 border-b p-5 sm:flex-row sm:items-center sm:justify-between"
									>
										<div>
											<p class="flex items-center gap-1.5 overline">
												<BookOpen class="h-3.5 w-3.5" />
												Balance Sheet
											</p>
											<p class="text-muted-foreground mt-1 text-sm">{periodLabel}</p>
										</div>
										{#if sheet.Balanced}
											<span class="badge-success w-fit">
												<CheckCircle2 class="h-3.5 w-3.5" />
												Balanced
											</span>
										{:else}
											<span class="badge-danger w-fit">
												<AlertTriangle class="h-3.5 w-3.5" />
												Out of balance
											</span>
										{/if}
									</div>

									<div class="bg-border grid gap-px md:grid-cols-2">
										<div class="bg-card p-6">
											<p class="overline">Assets</p>
											<p class="kpi-num text-foreground mt-3 text-3xl">{money(sheet.Assets)}</p>
											<p class="text-muted-foreground mt-2 text-xs">Everything the business owns</p>
										</div>
										<div class="bg-card p-6">
											<p class="overline">Liabilities & Equity</p>
											<div class="mt-3 space-y-2.5">
												<div class="flex items-baseline justify-between">
													<span class="text-foreground text-sm font-medium">Liabilities</span>
													<span class="num-mono text-foreground text-sm font-semibold">
														{money(sheet.Liabilities)}
													</span>
												</div>
												<div class="flex items-baseline justify-between">
													<span class="text-foreground text-sm font-medium">Equity</span>
													<span class="num-mono text-foreground text-sm font-semibold">
														{money(sheet.Equity)}
													</span>
												</div>
												<div class="flex items-baseline justify-between pl-4">
													<span class="text-muted-foreground text-xs">incl. Retained Earnings</span>
													<span class="num-mono text-muted-foreground text-xs">
														{money(sheet.RetainedEarnings)}
													</span>
												</div>
												<div
													class="border-border flex items-baseline justify-between border-t pt-2.5"
												>
													<span class="text-foreground text-sm font-bold"
														>Total Liabilities & Equity</span
													>
													<span class="num-mono text-foreground text-sm font-bold">
														{money(sheet.TotalLiabilitiesAndEquity)}
													</span>
												</div>
											</div>
										</div>
									</div>

									<div class="border-border bg-muted/50 border-t px-6 py-4">
										<div class="flex flex-wrap items-center justify-center gap-x-4 gap-y-2">
											<div class="text-center">
												<p class="overline">Assets</p>
												<p class="num-mono text-foreground text-lg font-bold">
													{money(sheet.Assets)}
												</p>
											</div>
											<span class="font-display text-muted-foreground text-xl">=</span>
											<div class="text-center">
												<p class="overline">Liabilities</p>
												<p class="num-mono text-foreground text-lg font-bold">
													{money(sheet.Liabilities)}
												</p>
											</div>
											<span class="font-display text-muted-foreground text-xl">+</span>
											<div class="text-center">
												<p class="overline">Equity</p>
												<p class="num-mono text-foreground text-lg font-bold">
													{money(sheet.Equity)}
												</p>
											</div>
										</div>
									</div>
								</div>
							{:else}
								<div class="liquid-panel rounded-lg">
									<EmptyState
										icon={BookOpen}
										title="No statement data"
										description="Post transactions, then run the report again."
									/>
								</div>
							{/if}
						{:else if trialRows && trialRows.length > 0}
							<div class="liquid-panel overflow-hidden rounded-lg">
								<div class="overflow-x-auto">
									<table class="ledger-table min-w-[52rem]">
										<thead>
											<tr>
												<th>Code</th>
												<th>Account</th>
												<th>Type</th>
												<th class="text-right">Debit</th>
												<th class="text-right">Credit</th>
												<th class="text-right">Balance</th>
											</tr>
										</thead>
										<tbody>
											{#each trialRows as row (row.AccountID)}
												<tr>
													<td class="num-mono text-muted-foreground">{row.Code}</td>
													<td class="text-foreground font-medium">{row.Name}</td>
													<td>
														<span class={cn(accountBadge(row.Type), 'tracking-wide uppercase')}>
															{row.Type}
														</span>
													</td>
													<td class="col-num text-foreground">
														{num(row.TotalDebit) !== 0 ? money(row.TotalDebit) : '—'}
													</td>
													<td class="col-num text-foreground">
														{num(row.TotalCredit) !== 0 ? money(row.TotalCredit) : '—'}
													</td>
													<td
														class="col-num font-semibold {num(row.Balance) < 0
															? 'text-danger'
															: 'text-foreground'}"
													>
														{money(row.Balance)}
													</td>
												</tr>
											{/each}
										</tbody>
										<tfoot>
											<tr class="border-border bg-muted/60 border-t-2">
												<td colspan="3" class="px-3.5 py-3 overline">
													Totals · {trialRows.length} account{trialRows.length === 1 ? '' : 's'}
												</td>
												<td
													class="num-mono text-foreground px-3.5 py-3 text-right text-sm font-bold"
												>
													{money(tbTotals.debit)}
												</td>
												<td
													class="num-mono text-foreground px-3.5 py-3 text-right text-sm font-bold"
												>
													{money(tbTotals.credit)}
												</td>
												<td class="px-3.5 py-3 text-right">
													{#if tbBalanced}
														<span class="badge-success">Debits = Credits ✓</span>
													{:else}
														<span class="badge-danger">
															Difference {money(Math.abs(tbTotals.debit - tbTotals.credit))}
														</span>
													{/if}
												</td>
											</tr>
										</tfoot>
									</table>
								</div>
							</div>
						{:else}
							<div class="liquid-panel rounded-lg">
								<EmptyState
									icon={BookOpen}
									title="No accounts to show"
									description="The ledger is empty for this period — adjust the date range or post transactions."
								/>
							</div>
						{/if}
					</div>
				{/key}
			</div>
		{/if}
	</div>
</div>

<Dialog.Root bind:open={payOpen}>
	<Dialog.Content class="max-h-[85vh] overflow-y-auto sm:max-w-[480px]">
		<Dialog.Header>
			<Dialog.Title>Record Payment</Dialog.Title>
			<Dialog.Description>
				{payTarget
					? `Applying to ${payTarget.Order?.OrderNumber ?? `receivable #${payTarget.ID}`}`
					: 'Log an incoming customer payment.'}
			</Dialog.Description>
		</Dialog.Header>

		<div class="space-y-4 py-4">
			{#if payTarget}
				<div
					class="border-border bg-muted/50 flex items-center justify-between rounded-md border px-3 py-2.5"
				>
					<div>
						<p class="text-foreground text-sm font-semibold">{customerName(payTarget)}</p>
						<p class="text-muted-foreground num-mono text-xs">
							{payTarget.Order?.OrderNumber || `#${payTarget.OrderID}`}
						</p>
					</div>
					<div class="text-right">
						<p class="overline">Outstanding</p>
						<p class="num-mono text-foreground text-sm font-bold">
							{money(outstandingOf(payTarget))}
						</p>
					</div>
				</div>
			{/if}

			<div class="space-y-2">
				<Label>Customer</Label>
				<div class="select-wrapper">
					<select bind:value={payCustomerId} disabled={!!payTarget}>
						{#each customerOptions as c (c.value)}
							<option value={c.value}>{c.label}</option>
						{/each}
					</select>
				</div>
			</div>

			<div class="space-y-2">
				<Label>Amount</Label>
				<Input
					type="number"
					min="0"
					step="0.01"
					bind:value={payAmount}
					placeholder="0.00"
					class="{inputClass} num-mono"
				/>
			</div>

			<div class="space-y-2">
				<Label>Method</Label>
				<div class="grid grid-cols-4 gap-2">
					{#each methods as m (m.value)}
						{@const Icon = m.icon}
						<button
							type="button"
							onclick={() => (payMethod = m.value)}
							class={cn(
								'flex flex-col items-center gap-1.5 rounded-md border px-2 py-2.5 text-xs font-semibold transition-all',
								payMethod === m.value
									? 'border-primary/40 bg-accent text-accent-foreground shadow-xs'
									: 'border-border bg-card text-muted-foreground hover:text-foreground hover:border-muted-foreground/40'
							)}
						>
							<Icon class="h-4 w-4" />
							{m.label}
						</button>
					{/each}
				</div>
			</div>

			<div class="space-y-2">
				<Label>Notes <span class="text-muted-foreground font-normal">(optional)</span></Label>
				<textarea
					bind:value={payNotes}
					rows="2"
					placeholder="Reference, cheque no, remarks..."
					class="liquid-textarea w-full px-3 py-2 text-sm"
				></textarea>
			</div>
		</div>

		<Dialog.Footer>
			<Button variant="outline" class={btnSecondary} onclick={() => (payOpen = false)}
				>Cancel</Button
			>
			<Button
				onclick={submitPayment}
				disabled={paying}
				class="bg-primary text-primary-foreground hover:bg-primary/90 gap-2"
			>
				{#if paying}
					<Loader2 class="h-4 w-4 animate-spin" />
				{:else}
					<HandCoins class="h-4 w-4" />
				{/if}
				Record Payment
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
