<script lang="ts">
	import { page } from '$app/stores';
</script>

<div class="flex min-h-[60vh] items-center justify-center p-6">
	<div class="panel w-full max-w-md p-10 text-center" role="alert">
		<p class="mb-2 overline">POS</p>
		<h1 class="kpi-num mb-2 text-5xl text-foreground">
			{$page.status}
		</h1>
		<p class="mb-6 text-muted-foreground">
			{$page.error?.message || 'POS encountered an error'}
		</p>
		<div class="flex justify-center gap-3">
			<a
				href="/pos"
				class="btn-soothe inline-flex h-10 items-center gap-2 rounded-lg px-5 text-sm font-semibold"
			>
				Retry
			</a>
			<a
				href="/"
				class="btn-soft inline-flex h-10 items-center gap-2 rounded-lg px-5 text-sm font-medium"
			>
				Go Home
			</a>
		</div>
	</div>
</div>
