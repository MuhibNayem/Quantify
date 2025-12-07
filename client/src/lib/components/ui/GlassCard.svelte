<script lang="ts">
  import { cn } from "$lib/utils";

  let { class: className, children, ...rest } = $props();
</script>

<div
  class={cn(
    "rounded-xl border border-white/20 bg-white/10 p-6 shadow-xl backdrop-blur-lg backdrop-saturate-150 dark:border-white/10 dark:bg-black/20",
    className
  )}
  {...rest}
>
  {@render children()}
</div>
