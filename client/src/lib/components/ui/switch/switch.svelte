<!-- client/src/lib/components/ui/switch/switch.svelte -->
<script lang="ts">
	import { Switch as SwitchPrimitive } from 'bits-ui';
	import { cn, type WithoutChildrenOrChild } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		class: className,
		checked = $bindable(false),
		...restProps
	}: WithoutChildrenOrChild<SwitchPrimitive.RootProps> = $props();
</script>

<SwitchPrimitive.Root
	bind:ref
	bind:checked
	data-slot="switch"
	class={cn(
		'peer inline-flex h-7 w-12 shrink-0 cursor-pointer items-center rounded-full border border-input bg-muted shadow-xs transition-all focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50',
		'data-[state=checked]:border-primary data-[state=checked]:bg-primary',
		'focus-visible:border-primary focus-visible:ring-ring/30',
		className
	)}
	{...restProps}
>
	<SwitchPrimitive.Thumb
		data-slot="switch-thumb"
		class={cn(
			'pointer-events-none block size-5 rounded-full bg-background shadow-sm ring-0 transition-transform data-[state=checked]:translate-x-[22px] data-[state=unchecked]:translate-x-[2px]'
		)}
	/>
</SwitchPrimitive.Root>
