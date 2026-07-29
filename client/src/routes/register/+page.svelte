<!-- client/src/routes/register/+page.svelte -->
<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import api from '$lib/api';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { fade, slide } from 'svelte/transition';
	import { ArrowLeft, ArrowRight, BarChart3, Check, ChevronDown, Loader2 } from 'lucide-svelte';
	import { superForm } from 'sveltekit-superforms';
	import { zod4 } from 'sveltekit-superforms/adapters';
	import { registerSchema } from '$lib/schemas/auth';
	import { t } from '$lib/i18n';

	const { form, errors, enhance, submitting } = superForm(
		{
			username: '',
			password: '',
			confirmPassword: '',
			role: 'Staff' as 'Admin' | 'Manager' | 'Staff',
			firstName: '',
			lastName: '',
			email: '',
			phoneNumber: '',
			address: ''
		},
		{
			validators: zod4(registerSchema),
			onSubmit: async ({ cancel }) => {
				cancel();
				const data = $form;
				try {
					await api.post('/users/register', {
						username: data.username,
						password: data.password,
						role: data.role,
						firstName: data.firstName || undefined,
						lastName: data.lastName || undefined,
						email: data.email || undefined,
						phoneNumber: data.phoneNumber || undefined,
						address: data.address || undefined
					});
					toast.success($t('auth.register.success'));
					goto('/login');
				} catch (error: any) {
					const errorMessage =
						error.response?.data?.error || error.message || 'Registration failed';
					toast.error($t('auth.register.failed'), {
						description: errorMessage
					});
				}
			}
		}
	);

	let availableRoles: { value: string; label: string }[] = $derived([
		{ value: 'Admin', label: $t('auth.register.roles.admin') },
		{ value: 'Manager', label: $t('auth.register.roles.manager') },
		{ value: 'Staff', label: $t('auth.register.roles.staff') }
	]);
	let passwordStrength = $state(0);

	// Multi-step state
	let currentStep = $state(1);
	const totalSteps = 2;

	// Dropdown state
	let isRoleDropdownOpen = $state(false);

	function calculatePasswordStrength(password: string) {
		let strength = 0;
		if (password.length >= 8) strength += 25;
		if (/[A-Z]/.test(password)) strength += 25;
		if (/[0-9]/.test(password)) strength += 25;
		if (/[^A-Za-z0-9]/.test(password)) strength += 25;
		return strength;
	}

	$effect(() => {
		passwordStrength = calculatePasswordStrength($form.password);
	});

	function nextStep() {
		if (currentStep === 1) {
			if (!$form.username || !$form.password || !$form.confirmPassword) {
				toast.error($t('auth.register.fillAllFields'));
				return;
			}
			if ($form.password !== $form.confirmPassword) {
				toast.error($t('auth.register.passwordMismatch'));
				return;
			}
			if (passwordStrength < 75) {
				toast.error($t('auth.register.weakPassword'));
				return;
			}
		}
		currentStep++;
	}

	function prevStep() {
		currentStep--;
	}

	function toggleRoleDropdown() {
		if (!$submitting) {
			isRoleDropdownOpen = !isRoleDropdownOpen;
		}
	}

	function selectRole(role: 'Admin' | 'Manager' | 'Staff') {
		$form.role = role;
		isRoleDropdownOpen = false;
	}

	function handleWindowClick(event: MouseEvent) {
		if (isRoleDropdownOpen) {
			const target = event.target as HTMLElement;
			if (!target.closest('#role-dropdown-trigger') && !target.closest('#role-dropdown-menu')) {
				isRoleDropdownOpen = false;
			}
		}
	}

	const inputClass =
		'liquid-input h-10 w-full rounded-md bg-card bg-none px-3 text-sm font-normal text-foreground shadow-none ring-0 backdrop-blur-none backdrop-saturate-100 from-card to-card placeholder:text-muted-foreground focus:scale-100 focus:border-primary focus:ring-0 md:text-sm';
	const labelClass = 'text-[0.6875rem] font-semibold tracking-[0.08em] text-muted-foreground uppercase';
	const primaryBtnClass =
		'h-10 rounded-md bg-primary bg-none px-4 text-sm font-medium text-primary-foreground shadow-none hover:bg-primary/90 hover:bg-none';
</script>

<svelte:window onclick={handleWindowClick} />

<div class="flex min-h-screen w-full">
	<!-- Left panel — Ledger brand surface -->
	<div
		class="relative hidden w-[45%] flex-col justify-between overflow-hidden bg-primary p-10 lg:flex xl:p-12"
	>
		<!-- Spreadsheet hairlines + ghost totals column -->
		<div class="pointer-events-none absolute inset-0 select-none" aria-hidden="true">
			<div class="absolute inset-0 flex flex-col justify-between">
				{#each Array(18) as _}
					<div class="h-px w-full bg-primary-foreground/5"></div>
				{/each}
			</div>
			<div class="absolute top-0 right-10 bottom-0 hidden flex-col justify-center gap-5 xl:flex">
				{#each ['9,412.00', '22,806.40', '1,204.90', '61,388.25', '4,072.10'] as n}
					<span class="num-mono text-xs text-primary-foreground/20">{n}</span>
				{/each}
			</div>
		</div>

		<!-- Brand -->
		<div class="relative flex items-center gap-3">
			<div
				class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary-foreground/10 ring-1 ring-primary-foreground/20"
			>
				<BarChart3 class="h-5 w-5 text-primary-foreground" />
			</div>
			<span class="font-display text-xl font-semibold tracking-tight text-primary-foreground"
				>Quantify</span
			>
		</div>

		<!-- Headline + ledger motif -->
		<div class="relative">
			<h1
				class="font-display max-w-md text-4xl leading-[1.08] font-semibold tracking-tight text-primary-foreground xl:text-[2.75rem]"
			>
				{$t('auth.register.heroTitle')}
			</h1>
			<p class="mt-4 max-w-sm text-[0.9375rem] leading-relaxed text-primary-foreground/70">
				{$t('auth.register.heroSubtitle')}
			</p>

			<div
				class="mt-12 max-w-sm overflow-hidden rounded-lg border border-primary-foreground/15"
				aria-hidden="true"
			>
				<div
					class="flex items-center justify-between gap-6 bg-primary-foreground/5 px-4 py-2.5"
				>
					<div class="h-1.5 w-20 rounded-sm bg-primary-foreground/25"></div>
					<div class="h-1.5 w-8 rounded-sm bg-primary-foreground/25"></div>
					<div class="h-1.5 w-12 rounded-sm bg-primary-foreground/25"></div>
				</div>
				{#each [
					{ w: 72, qty: '2,048', val: '14,810.00' },
					{ w: 104, qty: '9,306', val: '58,204.60' },
					{ w: 48, qty: '127', val: '986.40' },
					{ w: 88, qty: '4,715', val: '30,112.90' }
				] as { w, qty, val }}
					<div
						class="flex items-center justify-between gap-6 border-t border-primary-foreground/10 px-4 py-2.5"
					>
						<div class="h-1.5 rounded-sm bg-primary-foreground/15" style="width: {w}px"></div>
						<span class="num-mono text-xs text-primary-foreground/35">{qty}</span>
						<span class="num-mono text-xs text-primary-foreground/50">{val}</span>
					</div>
				{/each}
			</div>
		</div>

		<p class="relative text-xs font-medium tracking-wide text-primary-foreground/60">
			{$t('auth.login.securedBy')}
		</p>
	</div>

	<!-- Right panel — registration form -->
	<div class="relative flex flex-1 items-center justify-center px-6 py-12 sm:px-10">
		<div class="w-full max-w-md" in:fade={{ duration: 400 }}>
			<!-- Mobile brand -->
			<div class="mb-10 flex items-center gap-3 lg:hidden">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary">
					<BarChart3 class="h-5 w-5 text-primary-foreground" />
				</div>
				<span class="font-display text-xl font-semibold tracking-tight text-foreground"
					>Quantify</span
				>
			</div>

			<header>
				<div class="mb-5 flex items-center gap-3">
					<div class="flex flex-1 gap-1.5">
						{#each Array(totalSteps) as _, i}
							<div
								class="h-1 flex-1 rounded-sm transition-colors duration-300 {currentStep > i
									? 'bg-primary'
									: 'bg-border'}"
							></div>
						{/each}
					</div>
					<span class="num-mono text-xs text-muted-foreground">0{currentStep}/0{totalSteps}</span>
				</div>
				<h2
					class="font-display text-[1.625rem] font-semibold tracking-tight text-foreground"
				>
					{currentStep === 1 ? $t('auth.register.title') : $t('auth.register.step2Title')}
				</h2>
				<p class="mt-1.5 text-sm text-muted-foreground">
					{currentStep === 1 ? $t('auth.register.subtitle') : $t('auth.register.step2Subtitle')}
				</p>
			</header>

			<form method="POST" use:enhance class="mt-8">
				{#if currentStep === 1}
					<div class="space-y-4" in:slide={{ axis: 'x', duration: 300 }}>
						<div class="space-y-1.5">
							<Label for="username" class={labelClass}>
								{$t('auth.register.username')}
							</Label>
							<Input
								id="username"
								type="text"
								placeholder={$t('auth.register.usernamePlaceholder')}
								bind:value={$form.username}
								class={inputClass}
							/>
							{#if $errors.username}
								<p class="text-xs font-medium text-danger">{$errors.username}</p>
							{/if}
						</div>

						<div class="space-y-1.5">
							<Label for="password" class={labelClass}>
								{$t('auth.register.password')}
							</Label>
							<Input
								id="password"
								type="password"
								placeholder="••••••••"
								bind:value={$form.password}
								class={inputClass}
							/>
							{#if $errors.password}
								<p class="text-xs font-medium text-danger">{$errors.password}</p>
							{/if}
							{#if $form.password}
								<div class="space-y-1.5 pt-1">
									<div class="flex items-center justify-between">
										<span class="text-[0.625rem] font-semibold tracking-[0.08em] text-muted-foreground uppercase">
											{$t('auth.register.strength')}
										</span>
										<span class="num-mono text-xs text-muted-foreground">{passwordStrength}%</span>
									</div>
									<div class="flex gap-1">
										{#each Array(4) as _, i}
											<div
												class="h-1.5 flex-1 rounded-sm transition-colors duration-300 {passwordStrength >
												i * 25
													? 'bg-primary'
													: 'bg-muted'}"
											></div>
										{/each}
									</div>
								</div>
							{/if}
						</div>

						<div class="space-y-1.5">
							<Label for="confirmPassword" class={labelClass}>
								{$t('auth.register.confirmPassword')}
							</Label>
							<Input
								id="confirmPassword"
								type="password"
								placeholder="••••••••"
								bind:value={$form.confirmPassword}
								class={inputClass}
							/>
							{#if $errors.confirmPassword}
								<p class="text-xs font-medium text-danger">{$errors.confirmPassword}</p>
							{/if}
						</div>

						<div class="relative space-y-1.5">
							<Label for="role" class={labelClass}>
								{$t('auth.register.role')}
							</Label>

							<button
								id="role-dropdown-trigger"
								type="button"
								onclick={toggleRoleDropdown}
								disabled={$submitting}
								class="liquid-input flex h-10 w-full items-center justify-between rounded-md bg-card px-3 text-sm font-normal text-foreground shadow-none transition-colors outline-none disabled:cursor-not-allowed disabled:opacity-50"
							>
								<span class="flex items-center gap-2">
									{availableRoles.find((r) => r.value === $form.role)?.label}
								</span>
								<ChevronDown
									class="h-4 w-4 text-muted-foreground transition-transform duration-300 {isRoleDropdownOpen
										? 'rotate-180'
										: ''}"
								/>
							</button>

							{#if isRoleDropdownOpen}
								<div
									id="role-dropdown-menu"
									class="absolute top-full right-0 left-0 z-50 mt-2 overflow-hidden rounded-md border border-border bg-popover p-1.5 shadow-pop"
									in:slide={{ duration: 200, axis: 'y' }}
									out:fade={{ duration: 150 }}
								>
									<div class="space-y-0.5">
										{#each availableRoles as role}
											<button
												type="button"
												onclick={() => selectRole(role.value as any)}
												class="flex w-full items-center justify-between rounded-sm px-3 py-2 text-left text-sm transition-colors {$form
													.role === role.value
													? 'bg-accent font-medium text-accent-foreground'
													: 'text-foreground hover:bg-muted'}"
											>
												{role.label}
												{#if $form.role === role.value}
													<Check class="h-4 w-4 text-primary" />
												{/if}
											</button>
										{/each}
									</div>
								</div>
							{/if}
						</div>

						<Button type="button" onclick={nextStep} class="{primaryBtnClass} mt-2 w-full">
							{$t('auth.register.nextStep')}
							<ArrowRight class="h-4 w-4" />
						</Button>
					</div>
				{:else}
					<div class="space-y-4" in:slide={{ axis: 'x', duration: 300 }}>
						<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
							<div class="space-y-1.5">
								<Label for="firstName" class={labelClass}>
									{$t('auth.register.firstName')}
								</Label>
								<Input
									id="firstName"
									type="text"
									placeholder={$t('auth.register.firstNamePlaceholder')}
									bind:value={$form.firstName}
									class={inputClass}
								/>
								{#if $errors.firstName}
									<p class="text-xs font-medium text-danger">{$errors.firstName}</p>
								{/if}
							</div>

							<div class="space-y-1.5">
								<Label for="lastName" class={labelClass}>
									{$t('auth.register.lastName')}
								</Label>
								<Input
									id="lastName"
									type="text"
									placeholder={$t('auth.register.lastNamePlaceholder')}
									bind:value={$form.lastName}
									class={inputClass}
								/>
								{#if $errors.lastName}
									<p class="text-xs font-medium text-danger">{$errors.lastName}</p>
								{/if}
							</div>
						</div>

						<div class="space-y-1.5">
							<Label for="email" class={labelClass}>
								{$t('auth.register.email')}
							</Label>
							<Input
								id="email"
								type="email"
								placeholder={$t('auth.register.emailPlaceholder')}
								bind:value={$form.email}
								class={inputClass}
							/>
							{#if $errors.email}
								<p class="text-xs font-medium text-danger">{$errors.email}</p>
							{/if}
						</div>

						<div class="space-y-1.5">
							<Label for="phoneNumber" class={labelClass}>
								{$t('auth.register.phone')}
							</Label>
							<Input
								id="phoneNumber"
								type="tel"
								placeholder={$t('auth.register.phonePlaceholder')}
								bind:value={$form.phoneNumber}
								class={inputClass}
							/>
							{#if $errors.phoneNumber}
								<p class="text-xs font-medium text-danger">{$errors.phoneNumber}</p>
							{/if}
						</div>

						<div class="space-y-1.5">
							<Label for="address" class={labelClass}>
								{$t('auth.register.address')}
							</Label>
							<Input
								id="address"
								type="text"
								placeholder={$t('auth.register.addressPlaceholder')}
								bind:value={$form.address}
								class={inputClass}
							/>
						</div>

						<div class="flex gap-3 pt-2">
							<Button
								type="button"
								variant="outline"
								onclick={prevStep}
								class="h-10 w-1/3 rounded-md border-border bg-card bg-none text-sm font-medium text-foreground shadow-none backdrop-blur-none hover:bg-muted"
							>
								<ArrowLeft class="h-4 w-4" />
								{$t('auth.register.back')}
							</Button>
							<Button
								type="submit"
								class="{primaryBtnClass} flex-1"
								disabled={$submitting}
							>
								{#if $submitting}
									<Loader2 class="h-4 w-4 animate-spin" />
									{$t('auth.register.creating')}
								{:else}
									{$t('auth.register.complete')}
								{/if}
							</Button>
						</div>
					</div>
				{/if}
			</form>

			<p class="mt-6 text-center text-sm text-muted-foreground">
				{$t('auth.register.haveAccount')}{' '}
				<a
					href="/login"
					class="font-medium text-primary underline-offset-4 transition-colors hover:underline"
				>
					{$t('auth.register.signIn')}
				</a>
			</p>
		</div>
	</div>
</div>
