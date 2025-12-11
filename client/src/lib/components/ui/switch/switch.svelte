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
		'peer inline-flex h-7 w-12 shrink-0 cursor-pointer items-center rounded-full border border-white/40 bg-gradient-to-br from-slate-200/50 to-slate-200/50 shadow-[inset_0_2px_4px_rgba(0,0,0,0.05)] ring-1 ring-white/40 transition-all focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50',
		'data-[state=checked]:border-transparent data-[state=checked]:bg-gradient-to-r data-[state=checked]:from-blue-600 data-[state=checked]:to-indigo-600',
		'focus-visible:border-blue-500/30 focus-visible:ring-blue-500/10',
		className
	)}
	{...restProps}
>
	<SwitchPrimitive.Thumb
		data-slot="switch-thumb"
		class={cn(
			'pointer-events-none block size-5 rounded-full bg-white shadow-[0_2px_4px_rgba(0,0,0,0.1)] ring-0 transition-transform data-[state=checked]:translate-x-[22px] data-[state=unchecked]:translate-x-[2px]'
		)}
	/>
</SwitchPrimitive.Root>
