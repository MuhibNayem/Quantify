<script lang="ts">
	import { onMount } from 'svelte';
	import { settingsApi } from '$lib/api/settings';
	import { toast } from 'svelte-sonner';
	import * as Tabs from '$lib/components/ui/tabs';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Switch } from '$lib/components/ui/switch';
	import { Select } from '$lib/components/ui/select';
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
		Percent
	} from 'lucide-svelte';
	import RoleManager from '$lib/components/settings/RoleManager.svelte';
	import { fade, fly } from 'svelte/transition';
	import { cn } from '$lib/utils';

	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	$effect(() => {
		if (!auth.hasPermission('settings.view')) {
			toast.error('Access Denied', { description: 'You do not have permission to view settings.' });
			goto('/');
		}
	});

	let settings: any = {
		business_name: '',
		currency_symbol: '$',
		timezone: 'UTC',
		privacy_policy: '',
		terms_of_service: '',
		loyalty_points_earning_rate: '1',
		loyalty_points_redemption_rate: '0.01',
		loyalty_tier_silver: '500',
		loyalty_tier_gold: '2500',
		loyalty_tier_platinum: '10000',
		tax_rate_percentage: '0'
	};
	let activeTab = 'general';

	const currencyOptions = [
		{ value: '$', label: 'USD - United States Dollar' },
		{ value: '€', label: 'EUR - Euro' },
		{ value: '£', label: 'GBP - British Pound' },
		{ value: '৳', label: 'BDT - Bangladeshi Taka' },
		{ value: '₹', label: 'INR - Indian Rupee' },
		{ value: '₨', label: 'PKR - Pakistani Rupee' },
		{ value: '¥', label: '¥ - JPY / CNY' },
		{ value: 'C$', label: 'CAD - Canadian Dollar' },
		{ value: 'A$', label: 'AUD - Australian Dollar' }
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

	onMount(async () => {
		try {
			const settingsData = await settingsApi.getSettings();
			if (settingsData) {
				Object.values(settingsData)
					.flat()
					.forEach((s: any) => {
						settings[s.Key] = s.Value;
					});
				// Ensure defaults if missing
				if (!settings['currency_symbol']) settings['currency_symbol'] = '$';
				if (!settings['timezone']) settings['timezone'] = 'UTC';
			}
		} catch (e) {
			console.error('Error loading settings', e);
			toast.error('Failed to load settings');
		}
	});

	async function saveSetting(key: string, value: string) {
		try {
			await settingsApi.updateSetting(key, value);
			toast.success('Saved successfully');
		} catch (e) {
			toast.error('Failed to save');
		}
	}

	const allTabs = [
		{ id: 'general', label: 'General', icon: Settings, permission: 'settings.view' },
		{ id: 'business', label: 'Business Rules', icon: Coins, permission: 'settings.view' },
		{ id: 'security', label: 'Security & Roles', icon: ShieldCheck, permission: 'roles.view' },
		{ id: 'policies', label: 'Policies', icon: FileText, permission: 'settings.view' },
		{ id: 'notifications', label: 'Notifications', icon: Bell, permission: 'settings.view' }
	];

	let tabs = $derived(allTabs.filter((t) => !t.permission || auth.hasPermission(t.permission)));

	// Auto-select first available tab if current one becomes hidden
	$effect(() => {
		if (tabs.length > 0 && !tabs.find((t) => t.id === activeTab)) {
			activeTab = tabs[0].id;
		}
	});
</script>

<div class="relative min-h-screen overflow-hidden bg-slate-50 p-6 lg:p-10">
	<!-- Soothing Soft Background -->
	<div class="absolute left-0 top-0 -z-10 h-full w-full overflow-hidden bg-white/50">
		<div
			class="animate-blob absolute -left-[10%] -top-[10%] h-[50%] w-[50%] rounded-full bg-blue-100/60 blur-[100px]"
		></div>
		<div
			class="animate-blob animation-delay-2000 absolute -right-[10%] top-[20%] h-[60%] w-[60%] rounded-full bg-purple-100/60 blur-[100px]"
		></div>
		<div
			class="animate-blob animation-delay-4000 absolute -bottom-[20%] left-[20%] h-[50%] w-[50%] rounded-full bg-pink-100/60 blur-[100px]"
		></div>
	</div>

	<div class="relative z-10 mx-auto max-w-7xl space-y-8">
		<!-- Header -->
		<div class="flex flex-col gap-2 backdrop-blur-sm">
			<h1
				class="bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-4xl font-bold tracking-tight text-transparent drop-shadow-sm"
			>
				Configuration
			</h1>
			<p class="font-medium text-slate-500">
				Manage system preferences, security controls, and global policies.
			</p>
		</div>

		<Tabs.Root
			value={activeTab}
			class="w-full space-y-8"
			onValueChange={(val) => (activeTab = val)}
		>
			<!-- Light Glass Tabs List -->
			<Tabs.List
				class="inline-flex h-auto w-full rounded-2xl border border-white/60 bg-white/40 p-1.5 shadow-lg backdrop-blur-xl"
			>
				{#each tabs as tab}
					<Tabs.Trigger
						value={tab.id}
						class={cn(
							'flex-1 rounded-xl py-3 text-sm font-medium transition-all duration-300',
							'data-[state=active]:bg-white data-[state=active]:text-blue-600 data-[state=active]:shadow-md data-[state=active]:ring-1 data-[state=active]:ring-slate-200/50',
							'text-slate-500 hover:bg-white/40 hover:text-slate-700'
						)}
					>
						<div class="flex items-center justify-center gap-2">
							<tab.icon size={18} />
							{tab.label}
						</div>
					</Tabs.Trigger>
				{/each}
			</Tabs.List>

			<!-- General Settings -->
			<Tabs.Content value="general" class="space-y-6 pt-2 outline-none">
				<div in:fly={{ y: 20, duration: 300 }} class="grid gap-6">
					<!-- Business Profile Card -->
					<div
						class="rounded-3xl border border-white/60 bg-white/60 p-8 shadow-xl backdrop-blur-2xl"
					>
						<div class="mb-8 flex items-start gap-5">
							<div
								class="rounded-2xl border border-white bg-gradient-to-br from-blue-50 to-purple-50 p-4 text-blue-600 shadow-sm"
							>
								<Building2 size={32} />
							</div>
							<div>
								<h3 class="text-xl font-bold text-slate-800">Business Profile</h3>
								<p class="text-slate-500">
									Your organization's visible identity across the platform.
								</p>
							</div>
						</div>

						<div class="max-w-xl space-y-6">
							<div class="space-y-3">
								<Label class="ml-1 font-medium text-slate-600">Business Name</Label>
								<div class="flex gap-3">
									<Input
										class="h-12 rounded-xl border-slate-200 bg-white/50 text-slate-800 shadow-sm transition-all placeholder:text-slate-400 focus:border-blue-500 focus:bg-white focus:ring-blue-500/20"
										bind:value={settings['business_name']}
										placeholder="e.g. Acme Corp"
									/>
									<Button
										class="h-12 rounded-xl bg-gradient-to-r from-blue-600 to-purple-600 px-6 text-white shadow-lg transition-all hover:from-blue-700 hover:to-purple-700 active:scale-95"
										onclick={() => saveSetting('business_name', settings['business_name'])}
									>
										<Save size={20} class="mr-2" /> Save
									</Button>
								</div>
							</div>
						</div>
					</div>

					<!-- Localization -->
					<div class="grid gap-6 md:grid-cols-2">
						<div
							class="group rounded-3xl border border-white/60 bg-white/60 p-8 shadow-xl backdrop-blur-2xl transition-all hover:bg-white/80"
						>
							<div class="mb-6 flex items-center gap-4">
								<div class="rounded-xl bg-blue-50 p-3 text-blue-600 ring-1 ring-blue-100">
									<Globe size={24} />
								</div>
								<h3 class="text-lg font-bold text-slate-800">Currency</h3>
							</div>
							<div class="flex items-end gap-3">
								<div class="flex-1">
									<Select
										options={currencyOptions}
										bind:value={settings['currency_symbol']}
										placeholder="Select Currency"
										style="rounded-xl border-slate-200 bg-white/50 text-slate-800"
									/>
								</div>
								<Button
									class="h-12 w-12 rounded-xl bg-blue-600 text-white shadow-lg transition-all hover:bg-blue-700 active:scale-95"
									onclick={() => saveSetting('currency_symbol', settings['currency_symbol'])}
								>
									<Save size={20} />
								</Button>
							</div>
						</div>

						<div
							class="group rounded-3xl border border-white/60 bg-white/60 p-8 shadow-xl backdrop-blur-2xl transition-all hover:bg-white/80"
						>
							<div class="mb-6 flex items-center gap-4">
								<div class="rounded-xl bg-emerald-50 p-3 text-emerald-600 ring-1 ring-emerald-100">
									<Clock size={24} />
								</div>
								<h3 class="text-lg font-bold text-slate-800">Timezone</h3>
							</div>
							<div class="flex items-end gap-3">
								<div class="flex-1">
									<Select
										options={timezoneOptions}
										bind:value={settings['timezone']}
										placeholder="Select Timezone"
										style="rounded-xl border-slate-200 bg-white/50 text-slate-800"
									/>
								</div>
								<Button
									class="h-12 w-12 rounded-xl bg-emerald-600 text-white shadow-lg transition-all hover:bg-emerald-700 active:scale-95"
									onclick={() => saveSetting('timezone', settings['timezone'])}
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
					<div
						class="rounded-3xl border border-white/60 bg-white/60 p-8 shadow-xl backdrop-blur-2xl"
					>
						<div class="mb-8 flex items-start gap-5">
							<div
								class="rounded-2xl border border-white bg-gradient-to-br from-amber-50 to-orange-50 p-4 text-amber-600 shadow-sm"
							>
								<Sparkles size={32} />
							</div>
							<div>
								<h3 class="text-xl font-bold text-slate-800">Loyalty Program</h3>
								<p class="text-slate-500">
									Configure how customers earn and redeem points, and set tier thresholds.
								</p>
							</div>
						</div>

						<div class="grid gap-8 md:grid-cols-2">
							<!-- Earning & Redemption -->
							<div class="space-y-6">
								<h4 class="font-semibold text-slate-700">Points Configuration</h4>
								<div class="space-y-3">
									<Label class="ml-1 font-medium text-slate-600">Earning Rate (Points per $1)</Label
									>
									<div class="flex gap-3">
										<Input
											type="number"
											step="0.1"
											class="h-12 rounded-xl border-slate-200 bg-white/50 text-slate-800 shadow-sm"
											bind:value={settings['loyalty_points_earning_rate']}
										/>
										<Button
											class="h-12 rounded-xl bg-amber-600 text-white shadow-lg hover:bg-amber-700"
											onclick={() =>
												saveSetting(
													'loyalty_points_earning_rate',
													settings['loyalty_points_earning_rate']
												)}
										>
											<Save size={20} />
										</Button>
									</div>
									<p class="text-xs text-slate-400">
										How many points a customer earns for every unit of currency spent.
									</p>
								</div>

								<div class="space-y-3">
									<Label class="ml-1 font-medium text-slate-600"
										>Redemption Value ($ per Point)</Label
									>
									<div class="flex gap-3">
										<Input
											type="number"
											step="0.01"
											class="h-12 rounded-xl border-slate-200 bg-white/50 text-slate-800 shadow-sm"
											bind:value={settings['loyalty_points_redemption_rate']}
										/>
										<Button
											class="h-12 rounded-xl bg-amber-600 text-white shadow-lg hover:bg-amber-700"
											onclick={() =>
												saveSetting(
													'loyalty_points_redemption_rate',
													settings['loyalty_points_redemption_rate']
												)}
										>
											<Save size={20} />
										</Button>
									</div>
									<p class="text-xs text-slate-400">
										The monetary value of a single loyalty point when redeeming.
									</p>
								</div>
							</div>

							<!-- Tiers -->
							<div class="space-y-6">
								<h4 class="font-semibold text-slate-700">Tier Thresholds</h4>
								<div class="space-y-4">
									<div class="space-y-2">
										<Label class="ml-1 font-medium text-slate-600">Silver Tier (Points)</Label>
										<div class="flex gap-3">
											<Input
												type="number"
												class="h-12 rounded-xl border-slate-200 bg-white/50 text-slate-800 shadow-sm"
												bind:value={settings['loyalty_tier_silver']}
											/>
											<Button
												class="h-12 rounded-xl bg-slate-400 text-white shadow-lg hover:bg-slate-500"
												onclick={() =>
													saveSetting('loyalty_tier_silver', settings['loyalty_tier_silver'])}
											>
												<Save size={20} />
											</Button>
										</div>
									</div>
									<div class="space-y-2">
										<Label class="ml-1 font-medium text-slate-600">Gold Tier (Points)</Label>
										<div class="flex gap-3">
											<Input
												type="number"
												class="h-12 rounded-xl border-slate-200 bg-white/50 text-slate-800 shadow-sm"
												bind:value={settings['loyalty_tier_gold']}
											/>
											<Button
												class="h-12 rounded-xl bg-yellow-500 text-white shadow-lg hover:bg-yellow-600"
												onclick={() =>
													saveSetting('loyalty_tier_gold', settings['loyalty_tier_gold'])}
											>
												<Save size={20} />
											</Button>
										</div>
									</div>
									<div class="space-y-2">
										<Label class="ml-1 font-medium text-slate-600">Platinum Tier (Points)</Label>
										<div class="flex gap-3">
											<Input
												type="number"
												class="h-12 rounded-xl border-slate-200 bg-white/50 text-slate-800 shadow-sm"
												bind:value={settings['loyalty_tier_platinum']}
											/>
											<Button
												class="h-12 rounded-xl bg-slate-800 text-white shadow-lg hover:bg-slate-900"
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
					<div
						class="rounded-3xl border border-white/60 bg-white/60 p-8 shadow-xl backdrop-blur-2xl"
					>
						<div class="mb-8 flex items-start gap-5">
							<div
								class="rounded-2xl border border-white bg-gradient-to-br from-emerald-50 to-teal-50 p-4 text-emerald-600 shadow-sm"
							>
								<Percent size={32} />
							</div>
							<div>
								<h3 class="text-xl font-bold text-slate-800">Financial Settings</h3>
								<p class="text-slate-500">Manage tax rates and other financial parameters.</p>
							</div>
						</div>

						<div class="max-w-xl space-y-6">
							<div class="space-y-3">
								<Label class="ml-1 font-medium text-slate-600">Default Tax Rate (%)</Label>
								<div class="flex gap-3">
									<Input
										type="number"
										step="0.01"
										class="h-12 rounded-xl border-slate-200 bg-white/50 text-slate-800 shadow-sm"
										bind:value={settings['tax_rate_percentage']}
									/>
									<Button
										class="h-12 rounded-xl bg-emerald-600 text-white shadow-lg hover:bg-emerald-700"
										onclick={() =>
											saveSetting('tax_rate_percentage', settings['tax_rate_percentage'])}
									>
										<Save size={20} class="mr-2" /> Save
									</Button>
								</div>
								<p class="text-xs text-slate-400">
									This tax rate will be applied to all applicable sales.
								</p>
							</div>
						</div>
					</div>
				</div>
			</Tabs.Content>

			<!-- Security / RBAC -->
			<Tabs.Content value="security" class="pt-2 outline-none">
				<div in:fly={{ y: 20, duration: 300 }} class="overflow-hidden rounded-3xl shadow-2xl">
					<RoleManager />
				</div>
			</Tabs.Content>

			<!-- Policies -->
			<Tabs.Content value="policies" class="space-y-6 pt-2 outline-none">
				<div in:fly={{ y: 20, duration: 300 }} class="grid gap-6">
					<div
						class="rounded-3xl border border-white/60 bg-white/60 p-8 shadow-xl backdrop-blur-2xl"
					>
						<div class="mb-6 flex items-center gap-4">
							<div class="rounded-xl bg-violet-50 p-3 text-violet-600 ring-1 ring-violet-100">
								<Lock size={24} />
							</div>
							<h3 class="text-lg font-bold text-slate-800">Privacy Policy</h3>
						</div>
						<Textarea
							class="min-h-[200px] rounded-2xl border-slate-200 bg-white/50 p-6 leading-relaxed text-slate-600 shadow-sm placeholder:text-slate-400 focus:border-violet-500 focus:bg-white focus:ring-violet-500/20"
							placeholder="Enter your privacy policy (Markdown supported)..."
							bind:value={settings['privacy_policy']}
						/>
						<div class="mt-4 flex justify-end">
							<Button
								class="rounded-xl bg-violet-600 text-white shadow-lg shadow-violet-500/20 transition-all hover:bg-violet-700 active:scale-95"
								onclick={() => saveSetting('privacy_policy', settings['privacy_policy'])}
							>
								<Save size={18} class="mr-2" /> Save Policy
							</Button>
						</div>
					</div>

					<div
						class="rounded-3xl border border-white/60 bg-white/60 p-8 shadow-xl backdrop-blur-2xl"
					>
						<div class="mb-6 flex items-center gap-4">
							<div class="rounded-xl bg-pink-50 p-3 text-pink-600 ring-1 ring-pink-100">
								<LayoutTemplate size={24} />
							</div>
							<h3 class="text-lg font-bold text-slate-800">Terms of Service</h3>
						</div>
						<Textarea
							class="min-h-[200px] rounded-2xl border-slate-200 bg-white/50 p-6 leading-relaxed text-slate-600 shadow-sm placeholder:text-slate-400 focus:border-pink-500 focus:bg-white focus:ring-pink-500/20"
							placeholder="Enter your terms of service (Markdown supported)..."
							bind:value={settings['terms_of_service']}
						/>
						<div class="mt-4 flex justify-end">
							<Button
								class="rounded-xl bg-pink-600 text-white shadow-lg shadow-pink-500/20 transition-all hover:bg-pink-700 active:scale-95"
								onclick={() => saveSetting('terms_of_service', settings['terms_of_service'])}
							>
								<Save size={18} class="mr-2" /> Save Terms
							</Button>
						</div>
					</div>

					<!-- Return Policy Settings -->
					<div
						class="rounded-3xl border border-white/60 bg-white/60 p-8 shadow-xl backdrop-blur-2xl"
					>
						<div class="mb-6 flex items-center gap-4">
							<div class="rounded-xl bg-orange-50 p-3 text-orange-600 ring-1 ring-orange-100">
								<Clock size={24} />
							</div>
							<h3 class="text-lg font-bold text-slate-800">Return Policy</h3>
						</div>
						<div class="max-w-xl space-y-3">
							<Label class="ml-1 font-medium text-slate-600">Return Window (Days)</Label>
							<div class="flex gap-3">
								<Input
									type="number"
									class="h-12 rounded-xl border-slate-200 bg-white/50 text-slate-800 shadow-sm"
									bind:value={settings['return_window_days']}
									placeholder="e.g. 30"
								/>
								<Button
									class="h-12 rounded-xl bg-orange-600 text-white shadow-lg hover:bg-orange-700 active:scale-95"
									onclick={() => saveSetting('return_window_days', settings['return_window_days'])}
								>
									<Save size={20} class="mr-2" /> Save
								</Button>
							</div>
							<p class="text-xs text-slate-400">
								Number of days after purchase that a customer can request a return.
							</p>
						</div>
					</div>
				</div>
			</Tabs.Content>

			<!-- Notifications -->
			<Tabs.Content value="notifications" class="pt-2 outline-none">
				<div
					in:fly={{ y: 20, duration: 300 }}
					class="flex h-80 flex-col items-center justify-center rounded-3xl border border-dashed border-slate-300 bg-white/40 text-center backdrop-blur-xl"
				>
					<div
						class="mb-6 animate-pulse rounded-full bg-amber-50 p-6 text-amber-500 ring-1 ring-amber-100"
					>
						<Bell size={40} />
					</div>
					<h3 class="text-2xl font-bold text-slate-800">Global Alerts Center</h3>
					<p class="mt-2 max-w-sm text-slate-500">
						Advanced notification routing and webhooks configuration is coming in the next update.
					</p>
				</div>
			</Tabs.Content>
		</Tabs.Root>
	</div>
</div>

<style>
	@keyframes blob {
		0%,
		100% {
			transform: translate(0, 0) scale(1);
		}
		33% {
			transform: translate(30px, -50px) scale(1.1);
		}
		66% {
			transform: translate(-20px, 20px) scale(0.9);
		}
	}
	.animate-blob {
		animation: blob 15s infinite;
	}
	.animation-delay-2000 {
		animation-delay: 2s;
	}
	.animation-delay-4000 {
		animation-delay: 4s;
	}
</style>
