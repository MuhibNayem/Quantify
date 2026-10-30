<script lang="ts">
	import { onMount } from 'svelte';
	import { settingsApi } from '$lib/api/settings';
	import { notificationsApi } from '$lib/api/resources';
	import { optimistic } from '$lib/optimistic';
	import { toast } from 'svelte-sonner';
	import * as Tabs from '$lib/components/ui/tabs';
	import GlassTabs from '$lib/components/ui/glass-tabs.svelte';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Switch } from '$lib/components/ui/switch';
	import { Select } from '$lib/components/ui/select';
	import HelpTip from '$lib/components/ui/help-tip.svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import {
		Settings,
		Shield,
		FileText,
		Bell,
		Save,
		Building2,
		Globe,
		Clock,
		Lock,
		LayoutTemplate,
		ShieldCheck,
		Sparkles,
		Coins,
		Percent,
		Zap,
		Megaphone,
		Send,
		Loader2
	} from 'lucide-svelte';
	import RoleManager from '$lib/components/settings/RoleManager.svelte';
	import { fade, fly } from 'svelte/transition';
	import { cn } from '$lib/utils';
	import { t } from '$lib/i18n';

	import { settings as systemSettings } from '$lib/stores/settings';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	const inputClass = 'h-11 w-full rounded-md text-sm';
	const selectClass = 'h-11 w-full';
	const tabsClass = 'mx-auto flex-nowrap';

	$effect(() => {
		if (!auth.hasPermission('settings.view')) {
			toast.error('Access Denied', { description: 'You do not have permission to view settings.' });
			goto('/');
		}
	});

	let settings: any = $state({
		business_name: '',
		currency_symbol: '$',
		currency_code: 'USD',
		timezone: 'UTC',
		locale: 'en-US',
		privacy_policy: '',
		terms_of_service: '',
		loyalty_points_earning_rate: '1',
		loyalty_points_redemption_rate: '0.01',
		loyalty_tier_silver: '500',
		loyalty_tier_gold: '2500',
		loyalty_tier_platinum: '10000',
		tax_rate_percentage: '0'
	});
	let activeTab = $state('general');

	const ALERT_TYPES = [
		{ value: 'LOW_STOCK', label: 'Low Stock' },
		{ value: 'OUT_OF_STOCK', label: 'Out of Stock' },
		{ value: 'OVERSTOCK', label: 'Overstock' },
		{ value: 'EXPIRY_ALERT', label: 'Expiry Alert' }
	];

	let subscriptions = $state<Record<string, boolean>>({});
	let subscriptionsLoading = $state(false);

	let broadcastOpen = $state(false);
	let broadcastForm = $state({ title: '', message: '', type: 'SYSTEM' });
	let broadcastSending = $state(false);

	const loadSubscriptions = async () => {
		if (!$auth.user) return;
		subscriptionsLoading = true;
		try {
			const data = await notificationsApi.getSubscriptions($auth.user.ID);
			const subs: Record<string, boolean> = {};
			if (Array.isArray(data)) {
				data.forEach((s: any) => {
					subs[s.Type || s.type] = s.Enabled ?? s.enabled ?? true;
				});
			} else if (data && typeof data === 'object') {
				Object.entries(data).forEach(([k, v]) => {
					subs[k] = Boolean(v);
				});
			}
			ALERT_TYPES.forEach((t) => {
				if (!(t.value in subs)) subs[t.value] = true;
			});
			subscriptions = subs;
		} catch (e) {
			toast.error('Failed to load subscriptions');
		} finally {
			subscriptionsLoading = false;
		}
	};

	const toggleSubscription = async (type: string, enabled: boolean) => {
		if (!$auth.user) return;
		const userId = $auth.user.ID;
		const previousValue = subscriptions[type] ?? true;
		try {
			await optimistic(
				() => {
					subscriptions[type] = enabled;
				},
				() => {
					subscriptions[type] = previousValue;
				},
				() => notificationsApi.updateSubscription(userId, type, enabled),
				{ success: 'Subscription updated', error: 'Failed to update subscription' }
			);
		} catch {
			// optimistic() already rolled back and toasted
		}
	};

	const sendBroadcast = async () => {
		if (!broadcastForm.title || !broadcastForm.message) {
			toast.warning('Title and message are required');
			return;
		}
		broadcastSending = true;
		try {
			await notificationsApi.broadcast(broadcastForm);
			toast.success('Broadcast sent to all users');
			broadcastOpen = false;
			broadcastForm = { title: '', message: '', type: 'SYSTEM' };
		} catch (e: any) {
			toast.error('Broadcast failed', { description: e?.response?.data?.error || 'Unknown error' });
		} finally {
			broadcastSending = false;
		}
	};

	const currencyOptions = [
		{ value: 'USD', label: 'USD - United States Dollar', symbol: '$' },
		{ value: 'EUR', label: 'EUR - Euro', symbol: '€' },
		{ value: 'GBP', label: 'GBP - British Pound', symbol: '£' },
		{ value: 'BDT', label: 'BDT - Bangladeshi Taka', symbol: '৳' },
		{ value: 'INR', label: 'INR - Indian Rupee', symbol: '₹' },
		{ value: 'PKR', label: 'PKR - Pakistani Rupee', symbol: '₨' },
		{ value: 'JPY', label: 'JPY - Japanese Yen', symbol: '¥' },
		{ value: 'CAD', label: 'CAD - Canadian Dollar', symbol: 'C$' },
		{ value: 'AUD', label: 'AUD - Australian Dollar', symbol: 'A$' }
	];

	const timezoneOptions = [
		{ value: 'UTC', label: 'UTC - Coordinated Universal Time' },
		{ value: 'Asia/Dhaka', label: 'Asia/Dhaka (GMT+6)' },
		{ value: 'Asia/Kolkata', label: 'Asia/Kolkata (GMT+5:30)' },
		{ value: 'Asia/Karachi', label: 'Asia/Karachi (GMT+5)' },
		{ value: 'America/New_York', label: 'America/New_York (EST/EDT)' },
		{ value: 'Europe/London', label: 'Europe/London (GMT/BST)' },
		{ value: 'Europe/Paris', label: 'Europe/Paris (CET/CEST)' },
		{ value: 'Asia/Tokyo', label: 'Asia/Tokyo (JST)' },
		{ value: 'Asia/Shanghai', label: 'Asia/Shanghai (CST)' },
		{ value: 'Australia/Sydney', label: 'Australia/Sydney (AEST/AEDT)' },
		{ value: 'America/Los_Angeles', label: 'America/Los_Angeles (PST/PDT)' }
	];

	const localeOptions = [
		{ value: 'en-US', label: 'English (United States)' },
		{ value: 'bn-BD', label: 'Bengali (Bangladesh)' },
		{ value: 'en-GB', label: 'English (United Kingdom)' },
		{ value: 'hi-IN', label: 'Hindi (India)' },
		{ value: 'ja-JP', label: 'Japanese (Japan)' },
		{ value: 'fr-FR', label: 'French (France)' }
	];

	onMount(async () => {
		try {
			const settingsData = await settingsApi.getSettings();
			if (settingsData) {
				Object.values(settingsData)
					.flat()
					.forEach((s: any) => {
						settings[s.Key] = s.Value;
					});
				if (!settings['currency_symbol']) settings['currency_symbol'] = '$';
				if (!settings['currency_code']) settings['currency_code'] = 'USD';
				if (!settings['timezone']) settings['timezone'] = 'UTC';
				if (!settings['locale']) settings['locale'] = 'en-US';
				if (!settings['ai_wake_up_time']) settings['ai_wake_up_time'] = '07:00';
			}
		} catch (e) {
			console.error('Error loading settings', e);
			toast.error('Failed to load settings');
		}
		loadSubscriptions();
	});

	async function saveSetting(key: string, value: string) {
		try {
			await settingsApi.updateSetting(key, value);

			// Update global store to reflect changes immediately (e.g. locale change)
			systemSettings.update((s) => ({ ...s, [key]: value }));

			toast.success($t('common.saved_successfully'));
		} catch (e) {
			toast.error($t('common.failed_to_save'));
		}
	}

	const allTabs = [
		{ id: 'general', label: 'settings.tabs.general', icon: Settings, permission: 'settings.view' },
		{
			id: 'business',
			label: 'settings.tabs.business_rules',
			icon: Coins,
			permission: 'settings.view'
		},
		{ id: 'system', label: 'settings.tabs.system_ai', icon: Zap, permission: 'settings.view' },
		{
			id: 'security',
			label: 'settings.tabs.security_roles',
			icon: ShieldCheck,
			permission: 'roles.view'
		},
		{
			id: 'policies',
			label: 'settings.tabs.policies',
			icon: FileText,
			permission: 'settings.view'
		},
		{
			id: 'notifications',
			label: 'settings.tabs.notifications',
			icon: Bell,
			permission: 'settings.view'
		}
	];

	let tabs = $derived(allTabs.filter((t) => !t.permission || auth.hasPermission(t.permission)));

	// Auto-select first available tab if current one becomes hidden
	$effect(() => {
		if (tabs.length > 0 && !tabs.find((t) => t.id === activeTab)) {
			activeTab = tabs[0].id;
		}
	});
</script>

<div class="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8 lg:py-10">
	<!-- Page header -->
	<header class="page-header" data-animate="fade-up">
		<div>
			<p class="overline">System · Preferences</p>
			<h1 class="page-title mt-1">{$t('common.configuration')}</h1>
			<p class="page-sub">{$t('common.manage_preferences')}</p>
		</div>
	</header>

	<Tabs.Root
		value={activeTab}
		class="mt-6 w-full space-y-6"
		onValueChange={(val) => (activeTab = val)}
	>
		<!-- Ledger Tabs -->
		<div class="-mx-1 overflow-x-auto px-1 pb-1">
			<GlassTabs
				bind:value={activeTab}
				tabs={tabs.map((tab) => ({ value: tab.id, label: $t(tab.label) }))}
				class={tabsClass}
			/>
		</div>

		<!-- General Settings -->
		<Tabs.Content value="general" class="space-y-4 pt-1 outline-none">
			<div in:fly={{ y: 16, duration: 300 }} class="grid gap-4">
				<!-- Business Profile Card -->
				<section class="panel overflow-hidden">
					<div class="panel-header items-start">
						<div class="flex items-start gap-3.5">
							<span
								class="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
							>
								<Building2 size={18} />
							</span>
							<div class="space-y-0.5">
								<h2 class="section-title">{$t('settings.general.business_profile')}</h2>
								<p class="text-sm text-muted-foreground">
									{$t('settings.general.business_profile_desc')}
								</p>
							</div>
						</div>
					</div>

					<div class="panel-body">
						<div class="max-w-xl space-y-2">
							<Label class="ml-1" for="business-name">{$t('settings.general.business_name')}</Label>
							<div class="flex gap-3">
								<Input
									id="business-name"
									class={inputClass}
									bind:value={settings['business_name']}
									placeholder="e.g. Acme Corp"
								/>
								<Button
									class="btn-soothe h-11 shrink-0 px-5"
									onclick={() => saveSetting('business_name', settings['business_name'])}
								>
									<Save size={16} />
									{$t('common.save')}
								</Button>
							</div>
						</div>
					</div>
				</section>

				<!-- Localization -->
				<div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
					<section class="panel overflow-hidden">
						<div class="panel-header items-start">
							<div class="flex items-start gap-3">
								<span
									class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground ring-1 ring-border"
								>
									<Globe size={16} />
								</span>
								<div class="space-y-0.5">
									<h2 class="section-title">{$t('settings.general.currency')}</h2>
									<p class="text-xs text-muted-foreground">Symbol updates with the code.</p>
								</div>
							</div>
						</div>
						<div class="panel-body">
							<div class="flex items-end gap-3">
								<div class="flex-1">
									<Select
										options={currencyOptions}
										bind:value={settings['currency_code']}
										placeholder={$t('settings.general.select_currency')}
										style={selectClass}
									/>
								</div>
								<Button
									size="icon"
									class="btn-soothe h-11 w-11 shrink-0"
									aria-label={$t('common.save')}
									onclick={() => {
										const selected = currencyOptions.find(
											(c) => c.value === settings['currency_code']
										);
										if (selected) {
											saveSetting('currency_code', selected.value);
											saveSetting('currency_symbol', selected.symbol);
											settings['currency_symbol'] = selected.symbol; // Update local state
										}
									}}
								>
									<Save size={16} />
								</Button>
							</div>
						</div>
					</section>

					<section class="panel overflow-hidden">
						<div class="panel-header items-start">
							<div class="flex items-start gap-3">
								<span
									class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground ring-1 ring-border"
								>
									<Clock size={16} />
								</span>
								<div class="space-y-0.5">
									<h2 class="section-title">{$t('settings.general.timezone')}</h2>
									<p class="text-xs text-muted-foreground">Used for all timestamps.</p>
								</div>
							</div>
						</div>
						<div class="panel-body">
							<div class="flex items-end gap-3">
								<div class="flex-1">
									<Select
										options={timezoneOptions}
										bind:value={settings['timezone']}
										placeholder={$t('settings.general.select_timezone')}
										style={selectClass}
									/>
								</div>
								<Button
									size="icon"
									class="btn-soothe h-11 w-11 shrink-0"
									aria-label={$t('common.save')}
									onclick={() => saveSetting('timezone', settings['timezone'])}
								>
									<Save size={16} />
								</Button>
							</div>
						</div>
					</section>

					<section class="panel overflow-hidden">
						<div class="panel-header items-start">
							<div class="flex items-start gap-3">
								<span
									class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground ring-1 ring-border"
								>
									<Globe size={16} />
								</span>
								<div class="space-y-0.5">
									<h2 class="section-title">{$t('settings.general.locale')}</h2>
									<p class="text-xs text-muted-foreground">Interface language.</p>
								</div>
							</div>
						</div>
						<div class="panel-body">
							<div class="flex items-end gap-3">
								<div class="flex-1">
									<Select
										options={localeOptions}
										bind:value={settings['locale']}
										placeholder={$t('settings.general.select_locale')}
										style={selectClass}
									/>
								</div>
								<Button
									size="icon"
									class="btn-soothe h-11 w-11 shrink-0"
									aria-label={$t('common.save')}
									onclick={() => saveSetting('locale', settings['locale'])}
								>
									<Save size={16} />
								</Button>
							</div>
						</div>
					</section>
				</div>
			</div>
		</Tabs.Content>

		<!-- Business Rules -->
		<Tabs.Content value="business" class="space-y-4 pt-1 outline-none">
			<div in:fly={{ y: 16, duration: 300 }} class="grid gap-4">
				<!-- Loyalty Program -->
				<section class="panel overflow-hidden">
					<div class="panel-header items-start">
						<div class="flex items-start gap-3.5">
							<span
								class="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
							>
								<Sparkles size={18} />
							</span>
							<div class="space-y-0.5">
								<h2 class="section-title">{$t('settings.business.loyalty_program')}</h2>
								<p class="text-sm text-muted-foreground">
									{$t('settings.business.loyalty_program_desc')}
								</p>
							</div>
						</div>
					</div>

					<div class="panel-body">
						<div class="grid gap-6 md:grid-cols-2">
							<!-- Earning & Redemption -->
							<div class="space-y-5">
								<h3 class="overline">{$t('settings.business.points_configuration')}</h3>
								<div class="space-y-2">
									<Label class="ml-1 flex items-center gap-1.5">
										{$t('settings.business.earning_rate')}
										<HelpTip
											text="Loyalty points earned per unit of currency spent. For example, 1 means a customer earns 1 point for every $1 spent."
										/>
									</Label>
									<div class="flex gap-3">
										<Input
											type="number"
											step="0.1"
											class="{inputClass} num-mono"
											bind:value={settings['loyalty_points_earning_rate']}
										/>
										<Button
											size="icon"
											class="btn-soothe h-11 w-11 shrink-0"
											aria-label={$t('common.save')}
											onclick={() =>
												saveSetting(
													'loyalty_points_earning_rate',
													settings['loyalty_points_earning_rate']
												)}
										>
											<Save size={16} />
										</Button>
									</div>
									<p class="text-xs text-muted-foreground">
										{$t('settings.business.earning_rate_hint')}
									</p>
								</div>

								<div class="space-y-2">
									<Label class="ml-1">{$t('settings.business.redemption_value')}</Label>
									<div class="flex gap-3">
										<Input
											type="number"
											step="0.01"
											class="{inputClass} num-mono"
											bind:value={settings['loyalty_points_redemption_rate']}
										/>
										<Button
											size="icon"
											class="btn-soothe h-11 w-11 shrink-0"
											aria-label={$t('common.save')}
											onclick={() =>
												saveSetting(
													'loyalty_points_redemption_rate',
													settings['loyalty_points_redemption_rate']
												)}
										>
											<Save size={16} />
										</Button>
									</div>
									<p class="text-xs text-muted-foreground">
										{$t('settings.business.redemption_value_hint')}
									</p>
								</div>
							</div>

							<!-- Tiers -->
							<div class="space-y-4 md:border-l md:border-border/70 md:pl-6">
								<h3 class="overline">{$t('settings.business.tier_thresholds')}</h3>
								<div class="space-y-3">
									<div class="space-y-1.5">
										<Label class="ml-1">{$t('settings.business.silver_tier')}</Label>
										<div class="flex gap-3">
											<Input
												type="number"
												class="{inputClass} num-mono"
												bind:value={settings['loyalty_tier_silver']}
											/>
											<Button
												size="icon"
												class="btn-soothe h-11 w-11 shrink-0"
												aria-label={$t('common.save')}
												onclick={() =>
													saveSetting('loyalty_tier_silver', settings['loyalty_tier_silver'])}
											>
												<Save size={16} />
											</Button>
										</div>
									</div>
									<div class="space-y-1.5">
										<Label class="ml-1">{$t('settings.business.gold_tier')}</Label>
										<div class="flex gap-3">
											<Input
												type="number"
												class="{inputClass} num-mono"
												bind:value={settings['loyalty_tier_gold']}
											/>
											<Button
												size="icon"
												class="btn-soothe h-11 w-11 shrink-0"
												aria-label={$t('common.save')}
												onclick={() =>
													saveSetting('loyalty_tier_gold', settings['loyalty_tier_gold'])}
											>
												<Save size={16} />
											</Button>
										</div>
									</div>
									<div class="space-y-1.5">
										<Label class="ml-1">{$t('settings.business.platinum_tier')}</Label>
										<div class="flex gap-3">
											<Input
												type="number"
												class="{inputClass} num-mono"
												bind:value={settings['loyalty_tier_platinum']}
											/>
											<Button
												size="icon"
												class="btn-soothe h-11 w-11 shrink-0"
												aria-label={$t('common.save')}
												onclick={() =>
													saveSetting('loyalty_tier_platinum', settings['loyalty_tier_platinum'])}
											>
												<Save size={16} />
											</Button>
										</div>
									</div>
								</div>
							</div>
						</div>
					</div>
				</section>

				<!-- Tax Settings -->
				<section class="panel overflow-hidden">
					<div class="panel-header items-start">
						<div class="flex items-start gap-3.5">
							<span
								class="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
							>
								<Percent size={18} />
							</span>
							<div class="space-y-0.5">
								<h2 class="section-title">{$t('settings.business.financial_settings')}</h2>
								<p class="text-sm text-muted-foreground">
									{$t('settings.business.financial_settings_desc')}
								</p>
							</div>
						</div>
					</div>

					<div class="panel-body">
						<div class="max-w-xl space-y-2">
							<Label class="ml-1 flex items-center gap-1.5">
								{$t('settings.business.default_tax_rate')}
								<HelpTip
									text="Default tax percentage applied to sales at checkout. Enter 5 for 5%."
								/>
							</Label>
							<div class="flex gap-3">
								<Input
									type="number"
									step="0.01"
									class="{inputClass} num-mono"
									bind:value={settings['tax_rate_percentage']}
								/>
								<Button
									class="btn-soothe h-11 shrink-0 px-5"
									onclick={() =>
										saveSetting('tax_rate_percentage', settings['tax_rate_percentage'])}
								>
									<Save size={16} />
									{$t('common.save')}
								</Button>
							</div>
							<p class="text-xs text-muted-foreground">
								{$t('settings.business.default_tax_rate_hint')}
							</p>
						</div>
					</div>
				</section>
			</div>
		</Tabs.Content>

		<!-- Security / RBAC -->
		<Tabs.Content value="security" class="pt-1 outline-none">
			<div in:fly={{ y: 16, duration: 300 }}>
				<RoleManager />
			</div>
		</Tabs.Content>

		<!-- Policies -->
		<Tabs.Content value="policies" class="space-y-4 pt-1 outline-none">
			<div in:fly={{ y: 16, duration: 300 }} class="grid gap-4">
				<section class="panel overflow-hidden">
					<div class="panel-header items-start">
						<div class="flex items-start gap-3">
							<span
								class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
							>
								<Lock size={16} />
							</span>
							<div class="space-y-0.5">
								<h2 class="section-title">{$t('settings.policies.privacy_policy')}</h2>
								<p class="text-sm text-muted-foreground">Published on the customer storefront.</p>
							</div>
						</div>
					</div>
					<div class="panel-body space-y-4">
						<Textarea
							class="min-h-[200px] rounded-md text-sm leading-relaxed"
							placeholder={$t('settings.policies.privacy_policy_placeholder')}
							bind:value={settings['privacy_policy']}
						/>
						<div class="flex justify-end">
							<Button
								class="btn-soothe h-10 px-4"
								onclick={() => saveSetting('privacy_policy', settings['privacy_policy'])}
							>
								<Save size={16} />
								{$t('settings.policies.save_policy')}
							</Button>
						</div>
					</div>
				</section>

				<section class="panel overflow-hidden">
					<div class="panel-header items-start">
						<div class="flex items-start gap-3">
							<span
								class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
							>
								<LayoutTemplate size={16} />
							</span>
							<div class="space-y-0.5">
								<h2 class="section-title">{$t('settings.policies.terms_of_service')}</h2>
								<p class="text-sm text-muted-foreground">Shown at checkout and on invoices.</p>
							</div>
						</div>
					</div>
					<div class="panel-body space-y-4">
						<Textarea
							class="min-h-[200px] rounded-md text-sm leading-relaxed"
							placeholder={$t('settings.policies.terms_of_service_placeholder')}
							bind:value={settings['terms_of_service']}
						/>
						<div class="flex justify-end">
							<Button
								class="btn-soothe h-10 px-4"
								onclick={() => saveSetting('terms_of_service', settings['terms_of_service'])}
							>
								<Save size={16} />
								{$t('settings.policies.save_terms')}
							</Button>
						</div>
					</div>
				</section>

				<!-- Return Policy Settings -->
				<section class="panel overflow-hidden">
					<div class="panel-header items-start">
						<div class="flex items-start gap-3">
							<span
								class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
							>
								<Clock size={16} />
							</span>
							<div class="space-y-0.5">
								<h2 class="section-title">{$t('settings.policies.return_policy')}</h2>
								<p class="text-sm text-muted-foreground">Applies to all new return requests.</p>
							</div>
						</div>
					</div>
					<div class="panel-body">
						<div class="max-w-xl space-y-2">
							<Label class="ml-1">{$t('settings.policies.return_window')}</Label>
							<div class="flex gap-3">
								<Input
									type="number"
									class="{inputClass} num-mono"
									bind:value={settings['return_window_days']}
									placeholder="e.g. 30"
								/>
								<Button
									class="btn-soothe h-11 shrink-0 px-5"
									onclick={() => saveSetting('return_window_days', settings['return_window_days'])}
								>
									<Save size={16} />
									{$t('common.save')}
								</Button>
							</div>
							<p class="text-xs text-muted-foreground">
								{$t('settings.policies.return_window_hint')}
							</p>
						</div>
					</div>
				</section>
			</div>
		</Tabs.Content>

		<!-- System & AI -->
		<Tabs.Content value="system" class="space-y-4 pt-1 outline-none">
			<div in:fly={{ y: 16, duration: 300 }} class="grid gap-4">
				<section class="panel overflow-hidden">
					<div class="panel-header items-start">
						<div class="flex items-start gap-3.5">
							<span
								class="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
							>
								<Zap size={18} />
							</span>
							<div class="space-y-0.5">
								<h2 class="section-title">{$t('settings.system.ai_system')}</h2>
								<p class="text-sm text-muted-foreground">{$t('settings.system.ai_system_desc')}</p>
							</div>
						</div>
					</div>

					<div class="panel-body">
						<div class="max-w-xl space-y-2">
							<Label class="ml-1">{$t('settings.system.wake_up_time')}</Label>
							<div class="flex gap-3">
								<Input
									type="time"
									class="{inputClass} num-mono"
									bind:value={settings['ai_wake_up_time']}
								/>
								<Button
									class="btn-soothe h-11 shrink-0 px-5"
									onclick={() => saveSetting('ai_wake_up_time', settings['ai_wake_up_time'])}
								>
									<Save size={16} />
									{$t('common.save')}
								</Button>
							</div>
							<p class="text-xs text-muted-foreground">
								{$t('settings.system.wake_up_time_hint')}
							</p>
						</div>
					</div>
				</section>
			</div>
		</Tabs.Content>

		<!-- Notifications -->
		<Tabs.Content value="notifications" class="space-y-4 pt-1 outline-none">
			<div in:fly={{ y: 16, duration: 300 }} class="grid gap-4">
				<section class="panel overflow-hidden">
					<div class="panel-header flex-wrap items-start">
						<div class="flex items-start gap-3.5">
							<span
								class="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
							>
								<Bell size={18} />
							</span>
							<div class="space-y-0.5">
								<h2 class="section-title">Notification Subscriptions</h2>
								<p class="text-sm text-muted-foreground">
									Choose which alert types you want to receive notifications for.
								</p>
							</div>
						</div>
						{#if auth.hasPermission('notifications.broadcast')}
							<Button class="btn-soothe h-10 px-4" onclick={() => (broadcastOpen = true)}>
								<Megaphone size={16} />
								Broadcast
							</Button>
						{/if}
					</div>

					<div class="panel-body">
						{#if subscriptionsLoading}
							<div class="flex items-center justify-center py-12">
								<Loader2 class="h-6 w-6 animate-spin text-primary" />
							</div>
						{:else}
							<div class="grid gap-3 sm:grid-cols-2">
								{#each ALERT_TYPES as alertType}
									<div
										class="flex items-center justify-between gap-4 rounded-md border border-border/70 bg-muted/25 px-4 py-3 transition-colors hover:bg-muted/50"
									>
										<div class="min-w-0">
											<p class="text-sm font-semibold text-foreground">{alertType.label}</p>
											<p class="num-mono text-xs text-muted-foreground">{alertType.value}</p>
										</div>
										<Switch
											checked={subscriptions[alertType.value] ?? true}
											onCheckedChange={(checked) => toggleSubscription(alertType.value, checked)}
										/>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				</section>
			</div>
		</Tabs.Content>
	</Tabs.Root>
</div>

<Dialog.Root bind:open={broadcastOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title class="flex items-center gap-2">
				<Megaphone class="h-5 w-5 text-primary" />
				Broadcast Notification
			</Dialog.Title>
			<Dialog.Description>Send a notification to all users.</Dialog.Description>
		</Dialog.Header>
		<div class="space-y-4 py-4">
			<div class="space-y-2">
				<Label>Title</Label>
				<Input bind:value={broadcastForm.title} placeholder="Notification title" />
			</div>
			<div class="space-y-2">
				<Label>Message</Label>
				<Textarea bind:value={broadcastForm.message} placeholder="Notification message" rows={4} />
			</div>
			<div class="space-y-2">
				<Label>Type</Label>
				<Select
					options={[
						{ value: 'SYSTEM', label: 'System' },
						{ value: 'ALERT', label: 'Alert' },
						{ value: 'INFO', label: 'Info' }
					]}
					bind:value={broadcastForm.type}
				/>
			</div>
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (broadcastOpen = false)}>Cancel</Button>
			<Button class="btn-soothe" onclick={sendBroadcast} disabled={broadcastSending}>
				{#if broadcastSending}
					<Loader2 class="mr-2 h-4 w-4 animate-spin" />
				{:else}
					<Send class="mr-2 h-4 w-4" />
				{/if}
				Send
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
