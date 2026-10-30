<script lang="ts">
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { alertsApi } from '$lib/api/resources';
	import { optimistic } from '$lib/optimistic';
	import type { Alert } from '$lib/types';
	import { t } from '$lib/i18n';
	import DetailsModal from '$lib/components/DetailsModal.svelte';
	import type { DetailBuilderContext, DetailSection } from '$lib/components/DetailsModal.svelte';

	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import EmptyState from '$lib/components/ui/empty-state.svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import ProductSelector from '$lib/components/ui/product-selector.svelte';
	import UserSelector from '$lib/components/ui/user-selector.svelte';

	import {
		Bell,
		Activity,
		CalendarClock,
		Package,
		AlertTriangle,
		Plus,
		Trash2,
		Loader2,
		Users
	} from 'lucide-svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	const inputClass = 'h-11 w-full rounded-md text-sm';

	const filters = $state({ type: '', status: 'ACTIVE' });
	let alerts = $state<Alert[]>([]);
	let loading = $state(false);

	let roleSubscriptions = $state<any[]>([]);
	let roleSubsLoading = $state(false);
	let addSubOpen = $state(false);
	let newSub = $state({ alertType: 'LOW_STOCK', roleName: '' });
	let creatingSub = $state(false);

	const ALERT_TYPES = ['LOW_STOCK', 'OUT_OF_STOCK', 'OVERSTOCK', 'EXPIRY_ALERT'];

	$effect(() => {
		if (!auth.hasPermission('alerts.view')) {
			toast.error($t('alerts.toasts.access_denied'), {
				description: $t('alerts.toasts.access_denied_desc')
			});
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
					{
						label: $t('alerts.details.batch_id'),
						value: alert.BatchID ? `#${alert.BatchID}` : $t('alerts.details.na')
					},
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
		const previousAlerts = alerts.map((a) => ({ ...a }));
		try {
			await optimistic(
				() => {
					alerts = alerts.map((a) =>
						a.ID === alertId ? { ...a, Status: 'RESOLVED', UpdatedAt: new Date().toISOString() } : a
					);
				},
				() => {
					alerts = previousAlerts;
				},
				() => alertsApi.resolve(alertId),
				{ success: $t('alerts.toasts.resolve_success'), error: $t('alerts.toasts.resolve_fail') }
			);
		} catch {
			// optimistic() already rolled back and toasted
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
		if (auth.hasPermission('alerts.manage')) {
			loadRoleSubscriptions();
		}

		const hero = document.querySelector('.parallax-hero') as HTMLElement | null;
		if (!hero) return;
		const handleScroll = () => {
			const scrollY = window.scrollY / 6;
			hero.style.transform = `translateY(${scrollY}px)`;
		};
		window.addEventListener('scroll', handleScroll, { passive: true });
		return () => window.removeEventListener('scroll', handleScroll);
	});

	const loadRoleSubscriptions = async () => {
		roleSubsLoading = true;
		try {
			const data = await alertsApi.getRoleSubscriptions();
			roleSubscriptions = Array.isArray(data) ? data : data?.subscriptions || [];
		} catch (e) {
			toast.error('Failed to load role subscriptions');
		} finally {
			roleSubsLoading = false;
		}
	};

	const createRoleSubscription = async () => {
		if (!newSub.roleName) {
			toast.warning('Role name is required');
			return;
		}
		creatingSub = true;
		try {
			await alertsApi.createRoleSubscription(newSub);
			toast.success('Role subscription created');
			addSubOpen = false;
			newSub = { alertType: 'LOW_STOCK', roleName: '' };
			await loadRoleSubscriptions();
		} catch (e: any) {
			toast.error('Failed to create subscription', {
				description: e?.response?.data?.error || 'Unknown error'
			});
		} finally {
			creatingSub = false;
		}
	};

	const deleteRoleSubscription = async (id: number) => {
		try {
			await alertsApi.deleteRoleSubscription(id);
			toast.success('Role subscription deleted');
			await loadRoleSubscriptions();
		} catch (e: any) {
			toast.error('Failed to delete subscription', {
				description: e?.response?.data?.error || 'Unknown error'
			});
		}
	};
</script>

<div class="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8 lg:py-10">
	<!-- Page header -->
	<header class="page-header" data-animate="fade-up">
		<div class="parallax-hero transition-transform duration-700 ease-out will-change-transform">
			<p class="flex items-center gap-2 overline">
				<span class="inline-flex rounded-md bg-primary/10 p-1.5 text-primary">
					<Bell class="h-3.5 w-3.5" />
				</span>
				{$t('alerts.subtitle')}
			</p>
			<h1 class="page-title mt-1">{$t('alerts.title')}</h1>
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<Button variant="outline" onclick={loadAlerts}>
				{$t('alerts.refresh')}
			</Button>
			<Button href="/operations" class="btn-soothe">
				{$t('alerts.go_to_ops')}
			</Button>
		</div>
	</header>

	<DetailsModal
		bind:open={detailModalOpen}
		resourceId={detailAlertId}
		endpoint="/alerts"
		title={$t('alerts.details.title')}
		subtitle={detailModalSubtitle}
		buildSections={buildAlertSections}
	/>

	<!-- MAIN -->
	<div class="mt-6 space-y-4">
		<!-- Filters + Live Alerts -->
		<section class="panel overflow-hidden">
			<div class="panel-header flex-wrap items-end">
				<div>
					<h2 class="section-title">{$t('alerts.live_alerts')}</h2>
					<p class="text-sm text-muted-foreground">{$t('alerts.live_alerts_desc')}</p>
				</div>
				<div class="flex flex-wrap items-center gap-2">
					<select
						bind:value={filters.type}
						aria-label={$t('alerts.filters.placeholder_type')}
						class="h-10 w-[11rem] text-sm"
					>
						<option value="">{$t('alerts.filters.placeholder_type')}</option>
						<option value="LOW_STOCK">{$t('alerts.filters.type_options.low_stock')}</option>
						<option value="OVERSTOCK">{$t('alerts.filters.type_options.overstock')}</option>
						<option value="OUT_OF_STOCK">{$t('alerts.filters.type_options.out_of_stock')}</option>
						<option value="EXPIRY_ALERT">{$t('alerts.filters.type_options.expiry')}</option>
					</select>

					<select
						bind:value={filters.status}
						aria-label={$t('alerts.filters.placeholder_status')}
						class="h-10 w-[9.5rem] text-sm"
					>
						<option value="">{$t('alerts.filters.placeholder_status')}</option>
						<option value="ACTIVE">{$t('alerts.filters.status_options.active')}</option>
						<option value="RESOLVED">{$t('alerts.filters.status_options.resolved')}</option>
					</select>

					<Button variant="secondary" onclick={loadAlerts}>
						{$t('alerts.filters.refresh_btn')}
					</Button>
				</div>
			</div>
			<div class="overflow-x-auto">
				<table class="ledger-table">
					<thead>
						<tr>
							<th>{$t('alerts.table.type')}</th>
							<th>{$t('alerts.table.product')}</th>
							<th>{$t('alerts.table.message')}</th>
							<th>{$t('alerts.table.status')}</th>
							{#if auth.hasPermission('alerts.manage')}
								<th class="text-right">{$t('alerts.table.action')}</th>
							{/if}
						</tr>
					</thead>
					<tbody>
						{#if loading}
							{#each Array(4) as _, i}
								<tr>
									<td colspan={auth.hasPermission('alerts.manage') ? 5 : 4}>
										<Skeleton class="h-6 w-full bg-muted" />
									</td>
								</tr>
							{/each}
						{:else if alerts.length === 0}
							<tr>
								<td colspan={auth.hasPermission('alerts.manage') ? 5 : 4} class="p-0">
									<EmptyState
										icon={Bell}
										title="No alerts right now"
										description="Stock, expiry and overstock alerts will appear here as they trigger."
									/>
								</td>
							</tr>
						{:else}
							{#each alerts as alert}
								<tr class="cursor-pointer" onclick={() => openAlertDetails(alert)}>
									<td>
										<span
											class="{alert.Type === 'OUT_OF_STOCK'
												? 'badge-danger'
												: alert.Type === 'OVERSTOCK'
													? 'badge-info'
													: 'badge-warning'} capitalize"
										>
											{alert.Type}
										</span>
									</td>
									<td class="font-medium text-foreground"
										>{alert.Product?.Name ??
											`${$t('alerts.table.product_fallback')} ${alert.ProductID}`}</td
									>
									<td class="text-sm text-muted-foreground">{alert.Message}</td>
									<td>
										<span
											class="{alert.Status === 'ACTIVE'
												? 'badge-warning'
												: 'badge-neutral'} capitalize"
										>
											{alert.Status}
										</span>
									</td>
									{#if auth.hasPermission('alerts.manage')}
										<td class="text-right">
											{#if alert.Status !== 'RESOLVED'}
												<Button
													size="sm"
													variant="ghost"
													class="h-8 px-2.5 text-xs font-semibold text-primary hover:bg-primary/10"
													onclick={(event) => {
														event.stopPropagation();
														resolveAlert(alert.ID);
													}}
												>
													{$t('alerts.table.resolve')}
												</Button>
											{/if}
										</td>
									{/if}
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>
		</section>

		<!-- Forms -->
		{#if auth.hasPermission('alerts.manage')}
			<div class="grid gap-4 lg:grid-cols-2">
				<!-- Product thresholds -->
				<section class="panel overflow-hidden">
					<div class="panel-header items-start">
						<div class="space-y-0.5">
							<h2 class="section-title">{$t('alerts.thresholds.title')}</h2>
							<p class="text-sm text-muted-foreground">{$t('alerts.thresholds.subtitle')}</p>
						</div>
					</div>
					<div class="panel-body space-y-3">
						<div class="space-y-1.5">
							<label class="text-sm font-medium text-foreground"
								>{$t('alerts.thresholds.product_label')}</label
							>
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
								class="{inputClass} num-mono"
							/>
							<Input
								type="number"
								placeholder={$t('alerts.thresholds.overstock')}
								bind:value={productSettingsForm.overStockLevel}
								class="{inputClass} num-mono"
							/>
							<Input
								type="number"
								placeholder={$t('alerts.thresholds.expiry_days')}
								bind:value={productSettingsForm.expiryAlertDays}
								class="{inputClass} num-mono"
							/>
						</div>
						<Button class="btn-soothe w-full" onclick={updateProductSettings}>
							{$t('alerts.thresholds.save')}
						</Button>
					</div>
				</section>

				<!-- User notifications -->
				<section class="panel overflow-hidden">
					<div class="panel-header items-start">
						<div class="space-y-0.5">
							<h2 class="section-title">{$t('alerts.notifications.title')}</h2>
							<p class="text-sm text-muted-foreground">{$t('alerts.notifications.subtitle')}</p>
						</div>
					</div>
					<div class="panel-body space-y-3">
						<div class="space-y-1.5">
							<label class="text-sm font-medium text-foreground"
								>{$t('alerts.notifications.user_label')}</label
							>
							<UserSelector
								bind:value={userSettingsForm.userId}
								placeholder={$t('alerts.notifications.user_placeholder')}
								className="w-full"
							/>
						</div>
						<Input
							placeholder={$t('alerts.notifications.email_placeholder')}
							bind:value={userSettingsForm.emailAddress}
							class={inputClass}
						/>
						<Input
							placeholder={$t('alerts.notifications.phone_placeholder')}
							bind:value={userSettingsForm.phoneNumber}
							class={inputClass}
						/>
						<div class="flex flex-wrap items-center gap-5 text-sm text-foreground">
							<label class="flex cursor-pointer items-center gap-2">
								<input
									type="checkbox"
									class="size-4 rounded-sm accent-primary"
									bind:checked={userSettingsForm.emailNotificationsEnabled}
								/>
								{$t('alerts.notifications.email_label')}
							</label>
							<label class="flex cursor-pointer items-center gap-2">
								<input
									type="checkbox"
									class="size-4 rounded-sm accent-primary"
									bind:checked={userSettingsForm.smsNotificationsEnabled}
								/>
								{$t('alerts.notifications.sms_label')}
							</label>
						</div>
						<Button class="btn-soothe w-full" onclick={updateUserSettings}>
							{$t('alerts.notifications.save')}
						</Button>
					</div>
				</section>
			</div>

			<section class="panel overflow-hidden">
				<div class="panel-header flex-wrap items-start">
					<div class="flex items-start gap-3">
						<span
							class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
						>
							<Users class="h-4 w-4" />
						</span>
						<div class="space-y-0.5">
							<h2 class="section-title">Alert Role Subscriptions</h2>
							<p class="text-sm text-muted-foreground">
								Configure which roles receive which alert types.
							</p>
						</div>
					</div>
					<Button class="btn-soothe h-10 px-4" onclick={() => (addSubOpen = true)}>
						<Plus class="mr-1.5 h-4 w-4" />
						Add Subscription
					</Button>
				</div>
				<div class="overflow-x-auto">
					<table class="ledger-table">
						<thead>
							<tr>
								<th>Alert Type</th>
								<th>Role</th>
								<th class="text-right">Actions</th>
							</tr>
						</thead>
						<tbody>
							{#if roleSubsLoading}
								{#each Array(3) as _}
									<tr>
										<td colspan={3}>
											<Skeleton class="h-6 w-full bg-muted" />
										</td>
									</tr>
								{/each}
							{:else if roleSubscriptions.length === 0}
								<tr>
									<td colspan={3} class="py-8 text-center text-sm text-muted-foreground">
										No role subscriptions configured.
									</td>
								</tr>
							{:else}
								{#each roleSubscriptions as sub}
									<tr>
										<td>
											<span class="badge-info">
												{sub.AlertType || sub.alertType}
											</span>
										</td>
										<td class="font-medium text-foreground">
											{sub.RoleName || sub.roleName || sub.Role?.Name || '—'}
										</td>
										<td class="text-right">
											<Button
												size="sm"
												variant="ghost"
												class="h-8 w-8 p-0 text-danger hover:bg-danger/10 hover:text-danger"
												aria-label="Delete subscription"
												onclick={() => deleteRoleSubscription(sub.ID || sub.id)}
											>
												<Trash2 class="h-4 w-4" />
											</Button>
										</td>
									</tr>
								{/each}
							{/if}
						</tbody>
					</table>
				</div>
			</section>
		{/if}
	</div>
</div>

<Dialog.Root bind:open={addSubOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title class="flex items-center gap-2">
				<Users class="h-5 w-5 text-primary" />
				Add Role Subscription
			</Dialog.Title>
			<Dialog.Description>Assign an alert type to a role.</Dialog.Description>
		</Dialog.Header>
		<div class="space-y-4 py-4">
			<div class="space-y-2">
				<label class="text-sm font-medium text-foreground" for="sub-alert-type">Alert Type</label>
				<select id="sub-alert-type" bind:value={newSub.alertType} class="h-11 text-sm">
					{#each ALERT_TYPES as type}
						<option value={type}>{type}</option>
					{/each}
				</select>
			</div>
			<div class="space-y-2">
				<label class="text-sm font-medium text-foreground" for="sub-role-name">Role Name</label>
				<Input
					id="sub-role-name"
					bind:value={newSub.roleName}
					placeholder="e.g. Admin, Manager, Staff"
					class={inputClass}
				/>
			</div>
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (addSubOpen = false)}>Cancel</Button>
			<Button class="btn-soothe" onclick={createRoleSubscription} disabled={creatingSub}>
				{#if creatingSub}
					<Loader2 class="mr-2 h-4 w-4 animate-spin" />
				{/if}
				Create
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
