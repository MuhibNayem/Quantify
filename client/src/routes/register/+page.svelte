<!-- client/src/routes/register/+page.svelte -->
<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import api from '$lib/api';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { fade, slide } from 'svelte/transition';
	import { ArrowLeft, ArrowRight, ChevronDown, Check } from 'lucide-svelte';

	let username = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let firstName = $state('');
	let lastName = $state('');
	let email = $state('');
	let phoneNumber = $state('');
	let address = $state('');
	let selectedRole = $state<'Admin' | 'Manager' | 'Staff'>('Staff');
	let availableRoles: { value: string; label: string }[] = [
		{ value: 'Admin', label: 'Administrator' },
		{ value: 'Manager', label: 'Manager' },
		{ value: 'Staff', label: 'Staff Member' }
	];
	let loading = $state(false);
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
		passwordStrength = calculatePasswordStrength(password);
	});

	function nextStep() {
		if (currentStep === 1) {
			if (!username || !password || !confirmPassword) {
				toast.error('Please fill in all fields');
				return;
			}
			if (password !== confirmPassword) {
				toast.error('Passwords do not match');
				return;
			}
			if (passwordStrength < 75) {
				toast.error('Please choose a stronger password');
				return;
			}
		}
		currentStep++;
	}

	function prevStep() {
		currentStep--;
	}

	function toggleRoleDropdown() {
		if (!loading) {
			isRoleDropdownOpen = !isRoleDropdownOpen;
		}
	}

	function selectRole(role: 'Admin' | 'Manager' | 'Staff') {
		selectedRole = role;
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

	async function handleRegister() {
		loading = true;

		try {
			await api.post('/users/register', {
				username,
				password,
				role: selectedRole,
				firstName: firstName || undefined,
				lastName: lastName || undefined,
				email: email || undefined,
				phoneNumber: phoneNumber || undefined,
				address: address || undefined
			});
			toast.success('Registration successful! Please log in.');
			goto('/login');
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || error.message || 'Registration failed';
			toast.error('Registration Failed', {
				description: errorMessage
			});
		} finally {
			loading = false;
		}
	}
</script>

<svelte:window onclick={handleWindowClick} />

<div
	class="flex min-h-screen w-full bg-gradient-to-br from-slate-50/90 via-sky-50/70 to-indigo-50/80"
>
	<!-- Left side - Image section (60% width) -->
	<div class="relative hidden overflow-hidden lg:flex lg:flex-[1.5]">
		<!-- Background image -->
		<img
			src="/register-visual-feature.png"
			alt="POS and Growth Visualization"
			class="absolute inset-0 h-full w-full object-cover object-center"
		/>

		<!-- Strong glass overlay -->
		<div
			class="absolute inset-0 bg-gradient-to-t from-emerald-900/80 via-transparent to-transparent"
		></div>

		<div class="absolute bottom-20 left-12 right-12 max-w-2xl text-white">
			<h2 class="text-5xl font-bold tracking-tight drop-shadow-xl">Accelerate Your Growth</h2>
			<p class="mt-6 text-xl font-medium text-emerald-100 drop-shadow-lg">
				Join the platform that turns data into actionable insights for millions.
			</p>
		</div>
	</div>

	<!-- Right side - Registration form (40% width) -->
	<div class="relative flex flex-1 items-center justify-center overflow-hidden p-8 lg:p-12">
		<!-- Ambient Background -->
		<div class="absolute inset-0 -z-10 opacity-70">
			<div
				class="absolute -right-20 -top-20 h-[500px] w-[500px] rounded-full bg-emerald-200/50 blur-[100px]"
			></div>
			<div
				class="absolute bottom-0 left-0 h-[400px] w-[400px] rounded-full bg-teal-200/50 blur-[100px]"
			></div>
			<div
				class="absolute right-1/4 top-1/3 h-[300px] w-[300px] rounded-full bg-lime-100/40 blur-[80px]"
			></div>
		</div>

		<div class="relative z-10 w-full max-w-xl space-y-8" in:fade={{ duration: 400 }}>
			<!-- Mobile logo -->
			<div class="mb-8 flex justify-center lg:hidden">
				<div class="flex items-center space-x-3">
					<div
						class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-emerald-500 to-teal-600 shadow-lg shadow-emerald-500/30"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-6 w-6 text-white"
							viewBox="0 0 20 20"
							fill="currentColor"
						>
							<path
								d="M8 9a3 3 0 100-6 3 3 0 000 6zM8 11a6 6 0 016 6H2a6 6 0 016-6zM16 7a1 1 0 10-2 0v1h-1a1 1 0 100 2h1v1a1 1 0 102 0v-1h1a1 1 0 100-2h-1V7z"
							/>
						</svg>
					</div>
					<span
						class="bg-gradient-to-r from-slate-900 to-slate-700 bg-clip-text text-2xl font-bold text-transparent dark:from-white dark:to-slate-300"
						>InventoryPro</span
					>
				</div>
			</div>

			<!-- True Liquid Glass Panel -->
			<Card
				class="relative overflow-visible rounded-[40px] border border-white/20 bg-gradient-to-br from-white/70 to-white/30 p-0 shadow-[0_40px_100px_-20px_rgba(16,185,129,0.15),inset_0_1px_0_0_rgba(255,255,255,0.8),inset_0_-2px_5px_0_rgba(0,0,0,0.05)] backdrop-blur-3xl backdrop-saturate-[180%]"
			>
				<!-- Light flare overlay (Clipped Container) -->
				<div class="pointer-events-none absolute inset-0 overflow-hidden rounded-[40px]">
					<div
						class="absolute -left-1/2 -top-1/2 h-full w-full rotate-12 bg-gradient-to-b from-white/20 to-transparent blur-3xl"
					></div>
				</div>

				<CardHeader class="relative pb-2 pt-12 text-center">
					<div class="mb-6 flex justify-center gap-2">
						{#each Array(totalSteps) as _, i}
							<div
								class="h-1.5 w-10 rounded-full transition-all duration-300 {currentStep > i
									? 'bg-emerald-500 shadow-[0_0_10px_rgba(16,185,129,0.5)]'
									: 'bg-slate-300/50'}"
							></div>
						{/each}
					</div>
					<CardTitle class="text-4xl font-bold tracking-tight text-slate-900">
						{currentStep === 1 ? 'Create Account' : 'Personal Details'}
					</CardTitle>
					<CardDescription class="text-lg font-medium text-slate-600">
						{currentStep === 1 ? 'Start your journey with us' : 'Tell us a bit about yourself'}
					</CardDescription>
				</CardHeader>
				<CardContent class="relative p-10 pt-8">
					<form
						onsubmit={(event) => {
							event.preventDefault();
							if (currentStep === totalSteps) handleRegister();
						}}
						class="space-y-6"
					>
						{#if currentStep === 1}
							<div class="space-y-6" in:slide={{ axis: 'x', duration: 300 }}>
								<div class="space-y-2">
									<Label for="username" class="ml-1 text-sm font-bold tracking-wide text-slate-800">
										Username
									</Label>
									<Input
										id="username"
										type="text"
										placeholder="Choose a username"
										bind:value={username}
										required
										class="h-14 rounded-2xl border border-white/60 bg-gradient-to-br from-white/90 to-white/60 px-5 text-lg font-bold text-slate-900 shadow-[inset_0_2px_4px_rgba(255,255,255,0.9),0_4px_10px_rgba(0,0,0,0.05)] ring-1 ring-white/40 backdrop-blur-xl backdrop-saturate-[180%] transition-all placeholder:text-slate-500/80 focus:scale-[1.01] focus:border-emerald-500/30 focus:ring-4 focus:ring-emerald-500/10"
									/>
								</div>

								<div class="space-y-2">
									<Label for="password" class="ml-1 text-sm font-bold tracking-wide text-slate-800">
										Password
									</Label>
									<Input
										id="password"
										type="password"
										placeholder="••••••••"
										bind:value={password}
										required
										class="h-14 rounded-2xl border border-white/60 bg-gradient-to-br from-white/90 to-white/60 px-5 text-lg font-bold text-slate-900 shadow-[inset_0_2px_4px_rgba(255,255,255,0.9),0_4px_10px_rgba(0,0,0,0.05)] ring-1 ring-white/40 backdrop-blur-xl backdrop-saturate-[180%] transition-all placeholder:text-slate-500/80 focus:scale-[1.01] focus:border-emerald-500/30 focus:ring-4 focus:ring-emerald-500/10"
									/>
									{#if password}
										<div class="space-y-2 px-1 pt-2">
											<div class="flex justify-between text-xs font-bold text-slate-600">
												<span>STRENGTH</span>
												<span>{passwordStrength}%</span>
											</div>
											<div class="h-2 w-full rounded-full bg-slate-200/50">
												<div
													class="h-2 rounded-full shadow-sm transition-all duration-500 ease-out"
													class:bg-rose-500={passwordStrength < 50}
													class:bg-amber-500={passwordStrength >= 50 && passwordStrength < 75}
													class:bg-emerald-500={passwordStrength >= 75}
													style={`width: ${passwordStrength}%`}
												></div>
											</div>
										</div>
									{/if}
								</div>

								<div class="space-y-2">
									<Label
										for="confirmPassword"
										class="ml-1 text-sm font-bold tracking-wide text-slate-800"
									>
										Confirm Password
									</Label>
									<Input
										id="confirmPassword"
										type="password"
										placeholder="••••••••"
										bind:value={confirmPassword}
										required
										class="h-14 rounded-2xl border border-white/60 bg-gradient-to-br from-white/90 to-white/60 px-5 text-lg font-bold text-slate-900 shadow-[inset_0_2px_4px_rgba(255,255,255,0.9),0_4px_10px_rgba(0,0,0,0.05)] ring-1 ring-white/40 backdrop-blur-xl backdrop-saturate-[180%] transition-all placeholder:text-slate-500/80 focus:scale-[1.01] focus:border-emerald-500/30 focus:ring-4 focus:ring-emerald-500/10"
									/>
								</div>

								<div class="relative space-y-2">
									<Label for="role" class="ml-1 text-sm font-bold tracking-wide text-slate-800">
										Role
									</Label>

									<!-- Custom Liquid Dropdown Trigger -->
									<button
										id="role-dropdown-trigger"
										type="button"
										onclick={toggleRoleDropdown}
										class="flex h-14 w-full items-center justify-between rounded-2xl border border-white/60 bg-gradient-to-br from-white/90 to-white/60 px-5 text-lg font-bold text-slate-900 shadow-[inset_0_2px_4px_rgba(255,255,255,0.9),0_4px_10px_rgba(0,0,0,0.05)] outline-none ring-1 ring-white/40 backdrop-blur-xl backdrop-saturate-[180%] transition-all focus:scale-[1.01] focus:border-emerald-500/30 focus:ring-4 focus:ring-emerald-500/10"
										disabled={loading}
									>
										<span class="flex items-center gap-2">
											{availableRoles.find((r) => r.value === selectedRole)?.label}
										</span>
										<ChevronDown
											class="h-5 w-5 text-slate-600 transition-transform duration-300 {isRoleDropdownOpen
												? 'rotate-180'
												: ''}"
										/>
									</button>

									<!-- Custom Liquid Dropdown Menu -->
									{#if isRoleDropdownOpen}
										<div
											id="role-dropdown-menu"
											class="top-FULL absolute left-0 right-0 z-50 mt-2 overflow-hidden rounded-2xl border border-white/60 bg-white/90 shadow-[0_20px_40px_-10px_rgba(0,0,0,0.1),inset_0_1px_0_0_rgba(255,255,255,0.8)] ring-1 ring-white/40 backdrop-blur-2xl"
											in:slide={{ duration: 200, axis: 'y' }}
											out:fade={{ duration: 150 }}
										>
											<div class="space-y-1 p-1.5">
												{#each availableRoles as role}
													<button
														type="button"
														onclick={() => selectRole(role.value as any)}
														class="flex w-full items-center justify-between rounded-xl px-4 py-3 text-left font-bold transition-colors hover:bg-emerald-500/10 {selectedRole ===
														role.value
															? 'bg-emerald-500/10 text-emerald-800'
															: 'text-slate-700 hover:text-emerald-700'}"
													>
														{role.label}
														{#if selectedRole === role.value}
															<Check class="h-4 w-4 text-emerald-600" />
														{/if}
													</button>
												{/each}
											</div>
										</div>
									{/if}
								</div>
								<Button
									type="button"
									onclick={nextStep}
									class="glass-button h-14 w-full rounded-2xl bg-gradient-to-r from-emerald-500 to-teal-600 text-xl font-bold text-white shadow-[0_20px_40px_-15px_rgba(16,185,129,0.4),inset_0_1px_0_0_rgba(255,255,255,0.4)] transition-all hover:bg-gradient-to-r hover:from-emerald-400 hover:to-teal-500 active:scale-[0.98]"
								>
									Next Step <ArrowRight class="ml-2 h-6 w-6" />
								</Button>
							</div>
						{:else}
							<div class="space-y-6" in:slide={{ axis: 'x', duration: 300 }}>
								<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
									<div class="space-y-2">
										<Label
											for="firstName"
											class="ml-1 text-sm font-bold tracking-wide text-slate-800"
										>
											First Name
										</Label>
										<Input
											id="firstName"
											type="text"
											placeholder="John"
											bind:value={firstName}
											required
											class="h-14 rounded-2xl border border-white/60 bg-gradient-to-br from-white/90 to-white/60 px-5 text-lg font-bold text-slate-900 shadow-[inset_0_2px_4px_rgba(255,255,255,0.9),0_4px_10px_rgba(0,0,0,0.05)] ring-1 ring-white/40 backdrop-blur-xl backdrop-saturate-[180%] transition-all placeholder:text-slate-500/80 focus:scale-[1.01] focus:border-emerald-500/30 focus:ring-4 focus:ring-emerald-500/10"
										/>
									</div>

									<div class="space-y-2">
										<Label
											for="lastName"
											class="ml-1 text-sm font-bold tracking-wide text-slate-800"
										>
											Last Name
										</Label>
										<Input
											id="lastName"
											type="text"
											placeholder="Doe"
											bind:value={lastName}
											required
											class="h-14 rounded-2xl border border-white/60 bg-gradient-to-br from-white/90 to-white/60 px-5 text-lg font-bold text-slate-900 shadow-[inset_0_2px_4px_rgba(255,255,255,0.9),0_4px_10px_rgba(0,0,0,0.05)] ring-1 ring-white/40 backdrop-blur-xl backdrop-saturate-[180%] transition-all placeholder:text-slate-500/80 focus:scale-[1.01] focus:border-emerald-500/30 focus:ring-4 focus:ring-emerald-500/10"
										/>
									</div>
								</div>

								<div class="space-y-2">
									<Label for="email" class="ml-1 text-sm font-bold tracking-wide text-slate-800">
										Email
									</Label>
									<Input
										id="email"
										type="email"
										placeholder="john@example.com"
										bind:value={email}
										required
										class="h-14 rounded-2xl border border-white/60 bg-gradient-to-br from-white/90 to-white/60 px-5 text-lg font-bold text-slate-900 shadow-[inset_0_2px_4px_rgba(255,255,255,0.9),0_4px_10px_rgba(0,0,0,0.05)] ring-1 ring-white/40 backdrop-blur-xl backdrop-saturate-[180%] transition-all placeholder:text-slate-500/80 focus:scale-[1.01] focus:border-emerald-500/30 focus:ring-4 focus:ring-emerald-500/10"
									/>
								</div>

								<div class="space-y-2">
									<Label
										for="phoneNumber"
										class="ml-1 text-sm font-bold tracking-wide text-slate-800"
									>
										Phone
									</Label>
									<Input
										id="phoneNumber"
										type="tel"
										placeholder="(555) 123-4567"
										bind:value={phoneNumber}
										required
										class="h-14 rounded-2xl border border-white/60 bg-gradient-to-br from-white/90 to-white/60 px-5 text-lg font-bold text-slate-900 shadow-[inset_0_2px_4px_rgba(255,255,255,0.9),0_4px_10px_rgba(0,0,0,0.05)] ring-1 ring-white/40 backdrop-blur-xl backdrop-saturate-[180%] transition-all placeholder:text-slate-500/80 focus:scale-[1.01] focus:border-emerald-500/30 focus:ring-4 focus:ring-emerald-500/10"
									/>
								</div>

								<div class="space-y-2">
									<Label for="address" class="ml-1 text-sm font-bold tracking-wide text-slate-800">
										Address
									</Label>
									<Input
										id="address"
										type="text"
										placeholder="123 Main Street"
										bind:value={address}
										class="h-14 rounded-2xl border border-white/60 bg-gradient-to-br from-white/90 to-white/60 px-5 text-lg font-bold text-slate-900 shadow-[inset_0_2px_4px_rgba(255,255,255,0.9),0_4px_10px_rgba(0,0,0,0.05)] ring-1 ring-white/40 backdrop-blur-xl backdrop-saturate-[180%] transition-all placeholder:text-slate-500/80 focus:scale-[1.01] focus:border-emerald-500/30 focus:ring-4 focus:ring-emerald-500/10"
									/>
								</div>

								<div class="flex gap-4 pt-4">
									<Button
										type="button"
										variant="outline"
										onclick={prevStep}
										class="h-14 w-1/3 rounded-2xl border-white/60 bg-white/40 text-lg font-bold text-slate-700 hover:bg-white/60"
									>
										<ArrowLeft class="mr-2 h-5 w-5" /> Back
									</Button>
									<Button
										type="submit"
										class="glass-button h-14 flex-1 rounded-2xl bg-gradient-to-r from-emerald-500 to-teal-600 text-xl font-bold text-white shadow-[0_20px_40px_-15px_rgba(16,185,129,0.4),inset_0_1px_0_0_rgba(255,255,255,0.4)] transition-all hover:bg-gradient-to-r hover:from-emerald-400 hover:to-teal-500 active:scale-[0.98] disabled:opacity-70"
										disabled={loading}
									>
										{#if loading}
											<svg
												class="-ml-1 mr-3 h-5 w-5 animate-spin text-white"
												xmlns="http://www.w3.org/2000/svg"
												fill="none"
												viewBox="0 0 24 24"
											>
												<circle
													class="opacity-25"
													cx="12"
													cy="12"
													r="10"
													stroke="currentColor"
													stroke-width="4"
												></circle>
												<path
													class="opacity-75"
													fill="currentColor"
													d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
												></path>
											</svg>
											Creating...
										{:else}
											Complete
										{/if}
									</Button>
								</div>
							</div>
						{/if}
					</form>

					<div class="mt-8 text-center text-sm font-medium text-slate-600">
						Already have an account?{' '}
						<a
							href="/login"
							class="text-base font-bold text-emerald-700 transition-colors hover:text-emerald-600"
						>
							Sign in here
						</a>
					</div>
				</CardContent>
			</Card>
		</div>
	</div>
</div>

<style>
	@keyframes pulseGlow {
		0%,
		100% {
			transform: scale(1);
			opacity: 0.35;
			filter: blur(70px);
		}
		50% {
			transform: scale(1.08);
			opacity: 0.55;
			filter: blur(90px);
		}
	}
	.animate-pulseGlow {
		animation: pulseGlow 12s ease-in-out infinite;
	}
</style>
