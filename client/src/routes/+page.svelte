<script lang="ts">
  import { IconPlus, IconInbox } from "@tabler/icons-svelte";

  import { formatDate } from "$lib/formatters";
  import { STATUS_CONFIG, SEVERITY_CONFIG } from "$lib/tickets";

  import { Button } from "$lib/components/ui/button";
  import { Badge } from "$lib/components/ui/badge";
  import * as Card from "$lib/components/ui/card";
  import { H3, Small } from "$lib/components/ui/typography";

  const { data } = $props();
</script>

<div class="flex flex-col gap-8">
  <div class="flex justify-between items-center gap-6">
    <div class="flex flex-col items-start">
      <H3>Active Tickets</H3>
      <Small class="text-muted-foreground font-normal">
        Manage and track security vulnerability reports.
      </Small>
    </div>

    <Button href="/new">New ticket</Button>
  </div>

  {#if data.tickets.length > 0}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each data.tickets as ticket (ticket.id)}
        {@const status = STATUS_CONFIG[ticket.status]}
        {@const severity = SEVERITY_CONFIG[ticket.severity]}

        <a href="/{ticket.id}" class="group transition-all h-full">
          <Card.Root>
            <Card.Header>
              <div class="flex items-center justify-between gap-2">
                <Badge variant={status.variant} class="px-2 py-0.5 text-[10px] uppercase">
                  <status.icon class="mr-1 size-3" />
                  {status.label}
                </Badge>

                {#if ticket.severity === "HIGH"}
                  <div class="flex items-center gap-1 text-destructive animate-pulse">
                    <severity.icon class="size-4" />
                    <span class="text-[10px] tracking-tighter uppercase"> High </span>
                  </div>
                {/if}
              </div>

              <Card.Title class="line-clamp-2 leading-snug">
                {ticket.title}
              </Card.Title>
            </Card.Header>

            <Card.Content class="pb-6 flex-1">
              <p class="text-muted-foreground text-xs line-clamp-3">
                {ticket.description}
              </p>
            </Card.Content>

            <Card.Footer
              class="bg-muted/50 flex items-center justify-between border-t border-border/40"
            >
              <Badge
                variant={severity.variant}
                class="text-[9px] h-4.5 px-2 uppercase tracking-widest"
              >
                {severity.label}
              </Badge>

              <div class="flex flex-col items-end">
                <span class="text-[10px] text-muted-foreground uppercase">Reported</span>
                <span class="text-[11px] font-medium text-foreground">
                  {formatDate(ticket.createdAt)}
                </span>
              </div>
            </Card.Footer>
          </Card.Root>
        </a>
      {/each}
    </div>
  {:else}
    <div
      class="flex flex-col items-center justify-center py-32 px-4 text-center border-2 border-dashed rounded-[2rem] bg-muted/10 border-muted-foreground/20"
    >
      <div class="p-6 bg-muted/50 rounded-full mb-6 ring-8 ring-muted/20">
        <IconInbox class="size-16 text-muted-foreground" />
      </div>
      <h3 class="text-2xl tracking-tight">No tickets found</h3>
      <p class="text-muted-foreground mt-2 max-w-sm text-sm">
        Everything seems clear. Security reports will appear here.
      </p>

      <Button href="/new" variant="outline" class="mt-8">
        <IconPlus class="mr-2 size-4" />
        Create your first ticket
      </Button>
    </div>
  {/if}
</div>
