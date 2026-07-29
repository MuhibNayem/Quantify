<script lang="ts">
	import { timeAgo, formatDate } from '$lib/utils';

	let { date }: { date: string | Date } = $props();
</script>

<time datetime={String(date)} title={formatDate(date)}>
	{timeAgo(date)}
</time>
