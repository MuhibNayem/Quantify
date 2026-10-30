<script lang="ts">
	import { page } from '$app/stores';
</script>

<div class="flex min-h-screen items-center justify-center p-4">
	<div class="panel w-full max-w-md p-10 text-center" role="alert">
		<h1 class="kpi-num mb-2 font-display text-6xl text-foreground">
			{$page.status}
		</h1>
		<p class="mb-8 text-lg text-muted-foreground">
			{$page.error?.message || 'Something went wrong'}
		</p>
		<a href="/" class="btn-soothe inline-flex items-center gap-2 px-6 py-3 text-sm"> Go Home </a>
	</div>
</div>
