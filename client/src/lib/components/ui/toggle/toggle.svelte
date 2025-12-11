<script lang="ts" module>
	import { type VariantProps, tv } from 'tailwind-variants';

	export const toggleVariants = tv({
		base: "hover:bg-slate-100/50 hover:text-slate-900 data-[state=on]:bg-gradient-to-r data-[state=on]:from-blue-50 data-[state=on]:to-indigo-50 data-[state=on]:text-blue-700 data-[state=on]:shadow-sm focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 aria-invalid:border-destructive inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-xl text-sm font-bold outline-none transition-[color,box-shadow,transform] focus-visible:ring-[3px] disabled:pointer-events-none disabled:opacity-50 active:scale-[0.98] [&_svg:not([class*='size-'])]:size-5 [&_svg]:pointer-events-none [&_svg]:shrink-0",
		variants: {
			variant: {
				default: 'bg-transparent',
				outline: 'border-white/60 bg-white/40 shadow-sm hover:bg-white/60 text-slate-700 border'
			},
			size: {
				default: 'h-12 min-w-12 px-4',
				sm: 'h-10 min-w-10 px-3',
				lg: 'h-14 min-w-14 px-6'
			}
		},
		defaultVariants: {
			variant: 'default',
			size: 'default'
		}
	});

	export type ToggleVariant = VariantProps<typeof toggleVariants>['variant'];
	export type ToggleSize = VariantProps<typeof toggleVariants>['size'];
	export type ToggleVariants = VariantProps<typeof toggleVariants>;
</script>

<script lang="ts">
	import { Toggle as TogglePrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		pressed = $bindable(false),
		class: className,
		size = 'default',
		variant = 'default',
		...restProps
	}: TogglePrimitive.RootProps & {
		variant?: ToggleVariant;
		size?: ToggleSize;
	} = $props();
</script>

<TogglePrimitive.Root
	bind:ref
	bind:pressed
	data-slot="toggle"
	class={cn(toggleVariants({ variant, size }), className)}
	{...restProps}
/>
