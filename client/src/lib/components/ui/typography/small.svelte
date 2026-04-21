<script lang="ts">
  import type { HTMLAttributes } from "svelte/elements";
  import { cn, type WithElementRef } from "$lib/utils";

  let {
    ref = $bindable(null),
    class: className,
    children,
    ...restProps
  }: WithElementRef<HTMLAttributes<HTMLElement>> = $props();
</script>

<small
  bind:this={ref}
  class={cn("text-sm leading-none font-medium", className)}
  {...restProps}
>
  {@render children?.()}
</small>
