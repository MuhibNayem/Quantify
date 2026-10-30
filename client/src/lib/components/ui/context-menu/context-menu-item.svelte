<script lang="ts">
	import { ContextMenu as ContextMenuPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		class: className,
		inset,
		variant = 'default',
		...restProps
	}: ContextMenuPrimitive.ItemProps & {
		inset?: boolean;
		variant?: 'default' | 'destructive';
	} = $props();
</script>

<ContextMenuPrimitive.Item
	bind:ref
	data-slot="context-menu-item"
	data-inset={inset}
	data-variant={variant}
	class={cn(
		"relative flex cursor-default items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium text-foreground transition-colors outline-none select-none data-highlighted:bg-primary/10 data-highlighted:text-primary data-[disabled]:pointer-events-none data-[disabled]:opacity-50 data-[variant=destructive]:text-danger data-[variant=destructive]:data-highlighted:bg-danger/10 data-[variant=destructive]:data-highlighted:text-danger [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
		className
	)}
	{...restProps}
/>
