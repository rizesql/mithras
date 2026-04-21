<script lang="ts">
  import "./layout.css";
  import favicon from "$lib/assets/favicon.svg";

  import { Button } from "$lib/components/ui/button";
  import { Code, H4, Small } from "$lib/components/ui/typography";

  import { IconShieldSearch } from "@tabler/icons-svelte";
  import { logout } from "$lib/api/auth.remote";

  let { data, children } = $props();
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
  <title>AuthX | Ticket System</title>
</svelte:head>

<div
  class="min-h-svh grid grid-rows-[auto_1fr_auto] mx-auto max-w-5xl px-4 sm:px-6 lg:px-8"
>
  <header class="flex items-center justify-between py-6 border-b mb-8">
    <a href="/" class="flex items-center gap-2 group transition-all">
      <div
        class="p-2 bg-primary/10 rounded-lg group-hover:bg-primary/20 transition-colors text-primary"
      >
        <IconShieldSearch class="size-6" />
      </div>

      <div class="flex flex-col">
        <H4>AuthX</H4>
        <Small class="text-xs text-muted-foreground tracking-wide">
          Security Ticketing
        </Small>
      </div>
    </a>

    <nav class="flex items-center gap-4">
      <div class="hidden sm:flex flex-col items-end mr-2">
        <Code class="text-xs">{data.session.sub}</Code>
        <Small class="text-[10px] uppercase">
          {data.session.roles.join(", ")}
        </Small>
      </div>

      <form {...logout}>
        <Button type="submit" variant="ghost">Logout</Button>
      </form>
    </nav>
  </header>

  <main>
    {@render children?.()}
  </main>

  <footer class="py-4 border-t mt-16 text-center">
    <Small class="text-sm text-muted-foreground/60 font-normal">
      &copy; {new Date().getFullYear()} AuthX Ticketing System. All rights reserved.
    </Small>
  </footer>
</div>
