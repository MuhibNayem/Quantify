<script lang="ts">
	import { NavigationMenu as NavigationMenuPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		class: className,
		...restProps
	}: NavigationMenuPrimitive.ViewportProps = $props();
</script>

<div class={cn('absolute left-0 top-full isolate z-50 flex justify-center')}>
	<NavigationMenuPrimitive.Viewport
		bind:ref
		data-slot="navigation-menu-viewport"
		class={cn(
			'origin-top-center data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-90 relative mt-1.5 h-[var(--bits-navigation-menu-viewport-height)] w-full overflow-hidden rounded-xl border border-white/60 bg-white/90 shadow-[0_20px_40px_-5px_rgba(0,0,0,0.1),inset_0_1px_0_0_rgba(255,255,255,0.8)] backdrop-blur-3xl md:w-[var(--bits-navigation-menu-viewport-width)]',
			className
		)}
		{...restProps}
	/>
</div>
