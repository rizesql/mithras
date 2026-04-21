<script lang="ts">
  import type { HTMLAttributes } from "svelte/elements";
  import { cn, type WithElementRef } from "$lib/utils";

  let {
    ref = $bindable(null),
    class: className,
    children,
    ...restProps
  }: WithElementRef<HTMLAttributes<HTMLHeadingElement>> = $props();
</script>

<h2
  bind:this={ref}
  class={cn(
    "scroll-m-20 border-b pb-2 text-3xl font-medium tracking-tight transition-colors first:mt-0",
    className,
  )}
  {...restProps}
>
  {@render children?.()}
</h2>
