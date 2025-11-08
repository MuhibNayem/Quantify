<!-- client/src/lib/components/ui/sonner/sonner.svelte -->
<script lang="ts">
  import { Toaster as SvelteToaster } from "svelte-sonner";
  import { cn } from "$lib/utils";

  type Placement = "top-left" | "top-right" | "bottom-left" | "bottom-right" | "top-center" | "bottom-center";

  const {
    richColors = true,
    closeButton = false,
    position = "bottom-right",
    theme = "system",
    className = "",
  } = $props<{
    richColors?: boolean;
    closeButton?: boolean;
    position?: Placement;
    theme?: "light" | "dark" | "system";
    className?: string;
  }>();
</script>

<SvelteToaster
  {richColors}
  {closeButton}
  {position}
  {theme}
  class={cn("toaster group", className)}
/>
