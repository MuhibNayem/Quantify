<script lang="ts">
	let { class: className = '', ...restProps } = $props();
</script>

<kbd
	class="inline-block rounded border border-black/10 bg-black/[0.06] px-1.5 py-0.5 font-mono text-[0.65rem] leading-none {className}"
	{...restProps}
>
	<slot />
</kbd>
