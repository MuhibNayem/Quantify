<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { fade, fly } from 'svelte/transition';
	import { onMount } from 'svelte';
	import {
		ArrowLeft,
		Package,
		User,
		Receipt,
		CheckCircle2,
		XCircle,
		AlertCircle,
		Calendar,
		Phone,
		Mail,
		CreditCard,
		ShieldAlert,
		DollarSign
	} from 'lucide-svelte';
	import { cn, formatDate, formatCurrency, returnNumber } from '$lib/utils';
	import { returnsApi } from '$lib/api/returns';
	import { toast } from 'svelte-sonner';
	import Breadcrumbs from '$lib/components/ui/breadcrumbs.svelte';

	let returnId = $page.params.id;
	let returnRequest: any = null;
	let loading = true;
	let processing = false;

	onMount(async () => {
		try {
			const res = await returnsApi.getReturn(Number(returnId));
			returnRequest = res.return;
		} catch (error) {
			console.error('Failed to load return:', error);
			toast.error('Failed to load return details');
		} finally {
			loading = false;
		}
	});

	async function handleProcess(action: 'approve' | 'reject') {
		if (processing) return;
		processing = true;
		try {
			await returnsApi.processReturn(Number(returnId), action);
			toast.success(`Return request ${action}ed successfully`);
			const res = await returnsApi.getReturn(Number(returnId));
			returnRequest = res.return;
		} catch (error) {
			console.error(`Failed to ${action} return:`, error);
			toast.error(`Failed to ${action} return`);
		} finally {
			processing = false;
		}
	}
</script>

<div class="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8 lg:py-10">
	{#if returnRequest}
		<div class="mb-4">
			<Breadcrumbs
				items={[{ label: 'Returns', href: '/returns' }, { label: returnNumber(returnRequest.ID) }]}
			/>
		</div>
	{/if}

	<!-- Back to Orders -->
	<button
		type="button"
		onclick={() => goto('/orders')}
		class="group mb-6 flex items-center gap-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
	>
		<div
			class="flex h-8 w-8 items-center justify-center rounded-md border border-border/70 bg-card shadow-xs transition-colors group-hover:bg-muted"
		>
			<ArrowLeft size={16} />
		</div>
		Back to Orders
	</button>

	{#if loading}
		<div class="flex h-[60vh] flex-col items-center justify-center gap-4">
			<div class="relative h-12 w-12">
				<div class="absolute inset-0 animate-ping rounded-full bg-primary/20"></div>
				<div
					class="relative flex h-full w-full items-center justify-center rounded-full bg-primary text-primary-foreground shadow-md"
				>
					<Package size={20} class="animate-pulse" />
				</div>
			</div>
			<p class="animate-pulse text-sm font-medium text-muted-foreground">
				Loading return details...
			</p>
		</div>
	{:else if returnRequest}
		<div in:fade={{ duration: 400, delay: 100 }} class="space-y-4 pb-32">
			<!-- Header Section -->
			<header class="page-header" data-animate="fade-up">
				<div>
					<div class="flex flex-wrap items-center gap-3">
						<h1 class="num-mono page-title">{returnNumber(returnRequest.ID)}</h1>
						<div
							class={cn(
								returnRequest.Status === 'APPROVED'
									? 'badge-success'
									: returnRequest.Status === 'REJECTED'
										? 'badge-danger'
										: 'badge-warning'
							)}
						>
							{#if returnRequest.Status === 'APPROVED'}
								<CheckCircle2 size={12} />
							{:else if returnRequest.Status === 'REJECTED'}
								<XCircle size={12} />
							{:else}
								<AlertCircle size={12} />
							{/if}
							{returnRequest.Status}
						</div>
					</div>
					<div
						class="mt-2 flex flex-wrap items-center gap-5 text-sm font-medium text-muted-foreground"
					>
						<div class="flex items-center gap-2">
							<Calendar size={15} class="text-muted-foreground" />
							Requested <span class="num">{formatDate(returnRequest.CreatedAt)}</span>
						</div>
						<div class="flex items-center gap-2">
							<Package size={15} class="text-muted-foreground" />
							<span class="num">{returnRequest.ReturnItems?.length || 0}</span> Items
						</div>
					</div>
				</div>

				<div class="flex flex-col items-start gap-0.5 md:items-end">
					<span class="overline">Total Refund</span>
					<span class="kpi-num text-3xl text-foreground">
						{formatCurrency(returnRequest.RefundAmount)}
					</span>
				</div>
			</header>

			<div class="grid gap-4 lg:grid-cols-2">
				<!-- Request Details Card -->
				<section class="panel flex h-full flex-col overflow-hidden">
					<div class="panel-header">
						<h3 class="section-title flex items-center gap-2">
							<Receipt size={16} class="text-info" />
							Request Details
						</h3>
					</div>
					<div class="panel-body space-y-0">
						<div class="flex items-center justify-between border-b border-border/60 py-2.5">
							<span class="text-sm font-medium text-muted-foreground">Return ID</span>
							<span class="num-mono text-sm font-semibold text-foreground"
								>{returnNumber(returnRequest.ID)}</span
							>
						</div>
						<div class="flex items-center justify-between border-b border-border/60 py-2.5">
							<span class="text-sm font-medium text-muted-foreground">Date Requested</span>
							<span class="num text-sm font-medium text-foreground"
								>{formatDate(returnRequest.CreatedAt)}</span
							>
						</div>
						<div class="pt-3">
							<span class="mb-2 block text-sm font-medium text-muted-foreground"
								>Reason for Return</span
							>
							<div class="rounded-md bg-muted/50 p-4 text-sm text-foreground italic">
								"{returnRequest.Reason}"
							</div>
						</div>
					</div>
				</section>

				<!-- Order & Customer Card -->
				<section class="panel flex h-full flex-col overflow-hidden">
					<div class="panel-header">
						<h3 class="section-title flex items-center gap-2">
							<User size={16} class="text-primary" />
							Customer & Order
						</h3>
					</div>
					<div class="panel-body space-y-0">
						<div class="flex items-center justify-between border-b border-border/60 py-2.5">
							<span class="text-sm font-medium text-muted-foreground">Order Number</span>
							<div class="flex items-center gap-2">
								<Package size={14} class="text-muted-foreground" />
								<span class="num-mono text-sm font-semibold text-primary"
									>{returnRequest.Order?.OrderNumber}</span
								>
							</div>
						</div>
						<div class="flex items-center justify-between border-b border-border/60 py-2.5">
							<span class="text-sm font-medium text-muted-foreground">Customer Name</span>
							<span class="text-sm font-medium text-foreground">
								{returnRequest.User?.FirstName}
								{returnRequest.User?.LastName}
							</span>
						</div>
						<div class="flex items-center justify-between border-b border-border/60 py-2.5">
							<span class="text-sm font-medium text-muted-foreground">Contact Email</span>
							<div class="flex items-center gap-2">
								<Mail size={14} class="text-muted-foreground" />
								<span class="text-sm font-medium text-foreground">{returnRequest.User?.Email}</span>
							</div>
						</div>
						<div class="flex items-center justify-between py-2.5">
							<span class="text-sm font-medium text-muted-foreground">Original Total</span>
							<span class="num-mono text-sm font-semibold text-foreground"
								>{formatCurrency(returnRequest.Order?.TotalAmount || 0)}</span
							>
						</div>
					</div>
				</section>
			</div>

			<!-- Items List -->
			<section class="panel overflow-hidden">
				<div class="panel-header">
					<h3 class="section-title flex items-center gap-2">
						<Package size={16} class="text-success" />
						Items to Return
					</h3>
				</div>
				<div class="overflow-x-auto">
					<table class="ledger-table min-w-[36rem]">
						<thead>
							<tr>
								<th>Product</th>
								<th>Condition</th>
								<th class="text-right">Quantity</th>
								<th class="text-right">Refund Value</th>
							</tr>
						</thead>
						<tbody>
							{#each returnRequest.ReturnItems as item}
								<tr>
									<td>
										<div class="flex items-center gap-3">
											<div
												class="flex h-11 w-11 shrink-0 items-center justify-center overflow-hidden rounded-md border border-border/70 bg-muted/40"
											>
												{#if item.Product?.ImageURLs}
													<img
														src={item.Product.ImageURLs.split(',')[0]}
														alt={item.Product.Name}
														class="h-full w-full object-cover"
													/>
												{:else}
													<Package size={18} class="text-muted-foreground/60" />
												{/if}
											</div>
											<div class="min-w-0">
												<p class="truncate font-semibold text-foreground">{item.Product?.Name}</p>
												<p class="num-mono text-xs font-medium text-muted-foreground">
													SKU: {item.Product?.SKU}
												</p>
												{#if item.Reason}
													<p class="mt-0.5 text-xs text-warning italic">"{item.Reason}"</p>
												{/if}
											</div>
										</div>
									</td>
									<td>
										<span class="badge-neutral">{item.Condition}</span>
									</td>
									<td class="col-num font-semibold text-foreground">
										x{item.Quantity}
									</td>
									<td class="col-num font-semibold text-foreground">
										{formatCurrency(item.Product?.SellingPrice * item.Quantity)}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</section>
		</div>

		<!-- Floating Action Bar -->
		{#if returnRequest.Status === 'PENDING'}
			<div
				in:fly={{ y: 50, duration: 400 }}
				class="panel fixed bottom-6 left-1/2 z-50 flex -translate-x-1/2 items-center gap-2 rounded-full p-1.5 pr-2 pl-5 shadow-xl md:bottom-10"
			>
				<div class="flex items-center gap-2 border-r border-border/70 pr-4">
					<ShieldAlert size={16} class="text-warning" />
					<span class="text-sm font-semibold whitespace-nowrap text-foreground"
						>Action Required</span
					>
				</div>

				<button
					type="button"
					disabled={processing}
					onclick={() => handleProcess('reject')}
					class="rounded-full px-4 py-2 text-sm font-semibold text-danger transition-colors hover:bg-danger/10 disabled:opacity-50"
				>
					Reject
				</button>
				<button
					type="button"
					disabled={processing}
					onclick={() => handleProcess('approve')}
					class="btn-soothe flex items-center gap-2 rounded-full px-5 py-2.5 text-sm font-semibold disabled:opacity-50"
				>
					{#if processing}
						<div
							class="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white"
						></div>
					{:else}
						<CheckCircle2 size={16} />
					{/if}
					Approve Return
				</button>
			</div>
		{/if}
	{:else if !loading && !returnRequest}
		<div class="flex h-[50vh] flex-col items-center justify-center">
			<div class="panel px-10 py-12 text-center">
				<div
					class="mx-auto mb-5 flex h-14 w-14 items-center justify-center rounded-lg bg-muted text-muted-foreground"
				>
					<Package size={26} />
				</div>
				<h3 class="font-display text-lg font-semibold text-foreground">Return Not Found</h3>
				<p class="mt-1.5 text-sm text-muted-foreground">
					The requested return could not be located.
				</p>
				<button
					type="button"
					onclick={() => goto('/orders')}
					class="btn-soft mt-5 inline-flex h-10 items-center rounded-md px-4 text-sm font-medium"
				>
					Back to Orders
				</button>
			</div>
		</div>
	{/if}
</div>
