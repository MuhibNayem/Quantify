<script lang="ts">
	import { Tabs as TabsPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	let { ref = $bindable(null), class: className, ...restProps }: TabsPrimitive.ListProps = $props();
</script>

<TabsPrimitive.List
	bind:ref
	data-slot="tabs-list"
	class={cn(
		'inline-flex h-14 w-fit items-center justify-center rounded-2xl bg-slate-100/50 p-1.5 text-slate-500 backdrop-blur-md',
		className
	)}
	{...restProps}
/>
