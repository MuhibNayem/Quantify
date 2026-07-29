<script lang="ts">
	import { page } from '$app/stores';
</script>

<div class="flex min-h-screen items-center justify-center p-4">
	<div class="liquid-panel w-full max-w-md rounded-lg p-10 text-center" role="alert">
		<h1 class="font-display kpi-num mb-2 text-6xl text-foreground">
			{$page.status}
		</h1>
		<p class="mb-8 text-lg text-muted-foreground">
			{$page.error?.message || 'Something went wrong'}
		</p>
		<a
			href="/"
			class="inline-flex items-center gap-2 rounded-md bg-primary px-6 py-3 text-sm font-medium text-primary-foreground shadow-sm transition hover:opacity-90"
		>
			Go Home
		</a>
	</div>
</div>
