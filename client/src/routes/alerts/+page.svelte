<script lang="ts">
import { onMount } from 'svelte';
import { toast } from 'svelte-sonner';
import { alertsApi } from '$lib/api/resources';
import type { Alert } from '$lib/types';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
import { Input } from '$lib/components/ui/input';
import { Button } from '$lib/components/ui/button';
import { Skeleton } from '$lib/components/ui/skeleton';
import { cn } from '$lib/utils';

	const filters = $state({ type: '', status: 'ACTIVE' });
	let alerts = $state<Alert[]>([]);
	let loading = $state(false);

	const productSettingsForm = $state({ productId: '', lowStockLevel: '', overStockLevel: '', expiryAlertDays: '' });
	const userSettingsForm = $state({ userId: '', emailNotificationsEnabled: true, smsNotificationsEnabled: false, emailAddress: '', phoneNumber: '' });

	const loadAlerts = async () => {
		loading = true;
		try {
			alerts = await alertsApi.list({ type: filters.type || undefined, status: filters.status || undefined });
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to load alerts';
			toast.error('Failed to Load Alerts', {
				description: errorMessage,
			});
		} finally {
			loading = false;
		}
	};

	onMount(loadAlerts);

	const resolveAlert = async (alertId: number) => {
		try {
			await alertsApi.resolve(alertId);
			toast.success('Alert resolved');
			await loadAlerts();
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to resolve alert';
			toast.error('Failed to Resolve Alert', {
				description: errorMessage,
			});
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
				expiryAlertDays: Number(productSettingsForm.expiryAlertDays) || 0,
			});
			toast.success('Thresholds updated');
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to save thresholds';
			toast.error('Failed to Save Thresholds', {
				description: errorMessage,
			});
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
				phoneNumber: userSettingsForm.phoneNumber,
			});
			toast.success('Notification preferences saved');
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to save preferences';
			toast.error('Failed to Save Preferences', {
				description: errorMessage,
			});
		}
	};
</script>

<section class="space-y-8">
	<header>
		<p class="text-sm uppercase tracking-wide text-muted-foreground">Alerts & notifications</p>
		<h1 class="text-3xl font-semibold">Thresholds, escalations, and user messaging</h1>
	</header>

	<Card>
		<CardHeader class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
			<div>
				<CardTitle>Live alerts</CardTitle>
				<CardDescription>Filter by type or lifecycle state</CardDescription>
			</div>
			<div class="flex flex-wrap gap-2">
				<select class="rounded-md border border-border bg-background px-3 py-2 text-sm" bind:value={filters.type}>
					<option value="">All types</option>
					<option value="LOW_STOCK">Low stock</option>
					<option value="OVERSTOCK">Overstock</option>
					<option value="OUT_OF_STOCK">Out of stock</option>
					<option value="EXPIRY_ALERT">Expiry</option>
				</select>
				<select class="rounded-md border border-border bg-background px-3 py-2 text-sm" bind:value={filters.status}>
					<option value="">Any status</option>
					<option value="ACTIVE">Active</option>
					<option value="RESOLVED">Resolved</option>
				</select>
				<Button variant="secondary" onclick={loadAlerts}>Refresh</Button>
			</div>
		</CardHeader>
		<CardContent>
			<Table>
				<TableHeader>
					<TableRow>
						<TableHead>Type</TableHead>
						<TableHead>Product</TableHead>
						<TableHead>Message</TableHead>
						<TableHead>Status</TableHead>
						<TableHead class="text-right">Action</TableHead>
					</TableRow>
				</TableHeader>
				<TableBody>
					{#if loading}
						{#each Array(4) as _, i}
							<TableRow>
								<TableCell colspan="5"><Skeleton class="h-6 w-full" /></TableCell>
							</TableRow>
						{/each}
					{:else if alerts.length === 0}
						<TableRow>
							<TableCell colspan="5" class="text-center text-sm text-muted-foreground">No alerts found</TableCell>
						</TableRow>
					{:else}
						{#each alerts as alert}
							<TableRow>
								<TableCell class="text-xs font-semibold">{alert.Type}</TableCell>
								<TableCell>{alert.Product?.Name ?? `Product ${alert.ProductID}`}</TableCell>
								<TableCell class="text-sm text-muted-foreground">{alert.Message}</TableCell>
							<TableCell>
								<span
									class={cn(
										'rounded-full px-2 py-0.5 text-xs capitalize',
										alert.Status === 'ACTIVE' ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'
									)}
								>
									{alert.Status}
								</span>
							</TableCell>
								<TableCell class="text-right">
									{#if alert.Status !== 'RESOLVED'}
										<Button size="sm" variant="ghost" onclick={() => resolveAlert(alert.ID)}>Resolve</Button>
									{/if}
								</TableCell>
							</TableRow>
						{/each}
					{/if}
				</TableBody>
			</Table>
		</CardContent>
	</Card>

	<div class="grid gap-6 lg:grid-cols-2">
		<Card>
			<CardHeader>
				<CardTitle>Product thresholds</CardTitle>
				<CardDescription>Configure alerting per SKU</CardDescription>
			</CardHeader>
			<CardContent class="space-y-3">
				<Input type="number" min="1" placeholder="Product ID" bind:value={productSettingsForm.productId} />
				<div class="grid grid-cols-3 gap-2">
					<Input type="number" placeholder="Low stock" bind:value={productSettingsForm.lowStockLevel} />
					<Input type="number" placeholder="Overstock" bind:value={productSettingsForm.overStockLevel} />
					<Input type="number" placeholder="Expiry days" bind:value={productSettingsForm.expiryAlertDays} />
				</div>
				<Button class="w-full" onclick={updateProductSettings}>Save thresholds</Button>
			</CardContent>
		</Card>
		<Card>
			<CardHeader>
				<CardTitle>User notifications</CardTitle>
				<CardDescription>Escalation preferences per operator</CardDescription>
			</CardHeader>
			<CardContent class="space-y-3">
				<Input type="number" min="1" placeholder="User ID" bind:value={userSettingsForm.userId} />
				<Input placeholder="Email" bind:value={userSettingsForm.emailAddress} />
				<Input placeholder="Phone" bind:value={userSettingsForm.phoneNumber} />
				<div class="flex items-center gap-2 text-sm">
					<label class="flex items-center gap-2"><input type="checkbox" bind:checked={userSettingsForm.emailNotificationsEnabled} /> Email</label>
					<label class="flex items-center gap-2"><input type="checkbox" bind:checked={userSettingsForm.smsNotificationsEnabled} /> SMS</label>
				</div>
				<Button class="w-full" onclick={updateUserSettings}>Save preferences</Button>
			</CardContent>
		</Card>
	</div>
</section>
