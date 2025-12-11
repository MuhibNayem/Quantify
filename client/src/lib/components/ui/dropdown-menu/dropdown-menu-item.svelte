<script lang="ts">
	import { cn } from '$lib/utils.js';
	import { DropdownMenu as DropdownMenuPrimitive } from 'bits-ui';

	let {
		ref = $bindable(null),
		class: className,
		inset,
		variant = 'default',
		...restProps
	}: DropdownMenuPrimitive.ItemProps & {
		inset?: boolean;
		variant?: 'default' | 'destructive';
	} = $props();
</script>

<DropdownMenuPrimitive.Item
	bind:ref
	data-slot="dropdown-menu-item"
	data-inset={inset}
	data-variant={variant}
	class={cn(
		"data-highlighted:bg-blue-500/10 data-highlighted:text-blue-900 data-[variant=destructive]:data-highlighted:bg-rose-500/10 data-[variant=destructive]:data-highlighted:text-rose-900 relative flex cursor-default select-none items-center gap-2 rounded-lg px-3 py-2 text-sm font-bold text-slate-700 outline-none transition-colors data-[disabled]:pointer-events-none data-[variant=destructive]:text-rose-600 data-[disabled]:opacity-50 [&_svg:not([class*='size-'])]:size-4 [&_svg]:pointer-events-none [&_svg]:shrink-0",
		className
	)}
	{...restProps}
/>
