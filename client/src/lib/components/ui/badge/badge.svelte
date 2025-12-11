<script lang="ts" module>
	import { type VariantProps, tv } from 'tailwind-variants';

	export const badgeVariants = tv({
		base: 'focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 aria-invalid:border-destructive inline-flex w-fit shrink-0 items-center justify-center gap-1 overflow-hidden whitespace-nowrap rounded-full border px-2.5 py-0.5 text-xs font-bold transition-all focus-visible:ring-[3px] [&>svg]:pointer-events-none [&>svg]:size-3',
		variants: {
			variant: {
				default:
					'bg-gradient-to-r from-blue-600 to-indigo-600 text-white shadow-sm border-transparent [a&]:hover:bg-gradient-to-r [a&]:hover:from-blue-500 [a&]:hover:to-indigo-500',
				secondary:
					'bg-white/60 text-slate-800 border-white/40 backdrop-blur-md [a&]:hover:bg-white/80',
				destructive: 'bg-red-500/10 text-red-700 border-red-200 [a&]:hover:bg-red-500/20',
				outline:
					'text-slate-700 border-white/60 bg-white/20 backdrop-blur-sm [a&]:hover:bg-white/40'
			}
		},
		defaultVariants: {
			variant: 'default'
		}
	});

	export type BadgeVariant = VariantProps<typeof badgeVariants>['variant'];
</script>

<script lang="ts">
	import type { HTMLAnchorAttributes } from 'svelte/elements';
	import { cn, type WithElementRef } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		href,
		class: className,
		variant = 'default',
		children,
		...restProps
	}: WithElementRef<HTMLAnchorAttributes> & {
		variant?: BadgeVariant;
	} = $props();
</script>

<svelte:element
	this={href ? 'a' : 'span'}
	bind:this={ref}
	data-slot="badge"
	{href}
	class={cn(badgeVariants({ variant }), className)}
	{...restProps}
>
	{@render children?.()}
</svelte:element>
