<script lang="ts">
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
import { alertsApi } from '$lib/api/resources';
import type { Alert } from '$lib/types';
import DetailsModal from '$lib/components/DetailsModal.svelte';
import type { DetailBuilderContext, DetailSection } from '$lib/components/DetailsModal.svelte';

// UI components
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
import { Input } from '$lib/components/ui/input';
import { Button } from '$lib/components/ui/button';
import { Skeleton } from '$lib/components/ui/skeleton';
import { cn } from '$lib/utils';

// Icon
import { Bell, Activity, CalendarClock, Package, AlertTriangle } from 'lucide-svelte';

	// --- State ---
	const filters = $state({ type: '', status: 'ACTIVE' });
	let alerts = $state<Alert[]>([]);
	let loading = $state(false);

	const productSettingsForm = $state({
		productId: '',
		lowStockLevel: '',
		overStockLevel: '',
		expiryAlertDays: ''
	});

	const userSettingsForm = $state({
		userId: '',
		emailNotificationsEnabled: true,
		smsNotificationsEnabled: false,
		emailAddress: '',
		phoneNumber: ''
	});

	const dateTimeFormatter = new Intl.DateTimeFormat('en-US', { dateStyle: 'medium', timeStyle: 'short' });

	const formatDateTime = (value?: string | null) => {
		if (!value) return '—';
		const date = new Date(value);
		return Number.isNaN(date.getTime()) ? '—' : dateTimeFormatter.format(date);
	};

	let detailModalOpen = $state(false);
	let detailAlertId = $state<number | null>(null);
	let detailModalSubtitle = $state<string | null>(null);

	const buildAlertSections = ({ data }: DetailBuilderContext): DetailSection[] => {
		const alert = data as Alert;
		const productName = alert.Product?.Name ?? `Product #${alert.ProductID}`;
		return [
			{
				type: 'summary',
				cards: [
					{
						title: 'Alert Type',
						value: alert.Type,
						hint: `Alert #${alert.ID}`,
						icon: AlertTriangle,
						accent: 'amber',
					},
					{
						title: 'Status',
						value: alert.Status,
						hint: alert.ResolvedAt ? `Resolved ${formatDateTime(alert.ResolvedAt)}` : 'Awaiting action',
						icon: Activity,
						accent: alert.Status === 'RESOLVED' ? 'emerald' : 'rose',
					},
					{
						title: 'Triggered',
						value: formatDateTime(alert.TriggeredAt),
						hint: productName,
						icon: CalendarClock,
						accent: 'slate',
					},
				],
			},
			{
				type: 'description',
				title: 'Alert Context',
				items: [
					{ label: 'Product', value: productName, icon: Package },
					{ label: 'Message', value: alert.Message },
					{ label: 'Product ID', value: `#${alert.ProductID}` },
					{ label: 'Triggered At', value: formatDateTime(alert.TriggeredAt) },
					{ label: 'Resolved At', value: formatDateTime(alert.ResolvedAt) },
				],
			},
			{
				type: 'description',
				title: 'Batch Details',
				items: [
					{ label: 'Batch ID', value: alert.BatchID ? `#${alert.BatchID}` : 'N/A' },
					{
						label: 'Batch Number',
						value: alert.Batch?.BatchNumber ?? '—',
					},
					{
						label: 'Quantity',
						value: alert.Batch ? `${alert.Batch.Quantity}` : '—',
					},
					{
						label: 'Expiry Date',
						value: alert.Batch?.ExpiryDate ? formatDateTime(alert.Batch.ExpiryDate) : '—',
					},
				],
			},
		];
	};

	const openAlertDetails = (alert: Alert) => {
		detailAlertId = alert.ID;
		detailModalSubtitle = alert.Product?.Name ?? `Alert #${alert.ID}`;
		detailModalOpen = true;
	};

	// --- Data ops ---
	const loadAlerts = async () => {
		loading = true;
		try {
			alerts = await alertsApi.list({
				type: filters.type || undefined,
				status: filters.status || undefined
			});
		} catch (error) {
			const errorMessage = error?.response?.data?.error || 'Unable to load alerts';
			toast.error('Failed to Load Alerts', { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const resolveAlert = async (alertId: number) => {
		try {
			await alertsApi.resolve(alertId);
			toast.success('Alert resolved');
			await loadAlerts();
		} catch (error) {
			const errorMessage = error?.response?.data?.error || 'Unable to resolve alert';
			toast.error('Failed to Resolve Alert', { description: errorMessage });
		}
	};

	const updateProductSettings = async () => {
		if (!productSettingsForm.productId) {
			toast.warning('Select a product');
			return;
		}
		try {
			await alertsApi.updateProductSettings(Number(productSettingsForm.productId), {
				lowStockLevel: Number(productSettingsForm.lowStockLevel) || 0,
				overStockLevel: Number(productSettingsForm.overStockLevel) || 0,
				expiryAlertDays: Number(productSettingsForm.expiryAlertDays) || 0
			});
			toast.success('Thresholds updated');
		} catch (error) {
			const errorMessage = error?.response?.data?.error || 'Unable to save thresholds';
			toast.error('Failed to Save Thresholds', { description: errorMessage });
		}
	};

	const updateUserSettings = async () => {
		if (!userSettingsForm.userId) {
			toast.warning('Provide a user ID');
			return;
		}
		try {
			await alertsApi.updateUserNotifications(Number(userSettingsForm.userId), {
				emailNotificationsEnabled: userSettingsForm.emailNotificationsEnabled,
				smsNotificationsEnabled: userSettingsForm.smsNotificationsEnabled,
				emailAddress: userSettingsForm.emailAddress,
				phoneNumber: userSettingsForm.phoneNumber
			});
			toast.success('Notification preferences saved');
		} catch (error) {
			const errorMessage = error?.response?.data?.error || 'Unable to save preferences';
			toast.error('Failed to Save Preferences', { description: errorMessage });
		}
	};

	onMount(() => {
		loadAlerts();

		// Parallax for hero (scoped & cleaned up)
		const hero = document.querySelector('.parallax-hero') as HTMLElement | null;
		if (!hero) return;
		const handleScroll = () => {
			const scrollY = window.scrollY / 6;
			hero.style.transform = `translateY(${scrollY}px)`;
		};
		window.addEventListener('scroll', handleScroll, { passive: true });
		return () => window.removeEventListener('scroll', handleScroll);
	});
</script>

<!-- HERO -->
<section class="relative w-full overflow-hidden bg-gradient-to-r from-amber-50 via-orange-50 to-rose-100 animate-gradientShift py-16 sm:py-20 px-4 sm:px-6">
	<!-- soft glass veil -->
	<div class="absolute inset-0 bg-white/40 backdrop-blur-sm"></div>

	<!-- floating blobs -->
	<div class="pointer-events-none absolute inset-0">
		<div class="absolute -top-16 -left-16 h-56 w-56 rounded-full bg-gradient-to-tr from-amber-200 to-orange-200 blur-3xl opacity-60 animate-heroFloat"></div>
		<div class="absolute top-1/3 -right-10 h-48 w-48 rounded-full bg-gradient-to-tr from-rose-200 to-orange-200 blur-3xl opacity-50 animate-heroFloat delay-300"></div>
	</div>

	<!-- content -->
	<div class="relative z-10 max-w-5xl mx-auto text-center">
		<div class="parallax-hero transform transition-transform duration-700 ease-out will-change-transform">
			<div class="mx-auto mb-5 w-fit rounded-2xl p-4 shadow-lg bg-gradient-to-br from-amber-400 to-orange-500 animate-pulseGlow">
				<Bell class="h-8 w-8 text-white" />
			</div>
			<h1 class="animate-fadeUp text-3xl sm:text-5xl font-extrabold bg-gradient-to-r from-amber-600 via-orange-600 to-rose-600 bg-clip-text text-transparent leading-tight">
				Alerts & Notifications Control
			</h1>
			<p class="animate-fadeUp delay-200 mt-3 text-slate-600 text-sm sm:text-base">
				Thresholds, escalations, and targeted user messaging — all in one soothing, vibrant cockpit.
			</p>
			<div class="animate-fadeUp delay-300 mt-6 flex flex-wrap justify-center gap-3">
				<Button
					variant="secondary"
					class="bg-white/80 border border-amber-200 text-amber-700 hover:bg-amber-50 rounded-xl font-medium shadow-sm hover:shadow-md hover:scale-105 transition-all"
					onclick={loadAlerts}
				>
					Refresh Alerts
				</Button>
				<a href="/operations">
					<Button class="bg-gradient-to-r from-amber-500 to-orange-600 hover:from-amber-600 hover:to-orange-700 text-white rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all">
						Go to Operations
					</Button>
				</a>
			</div>
		</div>
	</div>
</section>

<DetailsModal
	bind:open={detailModalOpen}
	resourceId={detailAlertId}
	endpoint="/alerts"
	title="Alert Details"
	subtitle={detailModalSubtitle}
	buildSections={buildAlertSections}
/>

<!-- MAIN -->
<section class="max-w-7xl mx-auto py-12 px-4 sm:px-6 space-y-8">
	<!-- Filters + Live Alerts -->
	<Card class="rounded-2xl border-0 shadow-lg overflow-hidden bg-gradient-to-br from-amber-50 to-orange-100 hover:shadow-xl transition-all">
		<CardHeader class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between bg-white/75 backdrop-blur px-6 py-5 border-b border-white/60">
			<div>
				<CardTitle class="text-slate-800">Live Alerts</CardTitle>
				<CardDescription class="text-slate-600">Filter by type or lifecycle state</CardDescription>
			</div>
			<div class="flex gap-2">
				<div class="select-wrapper w-[200px]">
<select class="rounded-xl border border-amber-200 bg-white/90 px-3 py-2 text-sm focus:ring-2 focus:ring-amber-400" bind:value={filters.type}>
					<option value="">All types</option>
					<option value="LOW_STOCK">Low stock</option>
					<option value="OVERSTOCK">Overstock</option>
					<option value="OUT_OF_STOCK">Out of stock</option>
					<option value="EXPIRY_ALERT">Expiry</option>
				</select>
				</div>
				
				<div class="select-wrapper w-[150px]">
<select class="rounded-xl border border-amber-200 bg-white/90 px-3 py-2 text-sm focus:ring-2 focus:ring-amber-400" bind:value={filters.status}>
					<option value="">Any status</option>
					<option value="ACTIVE">Active</option>
					<option value="RESOLVED">Resolved</option>
				</select>
				</div>
				
				<Button
					variant="secondary"
					class="bg-white/90 border border-amber-200 text-amber-700 hover:bg-amber-50 rounded-xl"
					onclick={loadAlerts}
				>
					Refresh
				</Button>
			</div>
		</CardHeader>
		<CardContent class="p-0">
			<Table>
				<TableHeader class="sticky top-0 z-10 bg-gradient-to-r from-amber-100/85 to-orange-100/85 backdrop-blur">
					<TableRow class="border-y border-amber-200/70">
						<TableHead class="px-4 py-3 text-slate-700">Type</TableHead>
						<TableHead class="px-4 py-3 text-slate-700">Product</TableHead>
						<TableHead class="px-4 py-3 text-slate-700">Message</TableHead>
						<TableHead class="px-4 py-3 text-slate-700">Status</TableHead>
						<TableHead class="px-4 py-3 text-right text-slate-700">Action</TableHead>
					</TableRow>
				</TableHeader>
				<TableBody class="[&>tr:nth-child(even)]:bg-white/70 [&>tr:nth-child(odd)]:bg-white/50">
					{#if loading}
						{#each Array(4) as _, i}
							<TableRow class="hover:bg-white/80">
								<TableCell colspan="5" class="px-4 py-3">
									<Skeleton class="h-6 w-full bg-white/70" />
								</TableCell>
							</TableRow>
						{/each}
					{:else if alerts.length === 0}
						<TableRow>
							<TableCell colspan="5" class="text-center text-sm text-slate-600 py-6">No alerts found</TableCell>
						</TableRow>
					{:else}
						{#each alerts as alert}
							<TableRow class="hover:bg-white/90 transition-colors cursor-pointer" onclick={() => openAlertDetails(alert)}>
								<TableCell class="px-4 py-3 text-xs font-semibold text-slate-700">{alert.Type}</TableCell>
								<TableCell class="px-4 py-3 text-slate-800">{alert.Product?.Name ?? `Product ${alert.ProductID}`}</TableCell>
								<TableCell class="px-4 py-3 text-sm text-slate-600">{alert.Message}</TableCell>
								<TableCell class="px-4 py-3">
									<span
										class={cn(
											'rounded-full px-2.5 py-1 text-xs capitalize border shadow-sm',
											alert.Status === 'ACTIVE'
												? 'bg-amber-100 text-amber-700 border-amber-200'
												: 'bg-slate-100 text-slate-600 border-slate-200'
										)}
									>
										{alert.Status}
									</span>
								</TableCell>
								<TableCell class="px-4 py-3 text-right">
									{#if alert.Status !== 'RESOLVED'}
										<Button
											size="sm"
											variant="ghost"
											class="text-amber-700 hover:bg-amber-100 rounded-lg px-3 py-1.5"
											onclick={(event) => {
												event.stopPropagation();
												resolveAlert(alert.ID);
											}}
										>
											Resolve
										</Button>
									{/if}
								</TableCell>
							</TableRow>
						{/each}
					{/if}
				</TableBody>
			</Table>
		</CardContent>
	</Card>

	<!-- Forms -->
	<div class="grid gap-8 lg:grid-cols-2">
		<!-- Product thresholds -->
		<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all hover:scale-[1.01] overflow-hidden bg-gradient-to-br from-emerald-50 to-green-100">
			<CardHeader class="bg-white/75 backdrop-blur px-6 py-5 border-b border-white/60">
				<CardTitle class="text-slate-800">Product Thresholds</CardTitle>
				<CardDescription class="text-slate-600">Configure alerting per SKU</CardDescription>
			</CardHeader>
			<CardContent class="space-y-3 p-6">
				<Input type="number" min="1" placeholder="Product ID" bind:value={productSettingsForm.productId} class="rounded-xl border-emerald-200 bg-white/90 focus:ring-2 focus:ring-emerald-400" />
				<div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
					<Input type="number" placeholder="Low stock" bind:value={productSettingsForm.lowStockLevel} class="rounded-xl border-emerald-200 bg-white/90 focus:ring-2 focus:ring-emerald-400" />
					<Input type="number" placeholder="Overstock" bind:value={productSettingsForm.overStockLevel} class="rounded-xl border-emerald-200 bg-white/90 focus:ring-2 focus:ring-emerald-400" />
					<Input type="number" placeholder="Expiry days" bind:value={productSettingsForm.expiryAlertDays} class="rounded-xl border-emerald-200 bg-white/90 focus:ring-2 focus:ring-emerald-400" />
				</div>
				<Button class="w-full bg-gradient-to-r from-emerald-500 to-green-600 hover:from-emerald-600 hover:to-green-700 text-white rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all" onclick={updateProductSettings}>
					Save thresholds
				</Button>
			</CardContent>
		</Card>

		<!-- User notifications -->
		<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all hover:scale-[1.01] overflow-hidden bg-gradient-to-br from-violet-50 to-purple-100">
			<CardHeader class="bg-white/75 backdrop-blur px-6 py-5 border-b border-white/60">
				<CardTitle class="text-slate-800">User Notifications</CardTitle>
				<CardDescription class="text-slate-600">Escalation preferences per operator</CardDescription>
			</CardHeader>
			<CardContent class="space-y-3 p-6">
				<Input type="number" min="1" placeholder="User ID" bind:value={userSettingsForm.userId} class="rounded-xl border-violet-200 bg-white/90 focus:ring-2 focus:ring-violet-400" />
				<Input placeholder="Email" bind:value={userSettingsForm.emailAddress} class="rounded-xl border-violet-200 bg-white/90 focus:ring-2 focus:ring-violet-400" />
				<Input placeholder="Phone" bind:value={userSettingsForm.phoneNumber} class="rounded-xl border-violet-200 bg-white/90 focus:ring-2 focus:ring-violet-400" />
				<div class="flex flex-wrap items-center gap-4 text-sm">
					<label class="flex items-center gap-2">
						<input type="checkbox" bind:checked={userSettingsForm.emailNotificationsEnabled} />
						Email
					</label>
					<label class="flex items-center gap-2">
						<input type="checkbox" bind:checked={userSettingsForm.smsNotificationsEnabled} />
						SMS
					</label>
				</div>
				<Button class="w-full bg-gradient-to-r from-violet-500 to-purple-600 hover:from-violet-600 hover:to-purple-700 text-white rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all" onclick={updateUserSettings}>
					Save preferences
				</Button>
			</CardContent>
		</Card>
	</div>
</section>

<style lang="postcss">
	@keyframes gradientShift {
		0% { background-position: 0% 50%; }
		50% { background-position: 100% 50%; }
		100% { background-position: 0% 50%; }
	}
	.animate-gradientShift {
		background-size: 200% 200%;
		animation: gradientShift 22s ease infinite;
	}

	@keyframes heroFloat {
		0%, 100% { transform: translateY(0) scale(1); }
		50% { transform: translateY(-6px) scale(1.02); }
	}
	.animate-heroFloat { animation: heroFloat 10s ease-in-out infinite; }

	@keyframes pulseGlow {
		0%, 100% { transform: scale(1); box-shadow: 0 0 14px rgba(251, 191, 36, 0.28); }
		50% { transform: scale(1.06); box-shadow: 0 0 24px rgba(251, 146, 60, 0.45); }
	}
	.animate-pulseGlow { animation: pulseGlow 7.5s ease-in-out infinite; }

	@keyframes fadeUp {
		from { opacity: 0; transform: translateY(18px); }
		to { opacity: 1; transform: translateY(0); }
	}
	.animate-fadeUp { animation: fadeUp 1s ease forwards; }

	/* Smooth material-like transitions */
	* {
		transition-property: color, background-color, border-color, text-decoration-color, fill, stroke, opacity, box-shadow, transform, filter, backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
	}

	/* Thin pastel scrollbar */
	::-webkit-scrollbar { width: 8px; height: 8px; }
	::-webkit-scrollbar-track { background: transparent; }
	::-webkit-scrollbar-thumb { background: rgba(251, 191, 36, 0.28); border-radius: 9999px; }
	::-webkit-scrollbar-thumb:hover { background: rgba(251, 146, 60, 0.4); }
</style>
