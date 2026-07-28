<script lang="ts">
	import { page } from '$app/stores';
</script>

<div class="flex min-h-screen items-center justify-center p-4">
	<div
		class="liquid-panel w-full max-w-md rounded-3xl p-10 text-center"
		role="alert"
	>
		<h1 class="mb-2 text-6xl font-bold text-slate-800 dark:text-slate-100">
			{$page.status}
		</h1>
		<p class="mb-8 text-lg text-slate-600 dark:text-slate-400">
			{$page.error?.message || 'Something went wrong'}
		</p>
		<a
			href="/"
			class="glass-button inline-flex items-center gap-2 rounded-2xl px-6 py-3 text-sm font-medium text-slate-700 transition hover:scale-105 dark:text-slate-200"
		>
			Go Home
		</a>
	</div>
</div>
