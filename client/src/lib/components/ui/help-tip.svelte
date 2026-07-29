<script lang="ts">
	import { HelpCircle } from 'lucide-svelte';
	import * as Tooltip from '$lib/components/ui/tooltip';

	let { text }: { text: string } = $props();
</script>

<Tooltip.Provider>
	<Tooltip.Root>
		<Tooltip.Trigger>
			<button type="button" class="help-tip" aria-label="Help">
				<HelpCircle size={14} />
			</button>
		</Tooltip.Trigger>
		<Tooltip.Content side="top" class="help-tip-content">{text}</Tooltip.Content>
	</Tooltip.Root>
</Tooltip.Provider>

<style>
	.help-tip {
		display: inline-flex;
		opacity: 0.5;
		transition: opacity 0.15s;
		background: none;
		border: none;
		cursor: help;
		padding: 0;
	}
	.help-tip:hover {
		opacity: 1;
	}
</style>
