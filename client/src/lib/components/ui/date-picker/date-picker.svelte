<script lang="ts">
  import { Calendar as CalendarIcon } from "lucide-svelte";
  import { DateFormatter, fromDate, getLocalTimeZone, toCalendarDate, today } from "@internationalized/date";
  import type { DateValue } from "@internationalized/date";
  import { cn } from "$lib/utils";
  import { buttonVariants } from "$lib/components/ui/button";
  import { Calendar } from "$lib/components/ui/calendar";
  import { Popover, PopoverContent, PopoverTrigger } from "$lib/components/ui/popover";

  let {
    value = $bindable<Date | undefined>(undefined),
    placeholder = "Pick a date",
    disabled = false,
    className = "",
  }: {
    value?: Date | undefined;
    placeholder?: string;
    disabled?: boolean;
    className?: string;
  } = $props();

  let formatter = new DateFormatter("en-US", {
    dateStyle: "long",
  });

  const timeZone = getLocalTimeZone();

  let date = $derived(value ? toCalendarDate(fromDate(value, timeZone)) : undefined);

  function handleDateChange(newDate: DateValue | undefined) {
    value = newDate ? newDate.toDate(timeZone) : undefined;
  }
</script>

<Popover>
  <PopoverTrigger
    class={cn(
      buttonVariants({ variant: "outline" }),
      "w-[280px] justify-start text-left font-normal",
      !value && "text-muted-foreground",
      className
    )}
    {disabled}
  >
    <CalendarIcon class="mr-2 h-4 w-4" />
    {value && date ? formatter.format(date.toDate(timeZone)) : placeholder}
  </PopoverTrigger>
  <PopoverContent class="w-auto p-0">
    <Calendar
      value={date}
      onValueChange={handleDateChange}
      initialFocus
      minValue={today(timeZone)}
    />
  </PopoverContent>
</Popover>
