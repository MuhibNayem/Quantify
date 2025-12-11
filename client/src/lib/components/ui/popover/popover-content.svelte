<script lang="ts">
	import { cn } from '$lib/utils.js';
	import { Popover as PopoverPrimitive } from 'bits-ui';

	let {
		ref = $bindable(null),
		class: className,
		sideOffset = 4,
		align = 'center',
		portalProps,
		...restProps
	}: PopoverPrimitive.ContentProps & {
		portalProps?: PopoverPrimitive.PortalProps;
	} = $props();
</script>

<PopoverPrimitive.Portal {...portalProps}>
	<PopoverPrimitive.Content
		bind:ref
		data-slot="popover-content"
		{sideOffset}
		{align}
		class={cn(
			'data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 z-50 w-72 rounded-2xl border border-white/50 bg-gradient-to-br from-white/80 to-white/40 p-4 shadow-[inset_1px_1px_2px_0_rgba(255,255,255,0.8),inset_-1px_-1px_2px_0_rgba(255,255,255,0.2),0_40px_80px_-20px_rgba(50,60,90,0.3)] outline-none backdrop-blur-3xl backdrop-saturate-[180%]',
			className
		)}
		{...restProps}
	/>
</PopoverPrimitive.Portal>
