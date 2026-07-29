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

<div
	class="relative min-h-screen overflow-hidden"
>
	<div class="relative mx-auto max-w-7xl px-6 py-10 lg:px-10">
		{#if returnRequest}
			<div class="mb-4">
				<Breadcrumbs
					items={[
						{ label: 'Returns', href: '/returns' },
						{ label: returnNumber(returnRequest.ID) }
					]}
				/>
			</div>
		{/if}
		<!-- Back to Orders -->
		<button
			onclick={() => goto('/orders')}
			class="group text-muted-foreground hover:text-primary mb-8 flex items-center gap-2 text-sm font-medium transition-all"
		>
			<div
				class="border-border bg-card flex h-8 w-8 items-center justify-center rounded-md border shadow-sm transition-all group-hover:bg-accent"
			>
				<ArrowLeft size={16} />
			</div>
			Back to Orders
		</button>

		{#if loading}
			<div class="flex h-[60vh] flex-col items-center justify-center gap-4">
				<div class="relative h-12 w-12">
					<div class="bg-primary/20 absolute inset-0 animate-ping rounded-full"></div>
					<div
						class="bg-primary text-primary-foreground relative flex h-full w-full items-center justify-center rounded-full shadow-md"
					>
						<Package size={20} class="animate-pulse" />
					</div>
				</div>
				<p class="text-muted-foreground animate-pulse text-sm font-medium">Loading return details...</p>
			</div>
		{:else if returnRequest}
			<div in:fade={{ duration: 400, delay: 100 }} class="space-y-8 pb-32">
				<!-- Header Section -->
				<div class="flex flex-col gap-6 md:flex-row md:items-start md:justify-between">
					<div class="space-y-2">
						<div class="flex items-center gap-4">
							<h1
								class="num-mono font-display text-foreground text-4xl font-extrabold tracking-tight"
							>
								{returnNumber(returnRequest.ID)}
							</h1>
							<div
								class={cn(
									'uppercase',
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
						<div class="text-muted-foreground flex items-center gap-6 text-sm font-medium">
							<div class="flex items-center gap-2">
								<Calendar size={16} class="text-muted-foreground" />
								Requested <span class="num">{formatDate(returnRequest.CreatedAt)}</span>
							</div>
							<div class="flex items-center gap-2">
								<Package size={16} class="text-muted-foreground" />
								<span class="num">{returnRequest.ReturnItems?.length || 0}</span> Items
							</div>
						</div>
					</div>

					<div class="flex flex-col items-end gap-1">
						<span class="overline">Total Refund</span>
						<span class="kpi-num text-foreground text-4xl">
							{formatCurrency(returnRequest.RefundAmount)}
						</span>
					</div>
				</div>

				<div class="grid gap-8 lg:grid-cols-2">
					<!-- Request Details Card -->
					<div class="liquid-panel flex h-full flex-col overflow-hidden rounded-lg">
						<div class="border-border bg-muted/50 border-b px-6 py-4">
							<h3 class="text-foreground flex items-center gap-2 font-bold">
								<Receipt size={18} class="text-info" />
								Request Details
							</h3>
						</div>
						<div class="flex-1 space-y-4 p-6">
							<div class="border-border flex items-center justify-between border-b py-2">
								<span class="text-muted-foreground text-sm font-medium">Return ID</span>
								<span class="num-mono text-foreground font-bold"
									>{returnNumber(returnRequest.ID)}</span
								>
							</div>
							<div class="border-border flex items-center justify-between border-b py-2">
								<span class="text-muted-foreground text-sm font-medium">Date Requested</span>
								<span class="text-foreground num font-medium">{formatDate(returnRequest.CreatedAt)}</span
								>
							</div>
							<div>
								<span class="text-muted-foreground mb-2 block text-sm font-medium">Reason for Return</span>
								<div
									class="border-border bg-muted text-foreground rounded-md border p-4 text-sm italic"
								>
									"{returnRequest.Reason}"
								</div>
							</div>
						</div>
					</div>

					<!-- Order & Customer Card -->
					<div class="liquid-panel flex h-full flex-col overflow-hidden rounded-lg">
						<div class="border-border bg-muted/50 border-b px-6 py-4">
							<h3 class="text-foreground flex items-center gap-2 font-bold">
								<User size={18} class="text-primary" />
								Customer & Order
							</h3>
						</div>
						<div class="flex-1 space-y-4 p-6">
							<div class="border-border flex items-center justify-between border-b py-2">
								<span class="text-muted-foreground text-sm font-medium">Order Number</span>
								<div class="flex items-center gap-2">
									<Package size={14} class="text-muted-foreground" />
									<span class="num-mono text-primary font-bold">{returnRequest.Order?.OrderNumber}</span>
								</div>
							</div>
							<div class="border-border flex items-center justify-between border-b py-2">
								<span class="text-muted-foreground text-sm font-medium">Customer Name</span>
								<span class="text-foreground font-medium">
									{returnRequest.User?.FirstName}
									{returnRequest.User?.LastName}
								</span>
							</div>
							<div class="border-border flex items-center justify-between border-b py-2">
								<span class="text-muted-foreground text-sm font-medium">Contact Email</span>
								<div class="flex items-center gap-2">
									<Mail size={14} class="text-muted-foreground" />
									<span class="text-foreground font-medium">{returnRequest.User?.Email}</span>
								</div>
							</div>
							<div class="flex items-center justify-between py-2">
								<span class="text-muted-foreground text-sm font-medium">Original Total</span>
								<span class="num-mono text-foreground font-bold"
									>{formatCurrency(returnRequest.Order?.TotalAmount || 0)}</span
								>
							</div>
						</div>
					</div>
				</div>

				<!-- Items List -->
				<div class="liquid-panel overflow-hidden rounded-lg">
					<div class="border-border bg-muted/50 border-b px-6 py-4">
						<h3 class="text-foreground flex items-center gap-2 font-bold">
							<Package size={18} class="text-success" />
							Items to Return
						</h3>
					</div>
					<div class="p-2">
						<div
							class="border-border text-muted-foreground hidden grid-cols-12 gap-4 border-b px-4 py-3 text-xs font-bold tracking-wider uppercase md:grid"
						>
							<div class="col-span-6">Product</div>
							<div class="col-span-2 text-center">Condition</div>
							<div class="col-span-2 text-center">Quantity</div>
							<div class="col-span-2 text-right">Refund Value</div>
						</div>

						{#each returnRequest.ReturnItems as item}
							<div
								class="hover:bg-muted/50 group relative rounded-md border-b border-transparent p-4 transition-all last:border-0 md:grid md:grid-cols-12 md:items-center md:gap-4 md:border-0"
							>
								<!-- Mobile: Product & Image -->
								<div class="col-span-6 mb-4 flex items-center gap-4 md:mb-0">
									<div
										class="border-border bg-card flex h-12 w-12 flex-shrink-0 items-center justify-center overflow-hidden rounded-md border shadow-sm"
									>
										{#if item.Product?.ImageURLs}
											<img
												src={item.Product.ImageURLs.split(',')[0]}
												alt={item.Product.Name}
												class="h-full w-full object-cover"
											/>
										{:else}
											<Package size={20} class="text-muted-foreground/50" />
										{/if}
									</div>
									<div class="min-w-0">
										<p class="text-foreground truncate font-bold">{item.Product?.Name}</p>
										<p class="text-muted-foreground num-mono text-xs font-medium">SKU: {item.Product?.SKU}</p>
										{#if item.Reason}
											<p class="text-warning/80 mt-1 text-xs italic">"{item.Reason}"</p>
										{/if}
									</div>
								</div>

								<!-- Mobile: Details Grid -->
								<div
									class="col-span-6 grid grid-cols-3 items-center gap-4 md:col-start-7 md:grid-cols-6"
								>
									<!-- Condition -->
									<div class="col-span-1 flex flex-col md:col-span-2 md:items-center">
										<span class="text-muted-foreground mb-1 text-[10px] font-bold uppercase md:hidden"
											>Condition</span
										>
										<span class="badge-neutral">
											{item.Condition}
										</span>
									</div>

									<!-- Quantity -->
									<div class="col-span-1 flex flex-col md:col-span-2 md:items-center">
										<span class="text-muted-foreground mb-1 text-[10px] font-bold uppercase md:hidden"
											>Qty</span
										>
										<span class="num-mono text-foreground font-bold">x{item.Quantity}</span>
									</div>

									<!-- Value -->
									<div class="col-span-1 flex flex-col items-end md:col-span-2 md:items-end">
										<span class="text-muted-foreground mb-1 text-[10px] font-bold uppercase md:hidden"
											>Value</span
										>
										<span class="num-mono text-foreground font-bold"
											>{formatCurrency(item.Product?.SellingPrice * item.Quantity)}</span
										>
									</div>
								</div>
							</div>
						{/each}
					</div>
				</div>
			</div>

			<!-- Floating Action Bar -->
			{#if returnRequest.Status === 'PENDING'}
				<div
					in:fly={{ y: 50, duration: 400 }}
					class="liquid-panel fixed bottom-6 left-1/2 z-50 flex -translate-x-1/2 items-center gap-3 rounded-lg p-2 pl-4 shadow-lg md:bottom-10"
				>
					<div class="border-border flex items-center gap-2 border-r pr-4">
						<ShieldAlert size={18} class="text-warning" />
						<span class="text-foreground text-sm font-bold">Action Required</span>
					</div>

					<button
						disabled={processing}
						onclick={() => handleProcess('reject')}
						class="text-danger hover:bg-danger/10 rounded-md px-5 py-2.5 text-sm font-bold transition-all disabled:opacity-50"
					>
						Reject
					</button>
					<button
						disabled={processing}
						onclick={() => handleProcess('approve')}
						class="bg-success flex items-center gap-2 rounded-md px-6 py-2.5 text-sm font-bold text-white shadow-sm transition-all hover:bg-success/90 disabled:opacity-50 disabled:shadow-none"
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
				<div
					class="liquid-panel rounded-lg p-12 text-center"
				>
					<div
						class="bg-muted text-muted-foreground mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-lg"
					>
						<Package size={32} />
					</div>
					<h3 class="text-foreground text-xl font-bold">Return Not Found</h3>
					<p class="text-muted-foreground mt-2">The requested return could not be located.</p>
					<button
						onclick={() => goto('/orders')}
						class="border-border bg-card text-muted-foreground hover:text-foreground hover:bg-muted mt-6 rounded-md border px-6 py-2.5 text-sm font-semibold shadow-sm transition-all"
					>
						Back to Orders
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>
