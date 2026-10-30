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
		Wallet,
		Truck
	} from 'lucide-svelte';
	import { BarChart, Tooltip } from 'layerchart';
	import type {
		AccountsReceivable,
		ARAgingBucket,
		AccountsPayable,
		APAgingBucket,
		TrialBalanceRow,
		IncomeStatement,
		BalanceSheet
	} from '$lib/types';

	const inputClass =
		'glass-input h-10 w-full px-3 text-sm font-normal text-foreground placeholder:text-muted-foreground';
	const btnSecondary = 'shadow-none';

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
	const toneColor: Record<Tone, string> = {
		success: 'hsl(var(--success))',
		info: 'hsl(var(--info))',
		warning: 'hsl(var(--warning))',
		danger: 'hsl(var(--danger))'
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

	let apLoading = $state(true);
	let apBuckets = $state<APAgingBucket[]>([]);
	let payables = $state<AccountsPayable[]>([]);
	let apStatusFilter = $state('');

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

	let spOpen = $state(false);
	let spPaying = $state(false);
	let spTarget = $state<AccountsPayable | null>(null);
	let spSupplierId = $state('');
	let spAmount = $state('');
	let spMethod = $state('cash');
	let spNotes = $state('');

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

	function outstandingOfAP(ap: AccountsPayable): number {
		return num(ap.Amount) - num(ap.PaidAmount);
	}

	function isOverdue(ar: AccountsReceivable): boolean {
		return !!ar.DueDate && ar.Status !== 'PAID' && new Date(ar.DueDate).getTime() < Date.now();
	}

	function isOverdueAP(ap: AccountsPayable): boolean {
		return !!ap.DueDate && ap.Status !== 'PAID' && new Date(ap.DueDate).getTime() < Date.now();
	}

	function customerName(ar: AccountsReceivable): string {
		return ar.Customer?.Username || `Customer #${ar.CustomerID}`;
	}

	function supplierName(ap: AccountsPayable): string {
		return ap.Supplier?.Name || `Supplier #${ap.SupplierID}`;
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

	const filteredPayables = $derived(
		apStatusFilter ? payables.filter((p) => p.Status === apStatusFilter) : payables
	);

	const totalAPOutstanding = $derived(
		payables.reduce((s, p) => s + Math.max(0, outstandingOfAP(p)), 0)
	);

	const supplierOptions = $derived.by(() => {
		const opts: { value: string; label: string }[] = [];
		for (const p of payables) {
			const value = String(p.SupplierID);
			if (!opts.some((o) => o.value === value)) opts.push({ value, label: supplierName(p) });
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

	type AgingPoint = { label: string; amount: number; count: number; unit: string; color: string };
	function toAging(
		rows: { Label: string; Amount: string; Count: number }[],
		unit: string
	): AgingPoint[] {
		return rows.map((b, i) => ({
			label: b.Label,
			amount: num(b.Amount),
			count: b.Count,
			unit,
			color: toneColor[bucketTone(b.Label, i)]
		}));
	}

	const arAgingData = $derived(toAging(buckets, 'invoice'));
	const arAgingPeak = $derived(Math.max(...arAgingData.map((d) => d.amount), 1));
	const arAgingDomain = $derived(arAgingData.map((d) => d.label));
	const arAgingRange = $derived(arAgingData.map((d) => d.color));
	const apAgingData = $derived(toAging(apBuckets, 'bill'));
	const apAgingPeak = $derived(Math.max(...apAgingData.map((d) => d.amount), 1));
	const apAgingDomain = $derived(apAgingData.map((d) => d.label));
	const apAgingRange = $derived(apAgingData.map((d) => d.color));

	const sheetChartData = $derived.by(() => {
		if (!sheet) return [] as { key: string; value: number; color: string }[];
		return [
			{ key: 'Assets', value: num(sheet.Assets), color: 'hsl(var(--primary))' },
			{
				key: 'Liabilities & Equity',
				value: num(sheet.TotalLiabilitiesAndEquity),
				color: 'hsl(var(--info))'
			}
		];
	});
	const sheetPeak = $derived(Math.max(...sheetChartData.map((d) => d.value), 1));
	const sheetDomain = $derived(sheetChartData.map((d) => d.key));
	const sheetRange = $derived(sheetChartData.map((d) => d.color));

	const incomeChartData = $derived.by(() => {
		if (!income) return [] as { key: string; value: number; color: string }[];
		const ni = num(income.NetIncome);
		return [
			{ key: 'Revenue', value: num(income.Revenue), color: 'hsl(var(--primary))' },
			{ key: 'COGS', value: num(income.COGS), color: 'hsl(var(--warning))' },
			{ key: 'Gross Profit', value: num(income.GrossProfit), color: 'hsl(var(--success))' },
			{ key: 'Expenses', value: num(income.Expenses), color: 'hsl(var(--danger))' },
			{
				key: 'Net Income',
				value: ni,
				color: ni >= 0 ? 'hsl(var(--success))' : 'hsl(var(--danger))'
			}
		];
	});
	const incomePeak = $derived(Math.max(...incomeChartData.map((d) => d.value), 1));
	const incomeDomain = $derived(incomeChartData.map((d) => d.key));
	const incomeRange = $derived(incomeChartData.map((d) => d.color));

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
		loadPayables();
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

	async function loadPayables() {
		apLoading = true;
		try {
			const [aging, pay] = await Promise.all([financeApi.apAging(), financeApi.payables()]);
			apBuckets = aging.buckets || [];
			payables = pay.payables || [];
		} catch (e) {
			console.error(e);
			toast.error('Failed to load payables', { description: errMessage(e) });
		} finally {
			apLoading = false;
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

	function openSupplierPayment(ap?: AccountsPayable) {
		spTarget = ap ?? null;
		spSupplierId = ap ? String(ap.SupplierID) : (supplierOptions[0]?.value ?? '');
		spAmount = ap && outstandingOfAP(ap) > 0 ? outstandingOfAP(ap).toFixed(2) : '';
		spMethod = 'cash';
		spNotes = '';
		spOpen = true;
	}

	async function submitSupplierPayment() {
		const supplierId = Number(spSupplierId);
		const amount = parseFloat(spAmount);
		if (!supplierId) {
			toast.error('Select a supplier');
			return;
		}
		if (!Number.isFinite(amount) || amount <= 0) {
			toast.error('Enter a valid payment amount');
			return;
		}
		spPaying = true;
		try {
			await financeApi.recordSupplierPayment({
				supplierId,
				amount,
				method: spMethod,
				notes: spNotes.trim() || undefined,
				allocations: spTarget ? [{ AccountsPayableID: spTarget.ID, Amount: amount }] : undefined
			});
			toast.success('Supplier payment recorded', {
				description: `${money(amount)} paid via ${methodLabel(spMethod)}`
			});
			spOpen = false;
			await loadPayables();
		} catch (e) {
			toast.error('Failed to record supplier payment', { description: errMessage(e) });
		} finally {
			spPaying = false;
		}
	}

	function errMessage(e: unknown): string {
		const err = e as { response?: { data?: { error?: string } }; message?: string };
		return err?.response?.data?.error || err?.message || 'Unknown error';
	}
</script>

{#snippet agingTip({ context }: { context: any })}
	<Tooltip.Root
		{context}
		x="data"
		y="pointer"
		anchor="top"
		yOffset={-12}
		variant="none"
		class="ledger-tip"
	>
		{#snippet children({ data })}
			<p class="ledger-tip-label">
				<span class="mr-1.5 inline-block h-2 w-2 rounded-full" style="background:{data.color}"
				></span>
				{data.label}
			</p>
			<p class="ledger-tip-value num-mono">
				{money(data.amount)}
				<span>{data.count} {data.unit}{data.count === 1 ? '' : 's'}</span>
			</p>
		{/snippet}
	</Tooltip.Root>
{/snippet}

{#snippet stmtTip({ context }: { context: any })}
	<Tooltip.Root
		{context}
		x="data"
		y="pointer"
		anchor="top"
		yOffset={-12}
		variant="none"
		class="ledger-tip"
	>
		{#snippet children({ data })}
			<p class="ledger-tip-label">
				<span class="mr-1.5 inline-block h-2 w-2 rounded-full" style="background:{data.color}"
				></span>
				{data.key}
			</p>
			<p class="ledger-tip-value num-mono">{money(data.value)}</p>
		{/snippet}
	</Tooltip.Root>
{/snippet}

<div class="min-h-screen bg-background p-6 lg:p-8">
	<div class="mx-auto max-w-7xl space-y-6">
		<header class="page-header" data-animate="fade-up">
			<div class="min-w-0">
				<p class="flex items-center gap-2 overline">
					<span class="inline-flex rounded-md bg-primary/10 p-1 text-primary">
						<Landmark class="h-3.5 w-3.5" />
					</span>
					Control · Double-Entry Ledger
				</p>
				<h1 class="page-title mt-2">Finance</h1>
				<p class="page-sub">Financial statements, trial balance, receivables & payables</p>
			</div>
			<Button
				onclick={() => openPayment()}
				class="btn-soothe h-10 gap-2 px-5 text-sm font-semibold"
			>
				<Plus class="h-4 w-4" />
				Record Payment
			</Button>
		</header>

		<GlassTabs
			bind:value={activeTab}
			tabs={[
				{ value: 'receivables', label: 'Receivables' },
				{ value: 'payables', label: 'Payables' },
				{ value: 'statements', label: 'Financial Statements' }
			]}
			class="overflow-x-auto"
		/>

		{#if activeTab === 'receivables'}
			<div class="space-y-6">
				{#if arLoading}
					<div class="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-5">
						{#each [0, 80, 160, 240, 320] as delay (delay)}
							<div class="panel space-y-3 p-4">
								<Skeleton class="h-3 w-16" />
								<Skeleton class="h-7 w-24" style="animation-delay: {delay}ms;" />
								<Skeleton class="h-3 w-12" />
							</div>
						{/each}
					</div>
				{:else if buckets.length > 0}
					<div class="panel overflow-hidden" data-animate="fade-up">
						<div class="panel-header">
							<div>
								<p class="overline">Aging Overview</p>
								<p class="mt-0.5 text-sm text-muted-foreground">
									Outstanding receivables by age bucket
								</p>
							</div>
						</div>
						<div class="px-4 pt-5 pb-4">
							<BarChart
								data={arAgingData}
								x="label"
								y="amount"
								c="label"
								cDomain={arAgingDomain}
								cRange={arAgingRange}
								yDomain={[0, arAgingPeak * 1.12]}
								padding={{ top: 16, right: 8, bottom: 4, left: 8 }}
								axis="x"
								grid={{ y: { stroke: 'hsl(var(--border))', dashArray: '2 4' } }}
								tooltip={agingTip}
								props={{
									xAxis: {
										tickMarks: false,
										tickLabelProps: { fill: 'hsl(var(--muted-foreground))', fontSize: 11 }
									}
								}}
								class="h-52 w-full"
							/>
						</div>
					</div>

					<div class="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-5" data-animate="fade-up">
						{#each buckets as b, i (b.Label)}
							{@const tone = bucketTone(b.Label, i)}
							<div
								class="panel group border-l-2 p-4 transition-all hover:-translate-y-0.5 hover:shadow-md {toneBorder[
									tone
								]}"
							>
								<p class="overline">{b.Label}</p>
								<p class="kpi-num mt-2 text-xl md:text-2xl {toneText[tone]}">
									{money(b.Amount)}
								</p>
								<p class="num-mono mt-1 text-xs text-muted-foreground">
									{b.Count} invoice{b.Count === 1 ? '' : 's'}
								</p>
							</div>
						{/each}
					</div>
				{/if}

				<div class="panel overflow-hidden" data-animate="fade-up" style="animation-delay: 80ms;">
					<div
						class="panel-header flex-col items-stretch gap-3 sm:flex-row sm:items-center sm:justify-between"
					>
						<div>
							<p class="overline">Accounts Receivable</p>
							<p class="mt-0.5 text-sm font-medium text-foreground">
								{filteredReceivables.length} receivable{filteredReceivables.length === 1 ? '' : 's'}
								·
								<span class="num-mono font-semibold">{money(totalOutstanding)}</span> outstanding
							</p>
						</div>
						<div class="flex items-center gap-2">
							<div class="w-40">
								<div class="select-wrapper">
									<select bind:value={statusFilter} class="h-10">
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
						<div class="divide-y divide-border/70">
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
													<span class="font-semibold text-foreground">{customerName(ar)}</span>
													{#if ar.Customer?.Email}
														<span class="text-xs text-muted-foreground">{ar.Customer.Email}</span>
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
																overdue ? 'font-semibold text-danger' : 'text-muted-foreground'
															)}
														>
															{formatDate(ar.DueDate)}
														</span>
														{#if overdue}
															<span class="flex items-center gap-1 text-xs text-danger">
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
														<span class="text-xs text-muted-foreground">Settled</span>
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
		{:else if activeTab === 'payables'}
			<div class="space-y-6">
				{#if apLoading}
					<div class="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-5">
						{#each [0, 80, 160, 240, 320] as delay (delay)}
							<div class="panel space-y-3 p-4">
								<Skeleton class="h-3 w-16" />
								<Skeleton class="h-7 w-24" style="animation-delay: {delay}ms;" />
								<Skeleton class="h-3 w-12" />
							</div>
						{/each}
					</div>
				{:else if apBuckets.length > 0}
					<div class="panel overflow-hidden" data-animate="fade-up">
						<div class="panel-header">
							<div>
								<p class="overline">Aging Overview</p>
								<p class="mt-0.5 text-sm text-muted-foreground">
									Outstanding payables by age bucket
								</p>
							</div>
						</div>
						<div class="px-4 pt-5 pb-4">
							<BarChart
								data={apAgingData}
								x="label"
								y="amount"
								c="label"
								cDomain={apAgingDomain}
								cRange={apAgingRange}
								yDomain={[0, apAgingPeak * 1.12]}
								padding={{ top: 16, right: 8, bottom: 4, left: 8 }}
								axis="x"
								grid={{ y: { stroke: 'hsl(var(--border))', dashArray: '2 4' } }}
								tooltip={agingTip}
								props={{
									xAxis: {
										tickMarks: false,
										tickLabelProps: { fill: 'hsl(var(--muted-foreground))', fontSize: 11 }
									}
								}}
								class="h-52 w-full"
							/>
						</div>
					</div>

					<div class="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-5" data-animate="fade-up">
						{#each apBuckets as b, i (b.Label)}
							{@const tone = bucketTone(b.Label, i)}
							<div
								class="panel group border-l-2 p-4 transition-all hover:-translate-y-0.5 hover:shadow-md {toneBorder[
									tone
								]}"
							>
								<p class="overline">{b.Label}</p>
								<p class="kpi-num mt-2 text-xl md:text-2xl {toneText[tone]}">
									{money(b.Amount)}
								</p>
								<p class="num-mono mt-1 text-xs text-muted-foreground">
									{b.Count} bill{b.Count === 1 ? '' : 's'}
								</p>
							</div>
						{/each}
					</div>
				{/if}

				<div class="panel overflow-hidden" data-animate="fade-up" style="animation-delay: 80ms;">
					<div
						class="panel-header flex-col items-stretch gap-3 sm:flex-row sm:items-center sm:justify-between"
					>
						<div>
							<p class="overline">Accounts Payable</p>
							<p class="mt-0.5 text-sm font-medium text-foreground">
								{filteredPayables.length} payable{filteredPayables.length === 1 ? '' : 's'}
								·
								<span class="num-mono font-semibold">{money(totalAPOutstanding)}</span> outstanding
							</p>
						</div>
						<div class="flex items-center gap-2">
							<div class="w-40">
								<div class="select-wrapper">
									<select bind:value={apStatusFilter} class="h-10">
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
								onclick={() => openSupplierPayment()}
							>
								<Plus class="h-3 w-3" />
								Payment
							</Button>
						</div>
					</div>

					{#if apLoading}
						<div class="divide-y divide-border/70">
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
					{:else if payables.length === 0}
						<EmptyState
							icon={Truck}
							title="No payables yet"
							description="Purchase orders on credit will appear here for payment."
							actionLabel="Record Payment"
							onAction={() => openSupplierPayment()}
						/>
					{:else if filteredPayables.length === 0}
						<EmptyState
							icon={Search}
							title="No matching payables"
							description="Nothing with this status — try another filter."
							actionLabel="Clear Filter"
							onAction={() => (apStatusFilter = '')}
						/>
					{:else}
						<div class="overflow-x-auto">
							<table class="ledger-table min-w-[56rem]">
								<thead>
									<tr>
										<th>Supplier</th>
										<th>PO #</th>
										<th class="text-right">Amount</th>
										<th class="text-right">Paid</th>
										<th class="text-right">Outstanding</th>
										<th>Status</th>
										<th>Due Date</th>
										<th class="text-right">Actions</th>
									</tr>
								</thead>
								<tbody>
									{#each filteredPayables as ap (ap.ID)}
										{@const out = outstandingOfAP(ap)}
										{@const overdue = isOverdueAP(ap)}
										<tr>
											<td>
												<div class="flex flex-col">
													<span class="font-semibold text-foreground">{supplierName(ap)}</span>
													{#if ap.PurchaseOrder?.OrderDate}
														<span class="text-xs text-muted-foreground">
															Ordered {formatDate(ap.PurchaseOrder.OrderDate)}
														</span>
													{/if}
												</div>
											</td>
											<td class="num-mono text-muted-foreground">
												{ap.PurchaseOrder ? `#${ap.PurchaseOrder.ID}` : `#${ap.PurchaseOrderID}`}
											</td>
											<td class="col-num text-foreground">{money(ap.Amount)}</td>
											<td class="col-num text-muted-foreground">{money(ap.PaidAmount)}</td>
											<td
												class="col-num font-semibold {out > 0 ? 'text-foreground' : 'text-success'}"
											>
												{money(out)}
											</td>
											<td>
												<span
													class={cn(
														statusBadge[ap.Status] ?? 'badge-neutral',
														'tracking-wide uppercase'
													)}
												>
													{ap.Status}
												</span>
											</td>
											<td class="whitespace-nowrap">
												{#if ap.DueDate}
													<div class="flex flex-col">
														<span
															class={cn(
																'text-sm',
																overdue ? 'font-semibold text-danger' : 'text-muted-foreground'
															)}
														>
															{formatDate(ap.DueDate)}
														</span>
														{#if overdue}
															<span class="flex items-center gap-1 text-xs text-danger">
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
													{#if ap.Status !== 'PAID'}
														<Button
															variant="outline"
															size="sm"
															class="{btnSecondary} h-7 gap-1 px-2 text-xs"
															onclick={() => openSupplierPayment(ap)}
														>
															<HandCoins class="h-3.5 w-3.5" />
															Payment
														</Button>
													{:else}
														<span class="text-xs text-muted-foreground">Settled</span>
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
					class="panel flex flex-col gap-4 p-4 lg:flex-row lg:items-center lg:justify-between"
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
						<span class="text-sm text-muted-foreground">→</span>
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
							<div class="panel space-y-4 p-6">
								<Skeleton class="h-4 w-44" />
								{#each [0, 70, 140, 210, 280, 350] as delay (delay)}
									<Skeleton class="h-9 w-full" style="animation-delay: {delay}ms;" />
								{/each}
							</div>
						{:else if stmtTab === 'income'}
							{#if income}
								<div class="panel overflow-hidden">
									<div class="panel-header">
										<div>
											<p class="flex items-center gap-1.5 overline">
												<Scale class="h-3.5 w-3.5" />
												Income Statement
											</p>
											<p class="mt-1 text-sm text-muted-foreground">{periodLabel}</p>
										</div>
									</div>
									<div class="border-b border-border px-5 pt-5 pb-4">
										<BarChart
											data={incomeChartData}
											x="key"
											y="value"
											c="key"
											cDomain={incomeDomain}
											cRange={incomeRange}
											yDomain={[0, incomePeak * 1.12]}
											padding={{ top: 16, right: 8, bottom: 4, left: 8 }}
											axis="x"
											grid={{ y: { stroke: 'hsl(var(--border))', dashArray: '2 4' } }}
											tooltip={stmtTip}
											props={{
												xAxis: {
													tickMarks: false,
													tickLabelProps: { fill: 'hsl(var(--muted-foreground))', fontSize: 11 }
												}
											}}
											class="h-56 w-full"
										/>
									</div>
									<div class="mx-auto max-w-2xl px-6 py-6">
										<dl>
											<div class="flex items-baseline justify-between py-2">
												<dt class="text-sm font-medium text-foreground">Revenue</dt>
												<dd class="num-mono text-sm font-semibold text-foreground">
													{money(income.Revenue)}
												</dd>
											</div>
											<div class="flex items-baseline justify-between py-2 pl-6">
												<dt class="text-sm text-muted-foreground">− Cost of Goods Sold</dt>
												<dd class="num-mono text-sm text-muted-foreground">{money(income.COGS)}</dd>
											</div>
											<div
												class="flex items-baseline justify-between border-t border-border py-2.5"
											>
												<dt class="text-sm font-semibold text-foreground">
													Gross Profit
													{#if grossMargin !== null}
														<span class="num-mono text-xs font-normal text-muted-foreground">
															· {formatPercent(grossMargin)} margin
														</span>
													{/if}
												</dt>
												<dd class="num-mono text-sm font-semibold text-foreground">
													{money(income.GrossProfit)}
												</dd>
											</div>
											<div class="flex items-baseline justify-between py-2 pl-6">
												<dt class="text-sm text-muted-foreground">− Operating Expenses</dt>
												<dd class="num-mono text-sm text-muted-foreground">
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
								<div class="panel">
									<EmptyState
										icon={Scale}
										title="No statement data"
										description="Post transactions, then run the report again."
									/>
								</div>
							{/if}
						{:else if stmtTab === 'balance'}
							{#if sheet}
								<div class="panel overflow-hidden">
									<div
										class="panel-header flex-col items-stretch gap-3 sm:flex-row sm:items-center sm:justify-between"
									>
										<div>
											<p class="flex items-center gap-1.5 overline">
												<BookOpen class="h-3.5 w-3.5" />
												Balance Sheet
											</p>
											<p class="mt-1 text-sm text-muted-foreground">{periodLabel}</p>
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

									<div class="border-b border-border px-5 pt-5 pb-4">
										<div class="mx-auto max-w-md">
											<BarChart
												data={sheetChartData}
												x="key"
												y="value"
												c="key"
												cDomain={sheetDomain}
												cRange={sheetRange}
												yDomain={[0, sheetPeak * 1.12]}
												padding={{ top: 16, right: 8, bottom: 4, left: 8 }}
												axis="x"
												grid={{ y: { stroke: 'hsl(var(--border))', dashArray: '2 4' } }}
												tooltip={stmtTip}
												props={{
													xAxis: {
														tickMarks: false,
														tickLabelProps: { fill: 'hsl(var(--muted-foreground))', fontSize: 11 }
													}
												}}
												class="h-48 w-full"
											/>
										</div>
										<p class="mt-2 text-center text-xs text-muted-foreground">
											{sheet.Balanced
												? 'Assets equal liabilities plus equity — the ledger balances.'
												: 'Assets do not equal liabilities plus equity — review postings.'}
										</p>
									</div>

									<div class="grid gap-px bg-border md:grid-cols-2">
										<div class="bg-card p-6">
											<p class="overline">Assets</p>
											<p class="kpi-num mt-3 text-3xl text-foreground">{money(sheet.Assets)}</p>
											<p class="mt-2 text-xs text-muted-foreground">Everything the business owns</p>
										</div>
										<div class="bg-card p-6">
											<p class="overline">Liabilities & Equity</p>
											<div class="mt-3 space-y-2.5">
												<div class="flex items-baseline justify-between">
													<span class="text-sm font-medium text-foreground">Liabilities</span>
													<span class="num-mono text-sm font-semibold text-foreground">
														{money(sheet.Liabilities)}
													</span>
												</div>
												<div class="flex items-baseline justify-between">
													<span class="text-sm font-medium text-foreground">Equity</span>
													<span class="num-mono text-sm font-semibold text-foreground">
														{money(sheet.Equity)}
													</span>
												</div>
												<div class="flex items-baseline justify-between pl-4">
													<span class="text-xs text-muted-foreground">incl. Retained Earnings</span>
													<span class="num-mono text-xs text-muted-foreground">
														{money(sheet.RetainedEarnings)}
													</span>
												</div>
												<div
													class="flex items-baseline justify-between border-t border-border pt-2.5"
												>
													<span class="text-sm font-bold text-foreground"
														>Total Liabilities & Equity</span
													>
													<span class="num-mono text-sm font-bold text-foreground">
														{money(sheet.TotalLiabilitiesAndEquity)}
													</span>
												</div>
											</div>
										</div>
									</div>

									<div class="border-t border-border bg-muted/50 px-6 py-4">
										<div class="flex flex-wrap items-center justify-center gap-x-4 gap-y-2">
											<div class="text-center">
												<p class="overline">Assets</p>
												<p class="num-mono text-lg font-bold text-foreground">
													{money(sheet.Assets)}
												</p>
											</div>
											<span class="font-display text-xl text-muted-foreground">=</span>
											<div class="text-center">
												<p class="overline">Liabilities</p>
												<p class="num-mono text-lg font-bold text-foreground">
													{money(sheet.Liabilities)}
												</p>
											</div>
											<span class="font-display text-xl text-muted-foreground">+</span>
											<div class="text-center">
												<p class="overline">Equity</p>
												<p class="num-mono text-lg font-bold text-foreground">
													{money(sheet.Equity)}
												</p>
											</div>
										</div>
									</div>
								</div>
							{:else}
								<div class="panel">
									<EmptyState
										icon={BookOpen}
										title="No statement data"
										description="Post transactions, then run the report again."
									/>
								</div>
							{/if}
						{:else if trialRows && trialRows.length > 0}
							<div class="panel overflow-hidden">
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
													<td class="font-medium text-foreground">{row.Name}</td>
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
											<tr class="border-t-2 border-border bg-muted/60">
												<td colspan="3" class="px-3.5 py-3 overline">
													Totals · {trialRows.length} account{trialRows.length === 1 ? '' : 's'}
												</td>
												<td
													class="num-mono px-3.5 py-3 text-right text-sm font-bold text-foreground"
												>
													{money(tbTotals.debit)}
												</td>
												<td
													class="num-mono px-3.5 py-3 text-right text-sm font-bold text-foreground"
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
							<div class="panel">
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
					class="flex items-center justify-between rounded-md border border-border bg-muted/50 px-3 py-2.5"
				>
					<div>
						<p class="text-sm font-semibold text-foreground">{customerName(payTarget)}</p>
						<p class="num-mono text-xs text-muted-foreground">
							{payTarget.Order?.OrderNumber || `#${payTarget.OrderID}`}
						</p>
					</div>
					<div class="text-right">
						<p class="overline">Outstanding</p>
						<p class="num-mono text-sm font-bold text-foreground">
							{money(outstandingOf(payTarget))}
						</p>
					</div>
				</div>
			{/if}

			<div class="space-y-2">
				<Label>Customer</Label>
				<div class="select-wrapper">
					<select bind:value={payCustomerId} disabled={!!payTarget} class="h-10">
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
									: 'border-border bg-card text-muted-foreground hover:border-muted-foreground/40 hover:text-foreground'
							)}
						>
							<Icon class="h-4 w-4" />
							{m.label}
						</button>
					{/each}
				</div>
			</div>

			<div class="space-y-2">
				<Label>Notes <span class="font-normal text-muted-foreground">(optional)</span></Label>
				<textarea
					bind:value={payNotes}
					rows="2"
					placeholder="Reference, cheque no, remarks..."
					class="glass-textarea w-full px-3 py-2 text-sm"
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
				class="btn-soothe h-10 gap-2 px-5 text-sm font-semibold"
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

<Dialog.Root bind:open={spOpen}>
	<Dialog.Content class="max-h-[85vh] overflow-y-auto sm:max-w-[480px]">
		<Dialog.Header>
			<Dialog.Title>Record Supplier Payment</Dialog.Title>
			<Dialog.Description>
				{spTarget
					? `Paying PO #${spTarget.PurchaseOrder?.ID ?? spTarget.PurchaseOrderID}`
					: 'Log an outgoing payment to a supplier.'}
			</Dialog.Description>
		</Dialog.Header>

		<div class="space-y-4 py-4">
			{#if spTarget}
				<div
					class="flex items-center justify-between rounded-md border border-border bg-muted/50 px-3 py-2.5"
				>
					<div>
						<p class="text-sm font-semibold text-foreground">{supplierName(spTarget)}</p>
						<p class="num-mono text-xs text-muted-foreground">
							PO #{spTarget.PurchaseOrder?.ID ?? spTarget.PurchaseOrderID}
						</p>
					</div>
					<div class="text-right">
						<p class="overline">Outstanding</p>
						<p class="num-mono text-sm font-bold text-foreground">
							{money(outstandingOfAP(spTarget))}
						</p>
					</div>
				</div>
			{/if}

			<div class="space-y-2">
				<Label>Supplier</Label>
				<div class="select-wrapper">
					<select bind:value={spSupplierId} disabled={!!spTarget} class="h-10">
						{#each supplierOptions as s (s.value)}
							<option value={s.value}>{s.label}</option>
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
					bind:value={spAmount}
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
							onclick={() => (spMethod = m.value)}
							class={cn(
								'flex flex-col items-center gap-1.5 rounded-md border px-2 py-2.5 text-xs font-semibold transition-all',
								spMethod === m.value
									? 'border-primary/40 bg-accent text-accent-foreground shadow-xs'
									: 'border-border bg-card text-muted-foreground hover:border-muted-foreground/40 hover:text-foreground'
							)}
						>
							<Icon class="h-4 w-4" />
							{m.label}
						</button>
					{/each}
				</div>
			</div>

			<div class="space-y-2">
				<Label>Notes <span class="font-normal text-muted-foreground">(optional)</span></Label>
				<textarea
					bind:value={spNotes}
					rows="2"
					placeholder="Reference, cheque no, remarks..."
					class="glass-textarea w-full px-3 py-2 text-sm"
				></textarea>
			</div>
		</div>

		<Dialog.Footer>
			<Button variant="outline" class={btnSecondary} onclick={() => (spOpen = false)}>Cancel</Button
			>
			<Button
				onclick={submitSupplierPayment}
				disabled={spPaying}
				class="btn-soothe h-10 gap-2 px-5 text-sm font-semibold"
			>
				{#if spPaying}
					<Loader2 class="h-4 w-4 animate-spin" />
				{:else}
					<HandCoins class="h-4 w-4" />
				{/if}
				Record Payment
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<style>
	:global(.ledger-tip) {
		background: hsl(var(--card));
		border: 1px solid hsl(var(--border));
		border-radius: var(--radius);
		box-shadow: var(--shadow-md);
		padding: 0.5rem 0.625rem;
		min-width: 6.5rem;
	}
	:global(.ledger-tip .ledger-tip-label) {
		display: flex;
		align-items: center;
		font-size: 0.6875rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: hsl(var(--muted-foreground));
	}
	:global(.ledger-tip .ledger-tip-value) {
		margin-top: 0.125rem;
		font-size: 0.9375rem;
		font-weight: 600;
		color: hsl(var(--foreground));
	}
	:global(.ledger-tip .ledger-tip-value span) {
		margin-left: 0.25rem;
		font-size: 0.6875rem;
		font-weight: 500;
		color: hsl(var(--muted-foreground));
	}
</style>
