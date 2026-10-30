<!-- client/src/routes/login/+page.svelte -->
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
		authLabel,
		authLink,
		authMobileTile,
		authOverline,
		authSub,
		authTitle,
		authTrustLine
	} from '$lib/components/auth/styles';
	import api from '$lib/api';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { auth } from '$lib/stores/auth';
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';
	import { superForm } from 'sveltekit-superforms';
	import { zod4 } from 'sveltekit-superforms/adapters';
	import { loginSchema } from '$lib/schemas/auth';
	import { t } from '$lib/i18n';
	import { ArrowRight, BarChart3, KeyRound, Loader2, Lock, User } from 'lucide-svelte';

	const { form, errors, enhance, submitting } = superForm(
		{ username: '', password: '' },
		{
			validators: zod4(loginSchema),
			onSubmit: async ({ cancel }) => {
				cancel();
				const data = $form;
				try {
					const response = await api.post('/users/login', {
						username: data.username,
						password: data.password
					});
					// Session tokens are set as httpOnly cookies by the backend.
					const { user, permissions } = response.data;
					auth.login(user, permissions || []);
					toast.success($t('auth.login.success'), {
						description: $t('auth.login.successDesc')
					});
					goto('/');
				} catch (error: any) {
					const errorMessage = error.response?.data?.error || error.message || 'Login failed';
					toast.error($t('auth.login.failed'), {
						description: errorMessage
					});
				}
			}
		}
	);

	onMount(() => {
		auth.subscribe((state) => {
			if (state.isAuthenticated) {
				goto('/');
			}
		})();
	});
</script>

<div class="flex min-h-screen w-full">
	<BrandPanel title={$t('auth.login.heroTitle')} subtitle={$t('auth.login.heroSubtitle')} />

	<FormPanel maxWidth="max-w-sm">
		<div in:fade={{ duration: 400 }}>
			<!-- Mobile brand -->
			<div class="mb-10 flex items-center gap-3 lg:hidden">
				<div class={authMobileTile}>
					<BarChart3 class="h-5 w-5 text-white" />
				</div>
				<span class="font-display text-xl font-semibold tracking-tight text-foreground">
					Quantify
				</span>
			</div>

			<header data-animate="fade-up">
				<span class={authOverline}>{$t('auth.login.title')}</span>
				<h2 class={authTitle}>
					<span class="text-gradient-hero">{$t('auth.login.heroTitle')}</span>
				</h2>
				<p class={authSub}>{$t('auth.login.subtitle')}</p>
			</header>

			<div class="{authCard} mt-8 p-6 sm:p-7" data-animate="fade-up" style="animation-delay: 120ms">
				<form method="POST" use:enhance class="space-y-5">
					<div class="space-y-1.5">
						<Label for="username" class={authLabel}>
							{$t('auth.login.username')}
						</Label>
						<div class="relative">
							<User class={authIcon} />
							<Input
								id="username"
								type="text"
								placeholder={$t('auth.login.usernamePlaceholder')}
								bind:value={$form.username}
								class={authInput}
							/>
						</div>
						{#if $errors.username}
							<p class={authError}>{$errors.username}</p>
						{/if}
					</div>

					<div class="space-y-1.5">
						<div class="flex items-center justify-between">
							<Label for="password" class={authLabel}>
								{$t('auth.login.password')}
							</Label>
							<a href="#" class="{authLink} text-xs">{$t('auth.login.forgotPassword')}</a>
						</div>
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
					</div>

					<div class="pt-1">
						<Button type="submit" class="{authCta} group" disabled={$submitting}>
							{#if $submitting}
								<Loader2 class="h-4 w-4 animate-spin" />
								{$t('auth.login.submitting')}
							{:else}
								{$t('auth.login.submit')}
								<ArrowRight
									class="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5"
								/>
							{/if}
						</Button>

						<p class={authTrustLine}>
							<KeyRound class="h-3 w-3" />
							{$t('auth.login.securedBy')}
						</p>
					</div>
				</form>
			</div>

			<p
				class="mt-6 text-center text-sm text-muted-foreground"
				data-animate="fade-up"
				style="animation-delay: 220ms"
			>
				{$t('auth.login.noAccount')}{' '}
				<a href="/register" class={authLink}>{$t('auth.login.register')}</a>
			</p>
		</div>
	</FormPanel>
</div>
