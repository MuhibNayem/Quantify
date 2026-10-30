<!-- client/src/routes/register/+page.svelte -->
<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import BrandPanel from '$lib/components/auth/BrandPanel.svelte';
	import FormPanel from '$lib/components/auth/FormPanel.svelte';
	import {
		authCard,
		authCta,
		authError,
		authIcon,
		authInput,
		authInputPlain,
		authLabel,
		authLink,
		authMenu,
		authMenuItem,
		authMobileTile,
		authOverline,
		authSecondaryBtn,
		authSelect,
		authSub,
		authTitle
	} from '$lib/components/auth/styles';
	import api from '$lib/api';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { fade, slide } from 'svelte/transition';
	import {
		ArrowLeft,
		ArrowRight,
		Sparkles,
		Check,
		ChevronDown,
		Loader2,
		Lock,
		Mail,
		User
	} from 'lucide-svelte';
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

	let currentStep = $state(1);
	const totalSteps = 2;

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

	const strengthColor = $derived(
		passwordStrength >= 75 ? 'bg-success' : passwordStrength >= 50 ? 'bg-warning' : 'bg-danger'
	);

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
</script>

<svelte:window onclick={handleWindowClick} />

<div class="flex min-h-screen w-full">
	<BrandPanel title={$t('auth.register.heroTitle')} subtitle={$t('auth.register.heroSubtitle')} />

	<FormPanel maxWidth="max-w-md">
		<div in:fade={{ duration: 400 }}>
			<!-- Mobile brand -->
			<div class="mb-10 flex items-center gap-3 lg:hidden">
				<div class={authMobileTile}>
					<Sparkles class="h-4.5 w-4.5 text-white" />
				</div>
				<span class="font-serif text-[20px] font-medium tracking-[-0.02em] text-foreground">
					Quantify
				</span>
			</div>

			<header data-animate="fade-up">
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
					<span class="num-mono text-xs text-muted-foreground">
						0{currentStep}/0{totalSteps}
					</span>
				</div>
				<span class={authOverline}>Quantify · Ledger OS</span>
				<h2 class={authTitle}>
					{currentStep === 1 ? $t('auth.register.title') : $t('auth.register.step2Title')}
				</h2>
				<p class={authSub}>
					{currentStep === 1 ? $t('auth.register.subtitle') : $t('auth.register.step2Subtitle')}
				</p>
			</header>

			<div class="{authCard} mt-8 p-6 sm:p-7" data-animate="fade-up" style="animation-delay: 120ms">
				<form method="POST" use:enhance>
					{#if currentStep === 1}
						<div class="space-y-4" in:slide={{ axis: 'x', duration: 300 }}>
							<div class="space-y-1.5">
								<Label for="username" class={authLabel}>
									{$t('auth.register.username')}
								</Label>
								<div class="relative">
									<User class={authIcon} />
									<Input
										id="username"
										type="text"
										placeholder={$t('auth.register.usernamePlaceholder')}
										bind:value={$form.username}
										class={authInput}
									/>
								</div>
								{#if $errors.username}
									<p class={authError}>{$errors.username}</p>
								{/if}
							</div>

							<div class="space-y-1.5">
								<Label for="password" class={authLabel}>
									{$t('auth.register.password')}
								</Label>
								<div class="relative">
									<Lock class={authIcon} />
									<Input
										id="password"
										type="password"
										placeholder="••••••••"
										bind:value={$form.password}
										class={authInput}
									/>
								</div>
								{#if $errors.password}
									<p class={authError}>{$errors.password}</p>
								{/if}
								{#if $form.password}
									<div class="space-y-1.5 pt-1">
										<div class="flex items-center justify-between">
											<span
												class="text-[0.625rem] font-semibold tracking-[0.08em] text-muted-foreground uppercase"
											>
												{$t('auth.register.strength')}
											</span>
											<span class="num-mono text-xs text-muted-foreground">{passwordStrength}%</span
											>
										</div>
										<div class="flex gap-1">
											{#each Array(4) as _, i}
												<div
													class="h-1.5 flex-1 rounded-sm transition-colors duration-300 {passwordStrength >
													i * 25
														? strengthColor
														: 'bg-muted'}"
												></div>
											{/each}
										</div>
									</div>
								{/if}
							</div>

							<div class="space-y-1.5">
								<Label for="confirmPassword" class={authLabel}>
									{$t('auth.register.confirmPassword')}
								</Label>
								<div class="relative">
									<Lock class={authIcon} />
									<Input
										id="confirmPassword"
										type="password"
										placeholder="••••••••"
										bind:value={$form.confirmPassword}
										class={authInput}
									/>
								</div>
								{#if $errors.confirmPassword}
									<p class={authError}>{$errors.confirmPassword}</p>
								{/if}
							</div>

							<div class="relative space-y-1.5">
								<Label for="role" class={authLabel}>
									{$t('auth.register.role')}
								</Label>

								<button
									id="role-dropdown-trigger"
									type="button"
									onclick={toggleRoleDropdown}
									disabled={$submitting}
									class={authSelect}
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
										class={authMenu}
										in:slide={{ duration: 200, axis: 'y' }}
										out:fade={{ duration: 150 }}
									>
										<div class="space-y-0.5">
											{#each availableRoles as role}
												<button
													type="button"
													onclick={() => selectRole(role.value as any)}
													class="{authMenuItem} {$form.role === role.value
														? 'bg-primary/[0.07] font-medium text-primary'
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

							<Button type="button" onclick={nextStep} class="{authCta} group mt-2 w-full">
								{$t('auth.register.nextStep')}
								<ArrowRight
									class="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5"
								/>
							</Button>
						</div>
					{:else}
						<div class="space-y-4" in:slide={{ axis: 'x', duration: 300 }}>
							<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
								<div class="space-y-1.5">
									<Label for="firstName" class={authLabel}>
										{$t('auth.register.firstName')}
									</Label>
									<Input
										id="firstName"
										type="text"
										placeholder={$t('auth.register.firstNamePlaceholder')}
										bind:value={$form.firstName}
										class={authInputPlain}
									/>
									{#if $errors.firstName}
										<p class={authError}>{$errors.firstName}</p>
									{/if}
								</div>

								<div class="space-y-1.5">
									<Label for="lastName" class={authLabel}>
										{$t('auth.register.lastName')}
									</Label>
									<Input
										id="lastName"
										type="text"
										placeholder={$t('auth.register.lastNamePlaceholder')}
										bind:value={$form.lastName}
										class={authInputPlain}
									/>
									{#if $errors.lastName}
										<p class={authError}>{$errors.lastName}</p>
									{/if}
								</div>
							</div>

							<div class="space-y-1.5">
								<Label for="email" class={authLabel}>
									{$t('auth.register.email')}
								</Label>
								<div class="relative">
									<Mail class={authIcon} />
									<Input
										id="email"
										type="email"
										placeholder={$t('auth.register.emailPlaceholder')}
										bind:value={$form.email}
										class={authInput}
									/>
								</div>
								{#if $errors.email}
									<p class={authError}>{$errors.email}</p>
								{/if}
							</div>

							<div class="space-y-1.5">
								<Label for="phoneNumber" class={authLabel}>
									{$t('auth.register.phone')}
								</Label>
								<Input
									id="phoneNumber"
									type="tel"
									placeholder={$t('auth.register.phonePlaceholder')}
									bind:value={$form.phoneNumber}
									class={authInputPlain}
								/>
								{#if $errors.phoneNumber}
									<p class={authError}>{$errors.phoneNumber}</p>
								{/if}
							</div>

							<div class="space-y-1.5">
								<Label for="address" class={authLabel}>
									{$t('auth.register.address')}
								</Label>
								<Input
									id="address"
									type="text"
									placeholder={$t('auth.register.addressPlaceholder')}
									bind:value={$form.address}
									class={authInputPlain}
								/>
							</div>

							<div class="flex gap-3 pt-2">
								<Button
									type="button"
									variant="outline"
									onclick={prevStep}
									class="{authSecondaryBtn} w-1/3"
								>
									<ArrowLeft class="h-4 w-4" />
									{$t('auth.register.back')}
								</Button>
								<Button type="submit" class="{authCta} group flex-1" disabled={$submitting}>
									{#if $submitting}
										<Loader2 class="h-4 w-4 animate-spin" />
										{$t('auth.register.creating')}
									{:else}
										{$t('auth.register.complete')}
										<ArrowRight
											class="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5"
										/>
									{/if}
								</Button>
							</div>
						</div>
					{/if}
				</form>
			</div>

			<p class="mt-6 text-center text-sm text-muted-foreground">
				{$t('auth.register.haveAccount')}{' '}
				<a href="/login" class={authLink}>{$t('auth.register.signIn')}</a>
			</p>
		</div>
	</FormPanel>
</div>
