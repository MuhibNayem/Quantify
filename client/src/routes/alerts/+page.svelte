<script lang="ts">
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { alertsApi } from '$lib/api/resources';
	import type { Alert } from '$lib/types';
	import { t } from '$lib/i18n';
	import DetailsModal from '$lib/components/DetailsModal.svelte';
	import type { DetailBuilderContext, DetailSection } from '$lib/components/DetailsModal.svelte';

	// UI components
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { cn } from '$lib/utils';
	import ProductSelector from '$lib/components/ui/product-selector.svelte';
	import UserSelector from '$lib/components/ui/user-selector.svelte';

	// Icon
	import { Bell, Activity, CalendarClock, Package, AlertTriangle } from 'lucide-svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	// --- State ---
	const filters = $state({ type: '', status: 'ACTIVE' });
	let alerts = $state<Alert[]>([]);
	let loading = $state(false);

	$effect(() => {
		if (!auth.hasPermission('alerts.view')) {
			toast.error($t('alerts.toasts.access_denied'), { description: $t('alerts.toasts.access_denied_desc') });
			goto('/');
		}
	});

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

	const dateTimeFormatter = new Intl.DateTimeFormat('en-US', {
		dateStyle: 'medium',
		timeStyle: 'short'
	});

	const formatDateTime = (value?: string | null) => {
		if (!value) return '—';
		const date = new Date(value);
		return Number.isNaN(date.getTime()) ? '—' : dateTimeFormatter.format(date);
	};

	let detailModalOpen = $state(false);
	let detailAlertId = $state<number | null>(null);
	let detailModalSubtitle = $state<string | null>(null);

	const buildAlertSections = ({ data }: DetailBuilderContext): DetailSection[] => {
		const alert = data as unknown as Alert;
		const productName = alert.Product?.Name ?? `Product #${alert.ProductID}`;
		return [
			{
				type: 'summary',
				cards: [
					{
						title: $t('alerts.details.type'),
						value: alert.Type,
						hint: `Alert #${alert.ID}`,
						icon: AlertTriangle,
						accent: 'amber'
					},
					{
						title: $t('alerts.details.status'),
						value: alert.Status,
						hint:
							alert.Status === 'RESOLVED'
								? `${$t('alerts.details.resolved_at')} ${formatDateTime(alert.UpdatedAt)}`
								: $t('alerts.details.awaiting_action'),
						icon: Activity,
						accent: alert.Status === 'RESOLVED' ? 'emerald' : 'rose'
					},
					{
						title: $t('alerts.details.triggered'),
						value: formatDateTime(alert.TriggeredAt),
						hint: productName,
						icon: CalendarClock,
						accent: 'slate'
					}
				]
			},
			{
				type: 'description',
				title: $t('alerts.details.context'),
				items: [
					{ label: $t('alerts.details.product'), value: productName, icon: Package },
					{ label: $t('alerts.details.message'), value: alert.Message },
					{ label: $t('alerts.details.product_id'), value: `#${alert.ProductID}` },
					{ label: $t('alerts.details.triggered_at'), value: formatDateTime(alert.TriggeredAt) },
					{ label: $t('alerts.details.updated_at'), value: formatDateTime(alert.UpdatedAt) }
				]
			},
			{
				type: 'description',
				title: $t('alerts.details.batch_details'),
				items: [
					{ label: $t('alerts.details.batch_id'), value: alert.BatchID ? `#${alert.BatchID}` : $t('alerts.details.na') },
					{
						label: $t('alerts.details.batch_number'),
						value: alert.Batch?.BatchNumber ?? '—'
					},
					{
						label: $t('alerts.details.quantity'),
						value: alert.Batch ? `${alert.Batch.Quantity}` : '—'
					},
					{
						label: $t('alerts.details.expiry_date'),
						value: alert.Batch?.ExpiryDate ? formatDateTime(alert.Batch.ExpiryDate) : '—'
					}
				]
			}
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
		} catch (error: any) {
			const errorMessage = error?.response?.data?.error || $t('alerts.toasts.load_fail');
			toast.error($t('alerts.toasts.load_fail'), { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const resolveAlert = async (alertId: number) => {
		try {
			await alertsApi.resolve(alertId);
			toast.success($t('alerts.toasts.resolve_success'));
			await loadAlerts();
		} catch (error: any) {
			const errorMessage = error?.response?.data?.error || $t('alerts.toasts.resolve_fail');
			toast.error($t('alerts.toasts.resolve_fail'), { description: errorMessage });
		}
	};

	const updateProductSettings = async () => {
		if (!productSettingsForm.productId) {
			toast.warning($t('alerts.toasts.select_product'));
			return;
		}
		try {
			await alertsApi.updateProductSettings(Number(productSettingsForm.productId), {
				lowStockLevel: Number(productSettingsForm.lowStockLevel) || 0,
				overStockLevel: Number(productSettingsForm.overStockLevel) || 0,
				expiryAlertDays: Number(productSettingsForm.expiryAlertDays) || 0
			});
			toast.success($t('alerts.toasts.thresholds_updated'));
		} catch (error: any) {
			const errorMessage = error?.response?.data?.error || $t('alerts.toasts.thresholds_fail');
			toast.error($t('alerts.toasts.thresholds_fail'), { description: errorMessage });
		}
	};

	const updateUserSettings = async () => {
		if (!userSettingsForm.userId) {
			toast.warning($t('alerts.toasts.provide_user'));
			return;
		}
		try {
			await alertsApi.updateUserNotifications(Number(userSettingsForm.userId), {
				emailNotificationsEnabled: userSettingsForm.emailNotificationsEnabled,
				smsNotificationsEnabled: userSettingsForm.smsNotificationsEnabled,
				emailAddress: userSettingsForm.emailAddress,
				phoneNumber: userSettingsForm.phoneNumber
			});
			toast.success($t('alerts.toasts.prefs_saved'));
		} catch (error: any) {
			const errorMessage = error?.response?.data?.error || $t('alerts.toasts.prefs_fail');
			toast.error($t('alerts.toasts.prefs_fail'), { description: errorMessage });
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
<section
	class="animate-gradientShift relative w-full overflow-hidden bg-gradient-to-r from-amber-50 via-orange-50 to-rose-100 px-4 py-16 sm:px-6 sm:py-20"
>
	<!-- soft glass veil -->
	<div class="absolute inset-0 bg-white/40 backdrop-blur-sm"></div>

	<!-- floating blobs -->
	<div class="pointer-events-none absolute inset-0">
		<div
			class="animate-heroFloat absolute -left-16 -top-16 h-56 w-56 rounded-full bg-gradient-to-tr from-amber-200 to-orange-200 opacity-60 blur-3xl"
		></div>
		<div
			class="animate-heroFloat absolute -right-10 top-1/3 h-48 w-48 rounded-full bg-gradient-to-tr from-rose-200 to-orange-200 opacity-50 blur-3xl delay-300"
		></div>
	</div>

	<!-- content -->
	<div class="relative z-10 mx-auto max-w-5xl text-center">
		<div
			class="parallax-hero transform transition-transform duration-700 ease-out will-change-transform"
		>
			<div
				class="animate-pulseGlow mx-auto mb-5 w-fit rounded-2xl bg-gradient-to-br from-amber-400 to-orange-500 p-4 shadow-lg"
			>
				<Bell class="h-8 w-8 text-white" />
			</div>
			<h1
				class="animate-fadeUp bg-gradient-to-r from-amber-600 via-orange-600 to-rose-600 bg-clip-text text-3xl font-extrabold leading-tight text-transparent sm:text-5xl"
			>
				{$t('alerts.title')}
			</h1>
			<p class="animate-fadeUp mt-3 text-sm text-slate-600 delay-200 sm:text-base">
				{$t('alerts.subtitle')}
			</p>
			<div class="animate-fadeUp mt-6 flex flex-wrap justify-center gap-3 delay-300">
				<Button
					variant="secondary"
					class="rounded-xl border border-amber-200 bg-white/80 font-medium text-amber-700 shadow-sm transition-all hover:scale-105 hover:bg-amber-50 hover:shadow-md"
					onclick={loadAlerts}
				>
					{$t('alerts.refresh')}
				</Button>
				<a href="/operations">
					<Button
						class="rounded-xl bg-gradient-to-r from-amber-500 to-orange-600 text-white shadow-md transition-all hover:scale-105 hover:from-amber-600 hover:to-orange-700 hover:shadow-lg"
					>
						{$t('alerts.go_to_ops')}
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
	title={$t('alerts.details.title')}
	subtitle={detailModalSubtitle}
	buildSections={buildAlertSections}
/>

<!-- MAIN -->
<section class="mx-auto max-w-7xl space-y-8 px-4 py-12 sm:px-6">
	<!-- Filters + Live Alerts -->
	<Card
		class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-amber-50 to-orange-100 shadow-lg transition-all hover:shadow-xl"
	>
		<CardHeader
			class="flex flex-col gap-4 border-b border-white/60 bg-white/75 px-6 py-5 backdrop-blur sm:flex-row sm:items-end sm:justify-between"
		>
			<div>
				<CardTitle class="text-slate-800">{$t('alerts.live_alerts')}</CardTitle>
				<CardDescription class="text-slate-600">{$t('alerts.live_alerts_desc')}</CardDescription>
			</div>
			<div class="flex gap-2">
				<div class="select-wrapper w-[200px]">
					<select
						class="rounded-xl border border-amber-200 bg-white/90 px-3 py-2 text-sm focus:ring-2 focus:ring-amber-400"
						bind:value={filters.type}
					>
						<option value="">{$t('alerts.filters.placeholder_type')}</option>
						<option value="LOW_STOCK">{$t('alerts.filters.type_options.low_stock')}</option>
						<option value="OVERSTOCK">{$t('alerts.filters.type_options.overstock')}</option>
						<option value="OUT_OF_STOCK">{$t('alerts.filters.type_options.out_of_stock')}</option>
						<option value="EXPIRY_ALERT">{$t('alerts.filters.type_options.expiry')}</option>
					</select>
				</div>

				<div class="select-wrapper w-[150px]">
					<select
						class="rounded-xl border border-amber-200 bg-white/90 px-3 py-2 text-sm focus:ring-2 focus:ring-amber-400"
						bind:value={filters.status}
					>
						<option value="">{$t('alerts.filters.placeholder_status')}</option>
						<option value="ACTIVE">{$t('alerts.filters.status_options.active')}</option>
						<option value="RESOLVED">{$t('alerts.filters.status_options.resolved')}</option>
					</select>
				</div>

				<Button
					variant="secondary"
					class="rounded-xl border border-amber-200 bg-white/90 text-amber-700 hover:bg-amber-50"
					onclick={loadAlerts}
				>
					{$t('alerts.filters.refresh_btn')}
				</Button>
			</div>
		</CardHeader>
		<CardContent class="p-0">
			<Table>
				<TableHeader
					class="sticky top-0 z-10 bg-gradient-to-r from-amber-100/85 to-orange-100/85 backdrop-blur"
				>
					<TableRow class="border-y border-amber-200/70">
						<TableHead class="px-4 py-3 text-slate-700">{$t('alerts.table.type')}</TableHead>
						<TableHead class="px-4 py-3 text-slate-700">{$t('alerts.table.product')}</TableHead>
						<TableHead class="px-4 py-3 text-slate-700">{$t('alerts.table.message')}</TableHead>
						<TableHead class="px-4 py-3 text-slate-700">{$t('alerts.table.status')}</TableHead>
						{#if auth.hasPermission('alerts.manage')}
							<TableHead class="px-4 py-3 text-right text-slate-700">{$t('alerts.table.action')}</TableHead>
						{/if}
					</TableRow>
				</TableHeader>
				<TableBody class="[&>tr:nth-child(even)]:bg-white/70 [&>tr:nth-child(odd)]:bg-white/50">
					{#if loading}
						{#each Array(4) as _, i}
							<TableRow class="hover:bg-white/80">
								<TableCell colspan={auth.hasPermission('alerts.manage') ? 5 : 4} class="px-4 py-3">
									<Skeleton class="h-6 w-full bg-white/70" />
								</TableCell>
							</TableRow>
						{/each}
					{:else if alerts.length === 0}
						<TableRow>
							<TableCell
								colspan={auth.hasPermission('alerts.manage') ? 5 : 4}
								class="py-6 text-center text-sm text-slate-600">{$t('alerts.table.empty')}</TableCell
							>
						</TableRow>
					{:else}
						{#each alerts as alert}
							<TableRow
								class="cursor-pointer transition-colors hover:bg-white/90"
								onclick={() => openAlertDetails(alert)}
							>
								<TableCell class="px-4 py-3 text-xs font-semibold text-slate-700"
									>{alert.Type}</TableCell
								>
								<TableCell class="px-4 py-3 text-slate-800"
									>{alert.Product?.Name ?? `${$t('alerts.table.product_fallback')} ${alert.ProductID}`}</TableCell
								>
								<TableCell class="px-4 py-3 text-sm text-slate-600">{alert.Message}</TableCell>
								<TableCell class="px-4 py-3">
									<span
										class={cn(
											'rounded-full border px-2.5 py-1 text-xs capitalize shadow-sm',
											alert.Status === 'ACTIVE'
												? 'border-amber-200 bg-amber-100 text-amber-700'
												: 'border-slate-200 bg-slate-100 text-slate-600'
										)}
									>
										{alert.Status}
									</span>
								</TableCell>
								{#if auth.hasPermission('alerts.manage')}
									<TableCell class="px-4 py-3 text-right">
										{#if alert.Status !== 'RESOLVED'}
											<Button
												size="sm"
												variant="ghost"
												class="rounded-lg px-3 py-1.5 text-amber-700 hover:bg-amber-100"
												onclick={(event) => {
													event.stopPropagation();
													resolveAlert(alert.ID);
												}}
											>
												{$t('alerts.table.resolve')}
											</Button>
										{/if}
									</TableCell>
								{/if}
							</TableRow>
						{/each}
					{/if}
				</TableBody>
			</Table>
		</CardContent>
	</Card>

	<!-- Forms -->
	{#if auth.hasPermission('alerts.manage')}
		<div class="grid gap-8 lg:grid-cols-2">
			<!-- Product thresholds -->
			<Card
				class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-emerald-50 to-green-100 shadow-lg transition-all hover:scale-[1.01] hover:shadow-xl"
			>
				<CardHeader class="border-b border-white/60 bg-white/75 px-6 py-5 backdrop-blur">
					<CardTitle class="text-slate-800">{$t('alerts.thresholds.title')}</CardTitle>
					<CardDescription class="text-slate-600">{$t('alerts.thresholds.subtitle')}</CardDescription>
				</CardHeader>
				<CardContent class="space-y-3 p-6">
					<div class="space-y-1">
						<label class="text-sm font-medium text-slate-700">{$t('alerts.thresholds.product_label')}</label>
						<ProductSelector
							bind:value={productSettingsForm.productId}
							placeholder={$t('alerts.thresholds.product_placeholder')}
							className="w-full"
						/>
					</div>
					<div class="grid grid-cols-1 gap-2 sm:grid-cols-3">
						<Input
							type="number"
							placeholder={$t('alerts.thresholds.low_stock')}
							bind:value={productSettingsForm.lowStockLevel}
							class="rounded-xl border-emerald-200 bg-white/90 focus:ring-2 focus:ring-emerald-400"
						/>
						<Input
							type="number"
							placeholder={$t('alerts.thresholds.overstock')}
							bind:value={productSettingsForm.overStockLevel}
							class="rounded-xl border-emerald-200 bg-white/90 focus:ring-2 focus:ring-emerald-400"
						/>
						<Input
							type="number"
							placeholder={$t('alerts.thresholds.expiry_days')}
							bind:value={productSettingsForm.expiryAlertDays}
							class="rounded-xl border-emerald-200 bg-white/90 focus:ring-2 focus:ring-emerald-400"
						/>
					</div>
					<Button
						class="w-full rounded-xl bg-gradient-to-r from-emerald-500 to-green-600 text-white shadow-md transition-all hover:scale-105 hover:from-emerald-600 hover:to-green-700 hover:shadow-lg"
						onclick={updateProductSettings}
					>
						{$t('alerts.thresholds.save')}
					</Button>
				</CardContent>
			</Card>

			<!-- User notifications -->
			<Card
				class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-violet-50 to-purple-100 shadow-lg transition-all hover:scale-[1.01] hover:shadow-xl"
			>
				<CardHeader class="border-b border-white/60 bg-white/75 px-6 py-5 backdrop-blur">
					<CardTitle class="text-slate-800">{$t('alerts.notifications.title')}</CardTitle>
					<CardDescription class="text-slate-600"
						>{$t('alerts.notifications.subtitle')}</CardDescription
					>
				</CardHeader>
				<CardContent class="space-y-3 p-6">
					<div class="space-y-1">
						<label class="text-sm font-medium text-slate-700">{$t('alerts.notifications.user_label')}</label>
						<UserSelector
							bind:value={userSettingsForm.userId}
							placeholder={$t('alerts.notifications.user_placeholder')}
							className="w-full border-violet-200"
						/>
					</div>
					<Input
						placeholder={$t('alerts.notifications.email_placeholder')}
						bind:value={userSettingsForm.emailAddress}
						class="rounded-xl border-violet-200 bg-white/90 focus:ring-2 focus:ring-violet-400"
					/>
					<Input
						placeholder={$t('alerts.notifications.phone_placeholder')}
						bind:value={userSettingsForm.phoneNumber}
						class="rounded-xl border-violet-200 bg-white/90 focus:ring-2 focus:ring-violet-400"
					/>
					<div class="flex flex-wrap items-center gap-4 text-sm">
						<label class="flex items-center gap-2">
							<input type="checkbox" bind:checked={userSettingsForm.emailNotificationsEnabled} />
							{$t('alerts.notifications.email_label')}
						</label>
						<label class="flex items-center gap-2">
							<input type="checkbox" bind:checked={userSettingsForm.smsNotificationsEnabled} />
							{$t('alerts.notifications.sms_label')}
						</label>
					</div>
					<Button
						class="w-full rounded-xl bg-gradient-to-r from-violet-500 to-purple-600 text-white shadow-md transition-all hover:scale-105 hover:from-violet-600 hover:to-purple-700 hover:shadow-lg"
						onclick={updateUserSettings}
					>
						{$t('alerts.notifications.save')}
					</Button>
				</CardContent>
			</Card>
		</div>
	{/if}
</section>

<style lang="postcss">
	@keyframes gradientShift {
		0% {
			background-position: 0% 50%;
		}
		50% {
			background-position: 100% 50%;
		}
		100% {
			background-position: 0% 50%;
		}
	}
	.animate-gradientShift {
		background-size: 200% 200%;
		animation: gradientShift 22s ease infinite;
	}

	@keyframes heroFloat {
		0%,
		100% {
			transform: translateY(0) scale(1);
		}
		50% {
			transform: translateY(-6px) scale(1.02);
		}
	}
	.animate-heroFloat {
		animation: heroFloat 10s ease-in-out infinite;
	}

	@keyframes pulseGlow {
		0%,
		100% {
			transform: scale(1);
			box-shadow: 0 0 14px rgba(251, 191, 36, 0.28);
		}
		50% {
			transform: scale(1.06);
			box-shadow: 0 0 24px rgba(251, 146, 60, 0.45);
		}
	}
	.animate-pulseGlow {
		animation: pulseGlow 7.5s ease-in-out infinite;
	}

	@keyframes fadeUp {
		from {
			opacity: 0;
			transform: translateY(18px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
	.animate-fadeUp {
		animation: fadeUp 1s ease forwards;
	}

	/* Smooth material-like transitions */
	* {
		transition-property:
			color, background-color, border-color, text-decoration-color, fill, stroke, opacity,
			box-shadow, transform, filter, backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
	}

	/* Thin pastel scrollbar */
	::-webkit-scrollbar {
		width: 8px;
		height: 8px;
	}
	::-webkit-scrollbar-track {
		background: transparent;
	}
	::-webkit-scrollbar-thumb {
		background: rgba(251, 191, 36, 0.28);
		border-radius: 9999px;
	}
	::-webkit-scrollbar-thumb:hover {
		background: rgba(251, 146, 60, 0.4);
	}
</style>
