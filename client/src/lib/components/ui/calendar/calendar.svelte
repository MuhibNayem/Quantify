<script lang="ts">
	import { Calendar as CalendarPrimitive } from 'bits-ui';
	import type { DateValue } from '@internationalized/date';
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import { cn } from '$lib/utils.js';
	import { buttonVariants } from '$lib/components/ui/button';

	type CalendarProps = Omit<Extract<CalendarPrimitive.RootProps, { type: 'single' }>, 'type'> & {
		type?: 'single';
	};

	let {
		ref = $bindable(null),
		class: className,
		value = $bindable(),
		onValueChange,
		type = 'single',
		...restProps
	}: CalendarProps = $props();

	function handleValueChange(next: DateValue | undefined) {
		value = next;
		onValueChange?.(next);
	}
</script>

<CalendarPrimitive.Root
	bind:ref
	{value}
	{type}
	onValueChange={handleValueChange}
	fixedWeeks
	weekdayFormat="short"
	data-slot="calendar"
	class={cn('bg-background text-foreground p-3', className)}
	{...restProps}
>
	{#snippet children({ months, weekdays })}
		<CalendarPrimitive.Header class="flex items-center justify-between">
			<CalendarPrimitive.PrevButton
				class={cn(
					buttonVariants({ variant: 'outline' }),
					'size-8 bg-transparent p-0 opacity-50 hover:opacity-100'
				)}
			>
				<ChevronLeftIcon class="size-4" />
			</CalendarPrimitive.PrevButton>
			<CalendarPrimitive.Heading class="text-sm font-medium" />
			<CalendarPrimitive.NextButton
				class={cn(
					buttonVariants({ variant: 'outline' }),
					'size-8 bg-transparent p-0 opacity-50 hover:opacity-100'
				)}
			>
				<ChevronRightIcon class="size-4" />
			</CalendarPrimitive.NextButton>
		</CalendarPrimitive.Header>
		{#each months as month (month.value.toString())}
			<CalendarPrimitive.Grid class="mt-4 w-full border-collapse select-none">
				<CalendarPrimitive.GridHead>
					<CalendarPrimitive.GridRow class="flex w-full justify-between">
						{#each weekdays as weekday}
							<CalendarPrimitive.HeadCell
								class="text-muted-foreground w-8 rounded-md text-xs font-normal"
							>
								{weekday.slice(0, 2)}
							</CalendarPrimitive.HeadCell>
						{/each}
					</CalendarPrimitive.GridRow>
				</CalendarPrimitive.GridHead>
				<CalendarPrimitive.GridBody>
					{#each month.weeks as weekDates}
						<CalendarPrimitive.GridRow class="mt-1 flex w-full justify-between">
							{#each weekDates as date}
								<CalendarPrimitive.Cell
									{date}
									month={month.value}
									class="p-0 text-center text-sm"
								>
									<CalendarPrimitive.Day
										class="data-selected:bg-primary data-selected:text-primary-foreground hover:bg-accent hover:text-accent-foreground focus-visible:ring-ring inline-flex size-8 items-center justify-center rounded-md text-sm font-normal outline-none data-disabled:pointer-events-none data-disabled:opacity-50 data-unavailable:pointer-events-none data-unavailable:line-through"
									>
										{date.day}
									</CalendarPrimitive.Day>
								</CalendarPrimitive.Cell>
							{/each}
						</CalendarPrimitive.GridRow>
					{/each}
				</CalendarPrimitive.GridBody>
			</CalendarPrimitive.Grid>
		{/each}
	{/snippet}
</CalendarPrimitive.Root>
