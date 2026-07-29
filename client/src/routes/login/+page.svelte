<!-- client/src/routes/login/+page.svelte -->
<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
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
	import { BarChart3, Loader2 } from 'lucide-svelte';

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
					const { accessToken, refreshToken, csrfToken, user, permissions } = response.data;
					auth.login(accessToken, refreshToken, csrfToken, user, permissions || []);
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

	const inputClass =
		'liquid-input h-10 w-full rounded-md bg-card bg-none px-3 text-sm font-normal text-foreground shadow-none ring-0 backdrop-blur-none backdrop-saturate-100 from-card to-card placeholder:text-muted-foreground focus:scale-100 focus:border-primary focus:ring-0 md:text-sm';
	const labelClass = 'text-[0.6875rem] font-semibold tracking-[0.08em] text-muted-foreground uppercase';
	const submitClass =
		'h-10 w-full rounded-md bg-primary bg-none px-4 text-sm font-medium text-primary-foreground shadow-none hover:bg-primary/90 hover:bg-none';
</script>

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
				{#each ['48,210.75', '12,086.00', '3,941.50', '128,400.25', '7,212.80'] as n}
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
				{$t('auth.login.heroTitle')}
			</h1>
			<p class="mt-4 max-w-sm text-[0.9375rem] leading-relaxed text-primary-foreground/70">
				{$t('auth.login.heroSubtitle')}
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
					{ w: 96, qty: '1,204', val: '8,412.00' },
					{ w: 56, qty: '386', val: '2,093.50' },
					{ w: 120, qty: '12,480', val: '86,140.75' },
					{ w: 40, qty: '57', val: '341.10' }
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

	<!-- Right panel — sign-in form -->
	<div class="relative flex flex-1 items-center justify-center px-6 py-12 sm:px-10">
		<div class="w-full max-w-sm" in:fade={{ duration: 400 }}>
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
				<h2
					class="font-display text-[1.625rem] font-semibold tracking-tight text-foreground"
				>
					{$t('auth.login.title')}
				</h2>
				<p class="mt-1.5 text-sm text-muted-foreground">{$t('auth.login.subtitle')}</p>
			</header>

			<form method="POST" use:enhance class="mt-8 space-y-5">
				<div class="space-y-1.5">
					<Label for="username" class={labelClass}>
						{$t('auth.login.username')}
					</Label>
					<Input
						id="username"
						type="text"
						placeholder={$t('auth.login.usernamePlaceholder')}
						bind:value={$form.username}
						class={inputClass}
					/>
					{#if $errors.username}
						<p class="text-xs font-medium text-danger">{$errors.username}</p>
					{/if}
				</div>

				<div class="space-y-1.5">
					<div class="flex items-center justify-between">
						<Label for="password" class={labelClass}>
							{$t('auth.login.password')}
						</Label>
						<a
							href="#"
							class="text-xs font-medium text-primary underline-offset-4 transition-colors hover:underline"
							>{$t('auth.login.forgotPassword')}</a
						>
					</div>
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
				</div>

				<Button type="submit" class={submitClass} disabled={$submitting}>
					{#if $submitting}
						<Loader2 class="h-4 w-4 animate-spin" />
						{$t('auth.login.submitting')}
					{:else}
						{$t('auth.login.submit')}
					{/if}
				</Button>
			</form>

			<p class="mt-6 text-center text-sm text-muted-foreground">
				{$t('auth.login.noAccount')}{' '}
				<a
					href="/register"
					class="font-medium text-primary underline-offset-4 transition-colors hover:underline"
				>
					{$t('auth.login.register')}
				</a>
			</p>
		</div>
	</div>
</div>
