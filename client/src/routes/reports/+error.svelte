<script lang="ts">
	import { page } from '$app/stores';
</script>

<div class="flex min-h-[60vh] items-center justify-center p-4">
	<div class="liquid-panel w-full max-w-md rounded-lg p-10 text-center" role="alert">
		<h1 class="font-display kpi-num mb-2 text-5xl text-foreground">
			{$page.status}
		</h1>
		<p class="mb-6 text-muted-foreground">
			{$page.error?.message || 'Reports encountered an error'}
		</p>
		<div class="flex justify-center gap-3">
			<a
				href="/reports"
				class="inline-flex items-center gap-2 rounded-md bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground shadow-sm transition hover:opacity-90"
			>
				Retry
			</a>
			<a
				href="/"
				class="inline-flex items-center gap-2 rounded-md border border-border bg-card px-5 py-2.5 text-sm font-medium text-foreground shadow-sm transition hover:bg-muted"
			>
				Go Home
			</a>
		</div>
	</div>
</div>
