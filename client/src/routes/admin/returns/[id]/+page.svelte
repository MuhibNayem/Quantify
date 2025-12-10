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
	import { cn, formatDate, formatCurrency } from '$lib/utils';
	import { returnsApi } from '$lib/api/returns';
	import { toast } from 'svelte-sonner';
	import { liquidGlass } from '$lib/styles/liquid-glass';


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

<div class="relative min-h-screen overflow-hidden bg-[#F9FAFB] font-sans selection:bg-blue-100 selection:text-blue-900">
	<!-- Organic Mesh Gradient Background -->
	<div class="pointer-events-none absolute inset-0 overflow-hidden opacity-60">
		<div
			class="absolute left-[10%] top-[5%] h-[600px] w-[600px] rounded-full bg-gradient-to-br from-blue-200 via-cyan-100 to-transparent blur-[120px]"
		></div>
		<div
			class="absolute right-[5%] top-[30%] h-[500px] w-[500px] rounded-full bg-gradient-to-tr from-purple-200 via-pink-100 to-transparent blur-[100px]"
		></div>
		<div
			class="absolute bottom-[10%] left-[30%] h-[400px] w-[400px] rounded-full bg-gradient-to-tl from-indigo-200 via-violet-100 to-transparent blur-[90px]"
		></div>
	</div>

	<div class="relative mx-auto max-w-7xl px-6 py-10 lg:px-10">
		<!-- Back to Orders -->
		<button
			onclick={() => goto('/orders')}
			class="group mb-8 flex items-center gap-2 text-sm font-medium text-slate-500 transition-all hover:text-blue-600"
		>
			<div
				class="flex h-8 w-8 items-center justify-center rounded-full bg-white/50 shadow-sm ring-1 ring-slate-200/50 transition-all group-hover:bg-blue-50 group-hover:ring-blue-200"
			>
				<ArrowLeft size={16} />
			</div>
			Back to Orders
		</button>

		{#if loading}
			<div class="flex h-[60vh] flex-col items-center justify-center gap-4">
				<div class="relative h-12 w-12">
					<div class="absolute inset-0 animate-ping rounded-full bg-blue-100"></div>
					<div class="relative flex h-full w-full items-center justify-center rounded-full bg-gradient-to-br from-blue-500 to-indigo-600 text-white shadow-lg">
						<Package size={20} class="animate-pulse" />
					</div>
				</div>
				<p class="text-sm font-medium text-slate-500 animate-pulse">Loading return details...</p>
			</div>
		{:else if returnRequest}
			<div in:fade={{ duration: 400, delay: 100 }} class="space-y-8 pb-32">
				<!-- Header Section -->
				<div class="flex flex-col gap-6 md:flex-row md:items-start md:justify-between">
					<div class="space-y-2">
						<div class="flex items-center gap-4">
							<h1 class="text-4xl font-extrabold tracking-tight text-slate-900 drop-shadow-sm">
								Return <span class="bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">#{returnRequest.ID}</span>
							</h1>
							<div class={cn(
								'flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-bold uppercase tracking-wider shadow-sm ring-1 ring-inset backdrop-blur-md',
								returnRequest.Status === 'APPROVED' ? 'bg-emerald-50/60 text-emerald-600 ring-emerald-200/50' :
								returnRequest.Status === 'REJECTED' ? 'bg-red-50/60 text-red-600 ring-red-200/50' :
								'bg-amber-50/60 text-amber-600 ring-amber-200/50'
							)}>
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
						<div class="flex items-center gap-6 text-sm font-medium text-slate-500">
							<div class="flex items-center gap-2">
								<Calendar size={16} class="text-slate-400" />
								Requested {formatDate(returnRequest.CreatedAt)}
							</div>
							<div class="flex items-center gap-2">
								<Package size={16} class="text-slate-400" />
								{returnRequest.ReturnItems?.length || 0} Items
							</div>
						</div>
					</div>

					<div class="flex flex-col items-end gap-1">
						<span class="text-xs font-bold uppercase tracking-wide text-slate-400">Total Refund</span>
						<span class="text-4xl font-black tracking-tight text-slate-900 drop-shadow-sm">
							{formatCurrency(returnRequest.RefundAmount)}
						</span>
					</div>
				</div>

				<div class="grid gap-8 lg:grid-cols-2">
					<!-- Request Details Card -->
					<div class={cn(
						liquidGlass.radius.medium,
						liquidGlass.background.light,
						liquidGlass.blur.medium,
						liquidGlass.border.light,
						liquidGlass.shadow.medium,
						'overflow-hidden h-full flex flex-col'
					)}>
						<div class="border-b border-white/30 bg-white/10 px-6 py-4 backdrop-blur-md">
							<h3 class="flex items-center gap-2 font-bold text-slate-800">
								<Receipt size={18} class="text-blue-500" />
								Request Details
							</h3>
						</div>
						<div class="p-6 space-y-4 flex-1">
                            <div class="flex justify-between items-center py-2 border-b border-slate-100">
                                <span class="text-sm font-medium text-slate-500">Return ID</span>
                                <span class="font-bold text-slate-800">#{returnRequest.ID}</span>
                            </div>
                            <div class="flex justify-between items-center py-2 border-b border-slate-100">
                                <span class="text-sm font-medium text-slate-500">Date Requested</span>
                                <span class="font-medium text-slate-800">{formatDate(returnRequest.CreatedAt)}</span>
                            </div>
                            <div>
                                <span class="text-sm font-medium text-slate-500 block mb-2">Reason for Return</span>
                                <div class="rounded-xl bg-slate-50 p-4 text-sm text-slate-700 italic border border-slate-100">
                                    "{returnRequest.Reason}"
                                </div>
                            </div>
						</div>
					</div>

					<!-- Order & Customer Card -->
					<div class={cn(
						liquidGlass.radius.medium,
						liquidGlass.background.light,
						liquidGlass.blur.medium,
						liquidGlass.border.light,
						liquidGlass.shadow.medium,
						'overflow-hidden h-full flex flex-col'
					)}>
						<div class="border-b border-white/30 bg-white/10 px-6 py-4 backdrop-blur-md">
							<h3 class="flex items-center gap-2 font-bold text-slate-800">
								<User size={18} class="text-indigo-500" />
								Customer & Order
							</h3>
						</div>
						<div class="p-6 space-y-4 flex-1">
                            <div class="flex justify-between items-center py-2 border-b border-slate-100">
                                <span class="text-sm font-medium text-slate-500">Order Number</span>
                                <div class="flex items-center gap-2">
                                    <Package size={14} class="text-slate-400" />
                                    <span class="font-bold text-blue-600">{returnRequest.Order?.OrderNumber}</span>
                                </div>
                            </div>
                             <div class="flex justify-between items-center py-2 border-b border-slate-100">
                                <span class="text-sm font-medium text-slate-500">Customer Name</span>
                                <span class="font-medium text-slate-800">
                                    {returnRequest.User?.FirstName} {returnRequest.User?.LastName}
                                </span>
                            </div>
                            <div class="flex justify-between items-center py-2 border-b border-slate-100">
                                <span class="text-sm font-medium text-slate-500">Contact Email</span>
                                <div class="flex items-center gap-2">
                                    <Mail size={14} class="text-slate-400" />
                                    <span class="font-medium text-slate-800">{returnRequest.User?.Email}</span>
                                </div>
                            </div>
                             <div class="flex justify-between items-center py-2">
                                <span class="text-sm font-medium text-slate-500">Original Total</span>
                                <span class="font-bold text-slate-800">{formatCurrency(returnRequest.Order?.TotalAmount || 0)}</span>
                            </div>
						</div>
					</div>
				</div>

				<!-- Items List -->
				<div class={cn(
					liquidGlass.radius.medium,
					liquidGlass.background.light,
					liquidGlass.blur.heavy,
					liquidGlass.border.light,
					liquidGlass.shadow.medium,
					'overflow-hidden'
				)}>
					<div class="border-b border-white/30 bg-white/10 px-6 py-4 backdrop-blur-md">
						<h3 class="flex items-center gap-2 font-bold text-slate-800">
							<Package size={18} class="text-emerald-500" />
							Items to Return
						</h3>
					</div>
					<div class="p-2">
						<div class="hidden md:grid grid-cols-12 gap-4 px-4 py-3 text-xs font-bold uppercase tracking-wider text-slate-400 border-b border-slate-100">
							<div class="col-span-6">Product</div>
							<div class="col-span-2 text-center">Condition</div>
							<div class="col-span-2 text-center">Quantity</div>
							<div class="col-span-2 text-right">Refund Value</div>
						</div>

						{#each returnRequest.ReturnItems as item}
							<div class="group relative md:grid md:grid-cols-12 md:items-center md:gap-4 p-4 rounded-xl hover:bg-white/40 transition-all border-b border-transparent hover:border-white/50 last:border-0 md:border-0 border-slate-100">
								<!-- Mobile: Product & Image -->
								<div class="col-span-6 flex items-center gap-4 mb-4 md:mb-0">
									<div class="h-12 w-12 flex-shrink-0 overflow-hidden rounded-xl bg-white shadow-sm ring-1 ring-slate-100 flex items-center justify-center">
										{#if item.Product?.ImageURLs}
											<img src={item.Product.ImageURLs.split(',')[0]} alt={item.Product.Name} class="h-full w-full object-cover" />
										{:else}
											<Package size={20} class="text-slate-300" />
										{/if}
									</div>
									<div class="min-w-0">
										<p class="font-bold text-slate-800 truncate">{item.Product?.Name}</p>
										<p class="text-xs text-slate-500 font-medium">SKU: {item.Product?.SKU}</p>
                                        {#if item.Reason}
										    <p class="mt-1 text-xs text-amber-600/80 italic">"{item.Reason}"</p>
                                        {/if}
									</div>
								</div>

								<!-- Mobile: Details Grid -->
								<div class="col-span-6 md:col-start-7 grid grid-cols-3 md:grid-cols-6 gap-4 items-center">
									<!-- Condition -->
									<div class="col-span-1 md:col-span-2 flex flex-col md:items-center">
										<span class="text-[10px] uppercase font-bold text-slate-400 md:hidden mb-1">Condition</span>
										<span class="inline-flex items-center justify-center rounded-full bg-slate-100 px-2.5 py-0.5 text-xs font-bold text-slate-600 ring-1 ring-slate-200">
											{item.Condition}
										</span>
									</div>

									<!-- Quantity -->
									<div class="col-span-1 md:col-span-2 flex flex-col md:items-center">
										<span class="text-[10px] uppercase font-bold text-slate-400 md:hidden mb-1">Qty</span>
										<span class="font-bold text-slate-800">x{item.Quantity}</span>
									</div>

									<!-- Value -->
									<div class="col-span-1 md:col-span-2 flex flex-col items-end md:items-end">
										<span class="text-[10px] uppercase font-bold text-slate-400 md:hidden mb-1">Value</span>
										<span class="font-bold text-slate-900">{formatCurrency(item.Product?.SellingPrice * item.Quantity)}</span>
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
					class="fixed bottom-6 left-1/2 z-50 flex -translate-x-1/2 items-center gap-3 rounded-2xl border border-white/40 bg-white/80 p-2 pl-4 shadow-[0_30px_60px_-15px_rgba(0,0,0,0.15)] backdrop-blur-2xl md:bottom-10"
				>
                    <div class="flex items-center gap-2 pr-4 border-r border-slate-200">
                        <ShieldAlert size={18} class="text-amber-500" />
                        <span class="text-sm font-bold text-slate-600">Action Required</span>
                    </div>

					<button
						disabled={processing}
						onclick={() => handleProcess('reject')}
						class="rounded-xl px-5 py-2.5 text-sm font-bold text-red-600 transition-all hover:bg-red-50 disabled:opacity-50"
					>
						Reject
					</button>
					<button
						disabled={processing}
						onclick={() => handleProcess('approve')}
						class="flex items-center gap-2 rounded-xl bg-gradient-to-r from-emerald-500 to-teal-500 px-6 py-2.5 text-sm font-bold text-white shadow-lg shadow-emerald-200 transition-all hover:scale-105 hover:shadow-emerald-300 disabled:opacity-50 disabled:shadow-none"
					>
                        {#if processing}
                            <div class="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white"></div>
                        {:else}
                            <CheckCircle2 size={16} />
                        {/if}
						Approve Return
					</button>
				</div>
			{/if}
			
		{:else if !loading && !returnRequest}
			<div class="flex h-[50vh] flex-col items-center justify-center">
				<div class="liquid-panel rounded-[2rem] border-slate-200/50 bg-white/30 p-12 text-center backdrop-blur-xl">
					<div class="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-slate-100 text-slate-400">
						<Package size={32} />
					</div>
					<h3 class="text-xl font-bold text-slate-800">Return Not Found</h3>
					<p class="mt-2 text-slate-500">The requested return could not be located.</p>
					<button 
						onclick={() => goto('/orders')}
						class="mt-6 rounded-xl bg-white px-6 py-2.5 text-sm font-semibold text-slate-600 shadow-sm ring-1 ring-slate-200 transition-all hover:bg-slate-50 hover:shadow-md"
					>
						Back to Orders
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>

