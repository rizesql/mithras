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

<h3
  bind:this={ref}
  class={cn("scroll-m-20 text-2xl font-medium tracking-tight", className)}
  {...restProps}
>
  {@render children?.()}
</h3>
