<script lang="ts">
  import { Check, ChevronsUpDown } from "lucide-svelte";
  import { cn } from "$lib/utils";
  import { Button } from "$lib/components/ui/button";
  import {
    Command,
    CommandEmpty,
    CommandGroup,
    CommandInput,
    CommandItem,
    CommandList,
  } from "$lib/components/ui/command";
  import { Popover, PopoverContent, PopoverTrigger } from "$lib/components/ui/popover";

  let {
    items = [],
    value = $bindable(""),
    placeholder = "Select item...",
    className = "",
  } = $props<{
    items?: { value: string; label: string }[];
    value?: string;
    placeholder?: string;
    className?: string;
  }>();

  let open = $state(false);
</script>

<Popover bind:open>
  <PopoverTrigger asChild let:builder>
    <Button
      variant="outline"
      role="combobox"
      aria-expanded={open}
      class={cn("w-[200px] justify-between", className)}
      use:builder
    >
      {value ? items.find((item) => item.value === value)?.label : placeholder}
      <ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
    </Button>
  </PopoverTrigger>
  <PopoverContent class="w-[200px] p-0">
    <Command>
      <CommandInput placeholder="Search item..." />
      <CommandEmpty>No item found.</CommandEmpty>
      <CommandList>
        <CommandGroup>
          {#each items as item}
            <CommandItem
              value={item.label}
              onSelect={() => {
                value = item.value;
                open = false;
              }}
            >
              <Check
                class={cn(
                  "mr-2 h-4 w-4",
                  value === item.value ? "opacity-100" : "opacity-0"
                )}
              />
              {item.label}
            </CommandItem>
          {/each}
        </CommandGroup>
      </CommandList>
    </Command>
  </PopoverContent>
</Popover>
