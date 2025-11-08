<script lang="ts">
  import { Calendar as CalendarIcon } from "lucide-svelte";
  import { DateFormatter, getLocalTimeZone, today } from "@internationalized/date";
  import { cn } from "$lib/utils";
  import { Button } from "$lib/components/ui/button";
  import { Calendar } from "$lib/components/ui/calendar";
  import { Popover, PopoverContent, PopoverTrigger } from "$lib/components/ui/popover";

  let {
    value = $bindable<Date | undefined>(undefined),
    placeholder = "Pick a date",
    disabled = false,
    className = "",
  } = $props<{
    value?: Date | undefined;
    placeholder?: string;
    disabled?: boolean;
    className?: string;
  }>();

  let formatter = new DateFormatter("en-US", {
    dateStyle: "long",
  });

  let date = $state(value ? today(getLocalTimeZone()).setFromDate(value) : undefined);

  $effect(() => {
    date = value ? today(getLocalTimeZone()).setFromDate(value) : undefined;
  });

  function handleSelect(e: CustomEvent<any>) {
    value = e.detail.date?.toDate(getLocalTimeZone());
  }
</script>

<Popover>
  <PopoverTrigger asChild let:builder>
    <Button
      variant="outline"
      class={cn(
        "w-[280px] justify-start text-left font-normal",
        !value && "text-muted-foreground",
        className
      )}
      disabled={disabled}
      use:builder
    >
      <CalendarIcon class="mr-2 h-4 w-4" />
      {value ? formatter.format(date) : placeholder}
    </Button>
  </PopoverTrigger>
  <PopoverContent class="w-auto p-0">
    <Calendar
      bind:value={date}
      initialFocus
      onselect={handleSelect}
      minValue={today(getLocalTimeZone())}
    />
  </PopoverContent>
</Popover>
