<script lang="ts">
	import { page } from '$app/stores';
</script>

<div class="flex min-h-[60vh] items-center justify-center p-4">
	<div
		class="liquid-panel w-full max-w-md rounded-3xl p-10 text-center"
		role="alert"
	>
		<h1 class="mb-2 text-5xl font-bold text-slate-800 dark:text-slate-100">
			{$page.status}
		</h1>
		<p class="mb-6 text-slate-600 dark:text-slate-400">
			{$page.error?.message || 'Reports encountered an error'}
		</p>
		<div class="flex justify-center gap-3">
			<a
				href="/reports"
				class="glass-button inline-flex items-center gap-2 rounded-2xl px-5 py-2.5 text-sm font-medium text-slate-700 transition hover:scale-105 dark:text-slate-200"
			>
				Retry
			</a>
			<a
				href="/"
				class="glass-button inline-flex items-center gap-2 rounded-2xl px-5 py-2.5 text-sm font-medium text-slate-700 transition hover:scale-105 dark:text-slate-200"
			>
				Go Home
			</a>
		</div>
	</div>
</div>
