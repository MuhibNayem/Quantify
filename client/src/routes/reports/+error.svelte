<script lang="ts">
	import { page } from '$app/stores';
</script>

<div class="flex min-h-[60vh] items-center justify-center p-6">
	<div class="panel w-full max-w-md p-8 text-center" role="alert">
		<p class="overline">Reports error</p>
		<h1 class="kpi-num mt-3 text-5xl text-foreground">
			{$page.status}
		</h1>
		<p class="mt-3 text-sm text-muted-foreground">
			{$page.error?.message || 'Reports encountered an error'}
		</p>
		<div class="mt-6 flex justify-center gap-2">
			<a
				href="/reports"
				class="btn-soothe inline-flex items-center justify-center px-5 py-2.5 text-sm"
			>
				Retry
			</a>
			<a href="/" class="btn-soft inline-flex items-center justify-center px-5 py-2.5 text-sm">
				Go Home
			</a>
		</div>
	</div>
</div>
