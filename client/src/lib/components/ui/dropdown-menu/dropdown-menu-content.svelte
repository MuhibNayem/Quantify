<script lang="ts">
	import { cn } from '$lib/utils.js';
	import { DropdownMenu as DropdownMenuPrimitive } from 'bits-ui';

	let {
		ref = $bindable(null),
		sideOffset = 4,
		portalProps,
		class: className,
		...restProps
	}: DropdownMenuPrimitive.ContentProps & {
		portalProps?: DropdownMenuPrimitive.PortalProps;
	} = $props();
</script>

<DropdownMenuPrimitive.Portal {...portalProps}>
	<DropdownMenuPrimitive.Content
		bind:ref
		data-slot="dropdown-menu-content"
		{sideOffset}
		class={cn(
			'data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 z-50 min-w-[8rem] rounded-xl border border-white/50 bg-gradient-to-br from-white/80 to-white/40 p-1.5 shadow-[inset_1px_1px_2px_0_rgba(255,255,255,0.8),inset_-1px_-1px_2px_0_rgba(255,255,255,0.2),0_20px_40px_-10px_rgba(0,0,0,0.1)] outline-none backdrop-blur-3xl backdrop-saturate-[180%]',
			className
		)}
		{...restProps}
	/>
</DropdownMenuPrimitive.Portal>
