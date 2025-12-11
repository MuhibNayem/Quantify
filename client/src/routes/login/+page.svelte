<!-- client/src/routes/login/+page.svelte -->
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
	import { auth } from '$lib/stores/auth';
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';

	let username = $state('');
	let password = $state('');
	let loading = $state(false);

	onMount(() => {
		auth.subscribe((state) => {
			if (state.isAuthenticated) {
				goto('/');
			}
		})();
	});

	async function handleLogin() {
		loading = true;
		try {
			const response = await api.post('/users/login', { username, password });
			const { accessToken, refreshToken, csrfToken, user, permissions } = response.data;
			auth.login(accessToken, refreshToken, csrfToken, user, permissions || []);
			toast.success('Login successful!', {
				description: 'Welcome back to your workspace.'
			});
			goto('/');
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || error.message || 'Login failed';
			toast.error('Login Failed', {
				description: errorMessage
			});
		} finally {
			loading = false;
		}
	}
</script>

<div
	class="flex min-h-screen w-full bg-gradient-to-br from-slate-50/90 via-sky-50/70 to-indigo-50/80"
>
	<!-- Left side - Image section (60% width) -->
	<div class="relative hidden overflow-hidden lg:flex lg:flex-[1.5]">
		<!-- Background image covering the entire container -->
		<img
			src="/login-visual-feature.png"
			alt="Inventory and ERP Visualization"
			class="absolute inset-0 h-full w-full object-cover object-center"
		/>

		<!-- Stronger glass overlay for text legibility -->
		<div
			class="absolute inset-0 bg-gradient-to-t from-slate-900/80 via-transparent to-transparent"
		></div>

		<div class="absolute bottom-20 left-12 right-12 max-w-2xl text-white">
			<h2 class="text-5xl font-bold tracking-tight drop-shadow-xl">Orchestrate Your Business</h2>
			<p class="mt-6 text-xl font-medium text-slate-100 drop-shadow-lg">
				Seamlessly connect inventory, POS, and supply chain in one fluid system.
			</p>
		</div>
	</div>

	<!-- Right side - Login form (40% width) -->
	<div class="relative flex flex-1 items-center justify-center overflow-hidden p-8 lg:p-12">
		<!-- Exact Ambient Background from Time Tracking Route -->
		<div class="absolute inset-0 -z-10 opacity-80">
			<div class="absolute -left-36 top-10 h-80 w-80 rounded-full bg-sky-200/70 blur-[120px]"></div>
			<div
				class="absolute right-0 top-0 h-72 w-72 rounded-full bg-emerald-200/60 blur-[110px]"
			></div>
			<div
				class="absolute -bottom-24 left-1/4 h-64 w-64 rounded-full bg-indigo-200/60 blur-[150px]"
			></div>
		</div>

		<div class="relative z-10 w-full max-w-xl space-y-8" in:fade={{ duration: 400 }}>
			<!-- Mobile logo -->
			<div class="mb-8 flex justify-center lg:hidden">
				<div class="flex items-center space-x-3">
					<div
						class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-blue-500 to-indigo-600 shadow-lg shadow-blue-500/30"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-6 w-6 text-white"
							viewBox="0 0 20 20"
							fill="currentColor"
						>
							<path
								fill-rule="evenodd"
								d="M10 2a4 4 0 00-4 4v1H5a1 1 0 00-.994.89l-1 9A1 1 0 004 18h12a1 1 0 00.994-1.11l-1-9A1 1 0 0015 7h-1V6a4 4 0 00-4-4zm2 5V6a2 2 0 10-4 0v1h4zm-6 3a1 1 0 112 0 1 1 0 01-2 0zm7-1a1 1 0 100 2 1 1 0 000-2z"
								clip-rule="evenodd"
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
			<!-- 
				Properties for "Liquid Physics":
				- backdrop-saturate-150: Refracted colors pop
				- border-t-white/80 border-l-white/60: Directional light source (top-left)
				- inset ring/shadow: Specular highlights on edges
			-->
			<Card
				class="relative overflow-hidden rounded-[40px] border border-white/20 bg-gradient-to-br from-white/70 to-white/30 p-0 shadow-[0_40px_100px_-20px_rgba(50,60,90,0.2),inset_0_1px_0_0_rgba(255,255,255,0.8),inset_0_-2px_5px_0_rgba(0,0,0,0.05)] backdrop-blur-3xl backdrop-saturate-[180%]"
			>
				<!-- Light flare overlay -->
				<div
					class="pointer-events-none absolute -left-1/2 -top-1/2 h-full w-full rotate-12 bg-gradient-to-b from-white/20 to-transparent blur-3xl"
				></div>

				<CardHeader class="relative pb-2 pt-12 text-center">
					<CardTitle class="text-4xl font-bold tracking-tight text-slate-900">
						Welcome Back
					</CardTitle>
					<CardDescription class="text-lg font-medium text-slate-600">
						Sign in to access your dashboard
					</CardDescription>
				</CardHeader>
				<CardContent class="relative p-10 pt-8">
					<form
						onsubmit={(event) => {
							event.preventDefault();
							handleLogin();
						}}
						class="space-y-6"
					>
						<div class="space-y-6">
							<div class="space-y-2">
								<Label for="username" class="ml-1 text-sm font-bold tracking-wide text-slate-800">
									Username
								</Label>
								<!-- Liquid Input: Looks like a cutout filled with denser liquid -->
								<Input
									id="username"
									type="text"
									placeholder="Enter your username"
									bind:value={username}
									required
									class="h-14 rounded-2xl border border-white/60 bg-gradient-to-br from-white/90 to-white/60 px-5 text-lg font-bold text-slate-900 shadow-[inset_0_2px_4px_rgba(255,255,255,0.9),0_4px_10px_rgba(0,0,0,0.05)] ring-1 ring-white/40 backdrop-blur-xl backdrop-saturate-[180%] transition-all placeholder:text-slate-500/80 focus:scale-[1.01] focus:border-blue-500/30 focus:ring-4 focus:ring-blue-500/10"
								/>
							</div>
							<div class="space-y-2">
								<div class="flex items-center justify-between">
									<Label for="password" class="ml-1 text-sm font-bold tracking-wide text-slate-800">
										Password
									</Label>
									<a
										href="#"
										class="text-sm font-semibold text-blue-700 transition-colors hover:text-blue-600"
										>Forgot password?</a
									>
								</div>
								<Input
									id="password"
									type="password"
									placeholder="••••••••"
									bind:value={password}
									required
									class="h-14 rounded-2xl border-white/40 bg-white/60 px-5 text-lg font-medium text-slate-900 shadow-[inset_0_2px_4px_rgba(0,0,0,0.05)] backdrop-blur-xl transition-all placeholder:text-slate-500 focus:border-blue-500/50 focus:bg-white/90 focus:ring-4 focus:ring-blue-500/10"
								/>
							</div>
						</div>

						<Button
							type="submit"
							class="glass-button mt-6 h-14 w-full rounded-2xl bg-gradient-to-r from-blue-600 to-indigo-600 text-xl font-bold text-white shadow-[0_20px_40px_-15px_rgba(79,70,229,0.4),inset_0_1px_0_0_rgba(255,255,255,0.4)] transition-all hover:bg-gradient-to-r hover:from-blue-500 hover:to-indigo-500 active:scale-[0.98] disabled:opacity-70"
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
								Signing in...
							{:else}
								Sign In
							{/if}
						</Button>
					</form>

					<div class="mt-8 text-center text-sm font-medium text-slate-600">
						Don't have an account?{' '}
						<a
							href="/register"
							class="text-base font-bold text-blue-700 transition-colors hover:text-blue-600"
						>
							Create one here
						</a>
					</div>
				</CardContent>
			</Card>

			<div class="mt-8 flex justify-center gap-6 opacity-80">
				<!-- Trust indicators -->
				<p class="text-xs font-semibold uppercase tracking-wider text-slate-500">
					Secured by Industry Standard Text Encryption
				</p>
			</div>
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
