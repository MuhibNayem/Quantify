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

	const inputClass =
		'liquid-input h-12 w-full rounded-md bg-card px-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:ring-0';
	const selectClass =
		'liquid-input h-12 w-full rounded-md bg-card px-3 text-sm text-foreground';
	const saveBtn =
		'bg-primary text-primary-foreground h-12 rounded-md font-medium hover:bg-primary/90';
	const iconBtn =
		'bg-primary text-primary-foreground h-12 w-12 rounded-md hover:bg-primary/90';
	const tabsClass = 'rounded-lg border-border bg-card shadow-sm';

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

<div class="bg-background relative min-h-screen overflow-hidden p-8 lg:p-12">
	<div class="relative z-10 mx-auto max-w-7xl space-y-8">
		<!-- Header -->
		<div class="flex flex-col gap-2" data-animate="fade-up">
			<p class="overline">System · Preferences</p>
			<h1 class="font-display text-foreground text-4xl font-bold tracking-tight">
				{$t('common.configuration')}
			</h1>
			<p class="text-muted-foreground font-medium">
				{$t('common.manage_preferences')}
			</p>
		</div>

		<Tabs.Root
			value={activeTab}
			class="w-full space-y-8"
			onValueChange={(val) => (activeTab = val)}
		>
			<!-- Ledger Tabs -->
			<div class="overflow-x-auto pb-4">
				<GlassTabs
					bind:value={activeTab}
					tabs={tabs.map((tab) => ({ value: tab.id, label: $t(tab.label) }))}
					class="{tabsClass} mx-auto flex-nowrap"
				/>
			</div>

			<!-- General Settings -->
			<Tabs.Content value="general" class="space-y-6 pt-2 outline-none">
				<div in:fly={{ y: 20, duration: 300 }} class="grid gap-6">
					<!-- Business Profile Card -->
					<div class="liquid-panel rounded-lg p-8">
						<div class="mb-8 flex items-start gap-5">
							<div class="bg-accent text-primary rounded-md p-4">
								<Building2 size={32} />
							</div>
							<div>
								<h3 class="font-display text-foreground text-xl font-bold">
									{$t('settings.general.business_profile')}
								</h3>
								<p class="text-muted-foreground">
									{$t('settings.general.business_profile_desc')}
								</p>
							</div>
						</div>

						<div class="max-w-xl space-y-6">
							<div class="space-y-3">
								<Label class="text-foreground ml-1">{$t('settings.general.business_name')}</Label>
								<div class="flex gap-3">
									<Input
										class={inputClass}
										bind:value={settings['business_name']}
										placeholder="e.g. Acme Corp"
									/>
									<Button
										class="{saveBtn} px-6"
										onclick={() => saveSetting('business_name', settings['business_name'])}
									>
										<Save size={20} class="mr-2" />
										{$t('common.save')}
									</Button>
								</div>
							</div>
						</div>
					</div>

					<!-- Localization -->
					<div class="grid gap-6 md:grid-cols-2">
						<div class="liquid-panel rounded-lg p-8">
							<div class="mb-6 flex items-center gap-4">
								<div class="bg-accent text-primary rounded-md p-3">
									<Globe size={24} />
								</div>
								<h3 class="font-display text-foreground text-lg font-bold">
									{$t('settings.general.currency')}
								</h3>
							</div>
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
									class={iconBtn}
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
									<Save size={20} />
								</Button>
							</div>
						</div>

						<div class="liquid-panel rounded-lg p-8">
							<div class="mb-6 flex items-center gap-4">
								<div class="bg-accent text-primary rounded-md p-3">
									<Clock size={24} />
								</div>
								<h3 class="font-display text-foreground text-lg font-bold">
									{$t('settings.general.timezone')}
								</h3>
							</div>
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
									class={iconBtn}
									onclick={() => saveSetting('timezone', settings['timezone'])}
								>
									<Save size={20} />
								</Button>
							</div>
						</div>

						<div class="liquid-panel rounded-lg p-8">
							<div class="mb-6 flex items-center gap-4">
								<div class="bg-accent text-primary rounded-md p-3">
									<Globe size={24} />
								</div>
								<h3 class="font-display text-foreground text-lg font-bold">
									{$t('settings.general.locale')}
								</h3>
							</div>
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
									class={iconBtn}
									onclick={() => saveSetting('locale', settings['locale'])}
								>
									<Save size={20} />
								</Button>
							</div>
						</div>
					</div>
				</div>
			</Tabs.Content>

			<!-- Business Rules -->
			<Tabs.Content value="business" class="space-y-6 pt-2 outline-none">
				<div in:fly={{ y: 20, duration: 300 }} class="grid gap-6">
					<!-- Loyalty Program -->
					<div class="liquid-panel rounded-lg p-8">
						<div class="mb-8 flex items-start gap-5">
							<div class="bg-accent text-primary rounded-md p-4">
								<Sparkles size={32} />
							</div>
							<div>
								<h3 class="font-display text-foreground text-xl font-bold">
									{$t('settings.business.loyalty_program')}
								</h3>
								<p class="text-muted-foreground">
									{$t('settings.business.loyalty_program_desc')}
								</p>
							</div>
						</div>

						<div class="grid gap-8 md:grid-cols-2">
							<!-- Earning & Redemption -->
							<div class="space-y-6">
								<h4 class="overline">
									{$t('settings.business.points_configuration')}
								</h4>
								<div class="space-y-3">
									<Label class="text-foreground ml-1 flex items-center gap-1 font-medium">
										{$t('settings.business.earning_rate')}
										<HelpTip text="Loyalty points earned per unit of currency spent. For example, 1 means a customer earns 1 point for every $1 spent." />
									</Label>
									<div class="flex gap-3">
										<Input
											type="number"
											step="0.1"
											class="{inputClass} num-mono"
											bind:value={settings['loyalty_points_earning_rate']}
										/>
										<Button
											class={iconBtn}
											onclick={() =>
												saveSetting(
													'loyalty_points_earning_rate',
													settings['loyalty_points_earning_rate']
												)}
										>
											<Save size={20} />
										</Button>
									</div>
									<p class="text-muted-foreground text-xs">
										{$t('settings.business.earning_rate_hint')}
									</p>
								</div>

								<div class="space-y-3">
									<Label class="text-foreground ml-1 font-medium"
										>{$t('settings.business.redemption_value')}</Label
									>
									<div class="flex gap-3">
										<Input
											type="number"
											step="0.01"
											class="{inputClass} num-mono"
											bind:value={settings['loyalty_points_redemption_rate']}
										/>
										<Button
											class={iconBtn}
											onclick={() =>
												saveSetting(
													'loyalty_points_redemption_rate',
													settings['loyalty_points_redemption_rate']
												)}
										>
											<Save size={20} />
										</Button>
									</div>
									<p class="text-muted-foreground text-xs">
										{$t('settings.business.redemption_value_hint')}
									</p>
								</div>
							</div>

							<!-- Tiers -->
							<div class="space-y-6">
								<h4 class="overline">
									{$t('settings.business.tier_thresholds')}
								</h4>
								<div class="space-y-4">
									<div class="space-y-2">
										<Label class="text-foreground ml-1 font-medium"
											>{$t('settings.business.silver_tier')}</Label
										>
										<div class="flex gap-3">
											<Input
												type="number"
												class="{inputClass} num-mono"
												bind:value={settings['loyalty_tier_silver']}
											/>
											<Button
												class={iconBtn}
												onclick={() =>
													saveSetting('loyalty_tier_silver', settings['loyalty_tier_silver'])}
											>
												<Save size={20} />
											</Button>
										</div>
									</div>
									<div class="space-y-2">
										<Label class="text-foreground ml-1 font-medium"
											>{$t('settings.business.gold_tier')}</Label
										>
										<div class="flex gap-3">
											<Input
												type="number"
												class="{inputClass} num-mono"
												bind:value={settings['loyalty_tier_gold']}
											/>
											<Button
												class={iconBtn}
												onclick={() =>
													saveSetting('loyalty_tier_gold', settings['loyalty_tier_gold'])}
											>
												<Save size={20} />
											</Button>
										</div>
									</div>
									<div class="space-y-2">
										<Label class="text-foreground ml-1 font-medium"
											>{$t('settings.business.platinum_tier')}</Label
										>
										<div class="flex gap-3">
											<Input
												type="number"
												class="{inputClass} num-mono"
												bind:value={settings['loyalty_tier_platinum']}
											/>
											<Button
												class={iconBtn}
												onclick={() =>
													saveSetting('loyalty_tier_platinum', settings['loyalty_tier_platinum'])}
											>
												<Save size={20} />
											</Button>
										</div>
									</div>
								</div>
							</div>
						</div>
					</div>

					<!-- Tax Settings -->
					<div class="liquid-panel rounded-lg p-8">
						<div class="mb-8 flex items-start gap-5">
							<div class="bg-accent text-primary rounded-md p-4">
								<Percent size={32} />
							</div>
							<div>
								<h3 class="font-display text-foreground text-xl font-bold">
									{$t('settings.business.financial_settings')}
								</h3>
								<p class="text-muted-foreground">{$t('settings.business.financial_settings_desc')}</p>
							</div>
						</div>

						<div class="max-w-xl space-y-6">
							<div class="space-y-3">
								<Label class="text-foreground ml-1 flex items-center gap-1 font-medium">
									{$t('settings.business.default_tax_rate')}
									<HelpTip text="Default tax percentage applied to sales at checkout. Enter 5 for 5%." />
								</Label>
								<div class="flex gap-3">
									<Input
										type="number"
										step="0.01"
										class="{inputClass} num-mono"
										bind:value={settings['tax_rate_percentage']}
									/>
									<Button
										class="{saveBtn} px-6"
										onclick={() =>
											saveSetting('tax_rate_percentage', settings['tax_rate_percentage'])}
									>
										<Save size={20} class="mr-2" />
										{$t('common.save')}
									</Button>
								</div>
								<p class="text-muted-foreground text-xs">
									{$t('settings.business.default_tax_rate_hint')}
								</p>
							</div>
						</div>
					</div>
				</div>
			</Tabs.Content>

			<!-- Security / RBAC -->
			<Tabs.Content value="security" class="pt-2 outline-none">
				<div in:fly={{ y: 20, duration: 300 }} class="overflow-hidden rounded-lg">
					<RoleManager />
				</div>
			</Tabs.Content>

			<!-- Policies -->
			<Tabs.Content value="policies" class="space-y-6 pt-2 outline-none">
				<div in:fly={{ y: 20, duration: 300 }} class="grid gap-6">
					<div class="liquid-panel rounded-lg p-8">
						<div class="mb-6 flex items-center gap-4">
							<div class="bg-accent text-primary rounded-md p-3">
								<Lock size={24} />
							</div>
							<h3 class="font-display text-foreground text-lg font-bold">
								{$t('settings.policies.privacy_policy')}
							</h3>
						</div>
						<Textarea
							class="liquid-textarea text-foreground min-h-[200px] rounded-md p-6 leading-relaxed placeholder:text-muted-foreground"
							placeholder={$t('settings.policies.privacy_policy_placeholder')}
							bind:value={settings['privacy_policy']}
						/>
						<div class="mt-4 flex justify-end">
							<Button
								class="bg-primary text-primary-foreground rounded-md font-medium hover:bg-primary/90"
								onclick={() => saveSetting('privacy_policy', settings['privacy_policy'])}
							>
								<Save size={18} class="mr-2" />
								{$t('settings.policies.save_policy')}
							</Button>
						</div>
					</div>

					<div class="liquid-panel rounded-lg p-8">
						<div class="mb-6 flex items-center gap-4">
							<div class="bg-accent text-primary rounded-md p-3">
								<LayoutTemplate size={24} />
							</div>
							<h3 class="font-display text-foreground text-lg font-bold">
								{$t('settings.policies.terms_of_service')}
							</h3>
						</div>
						<Textarea
							class="liquid-textarea text-foreground min-h-[200px] rounded-md p-6 leading-relaxed placeholder:text-muted-foreground"
							placeholder={$t('settings.policies.terms_of_service_placeholder')}
							bind:value={settings['terms_of_service']}
						/>
						<div class="mt-4 flex justify-end">
							<Button
								class="bg-primary text-primary-foreground rounded-md font-medium hover:bg-primary/90"
								onclick={() => saveSetting('terms_of_service', settings['terms_of_service'])}
							>
								<Save size={18} class="mr-2" />
								{$t('settings.policies.save_terms')}
							</Button>
						</div>
					</div>

					<!-- Return Policy Settings -->
					<div class="liquid-panel rounded-lg p-8">
						<div class="mb-6 flex items-center gap-4">
							<div class="bg-accent text-primary rounded-md p-3">
								<Clock size={24} />
							</div>
							<h3 class="font-display text-foreground text-lg font-bold">
								{$t('settings.policies.return_policy')}
							</h3>
						</div>
						<div class="max-w-xl space-y-3">
							<Label class="text-foreground ml-1 font-medium"
								>{$t('settings.policies.return_window')}</Label
							>
							<div class="flex gap-3">
								<Input
									type="number"
									class="{inputClass} num-mono"
									bind:value={settings['return_window_days']}
									placeholder="e.g. 30"
								/>
								<Button
									class="{saveBtn} px-6"
									onclick={() => saveSetting('return_window_days', settings['return_window_days'])}
								>
									<Save size={20} class="mr-2" />
									{$t('common.save')}
								</Button>
							</div>
							<p class="text-muted-foreground text-xs">
								{$t('settings.policies.return_window_hint')}
							</p>
						</div>
					</div>
				</div>
			</Tabs.Content>

			<!-- System & AI -->
			<Tabs.Content value="system" class="space-y-6 pt-2 outline-none">
				<div in:fly={{ y: 20, duration: 300 }} class="grid gap-6">
					<div class="liquid-panel rounded-lg p-8">
						<div class="mb-8 flex items-start gap-5">
							<div class="bg-accent text-primary rounded-md p-4">
								<Zap size={32} />
							</div>
							<div>
								<h3 class="font-display text-foreground text-xl font-bold">
									{$t('settings.system.ai_system')}
								</h3>
								<p class="text-muted-foreground">
									{$t('settings.system.ai_system_desc')}
								</p>
							</div>
						</div>

						<div class="max-w-xl space-y-6">
							<div class="space-y-3">
								<Label class="text-foreground ml-1 font-medium"
									>{$t('settings.system.wake_up_time')}</Label
								>
								<div class="flex gap-3">
									<Input
										type="time"
										class="{inputClass} num-mono"
										bind:value={settings['ai_wake_up_time']}
									/>
									<Button
										class="{saveBtn} px-6"
										onclick={() => saveSetting('ai_wake_up_time', settings['ai_wake_up_time'])}
									>
										<Save size={20} class="mr-2" />
										{$t('common.save')}
									</Button>
								</div>
								<p class="text-muted-foreground text-xs">
									{$t('settings.system.wake_up_time_hint')}
								</p>
							</div>
						</div>
					</div>
				</div>
			</Tabs.Content>

			<!-- Notifications -->
			<Tabs.Content value="notifications" class="space-y-6 pt-2 outline-none">
				<div in:fly={{ y: 20, duration: 300 }} class="grid gap-6">
					<div class="liquid-panel rounded-lg p-8">
						<div class="mb-8 flex items-start justify-between gap-5">
							<div class="flex items-start gap-5">
								<div class="bg-accent text-primary rounded-md p-4">
									<Bell size={32} />
								</div>
								<div>
									<h3 class="font-display text-foreground text-xl font-bold">
										Notification Subscriptions
									</h3>
									<p class="text-muted-foreground">
										Choose which alert types you want to receive notifications for.
									</p>
								</div>
							</div>
							{#if auth.hasPermission('notifications.broadcast')}
								<Button
									class="bg-primary text-primary-foreground rounded-md px-5 font-medium hover:bg-primary/90"
									onclick={() => (broadcastOpen = true)}
								>
									<Megaphone size={18} class="mr-2" />
									Broadcast
								</Button>
							{/if}
						</div>

						{#if subscriptionsLoading}
							<div class="flex items-center justify-center py-12">
								<Loader2 class="text-primary h-8 w-8 animate-spin" />
							</div>
						{:else}
							<div class="grid gap-4 sm:grid-cols-2">
								{#each ALERT_TYPES as alertType}
									<div
										class="border-border bg-card hover:bg-muted/60 flex items-center justify-between rounded-md border px-5 py-4 shadow-xs transition-colors"
									>
										<div>
											<p class="text-foreground font-semibold">{alertType.label}</p>
											<p class="text-muted-foreground num-mono text-xs">{alertType.value}</p>
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
				</div>
			</Tabs.Content>
		</Tabs.Root>
	</div>
</div>

<Dialog.Root bind:open={broadcastOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title class="text-foreground flex items-center gap-2">
				<Megaphone class="text-primary h-5 w-5" />
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
			<Button
				class="bg-primary text-primary-foreground hover:bg-primary/90"
				onclick={sendBroadcast}
				disabled={broadcastSending}
			>
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
