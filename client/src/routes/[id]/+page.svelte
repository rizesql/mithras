<script lang="ts">
  import {
    IconArrowLeft,
    IconEdit,
    IconTrash,
    IconCalendar,
    IconUser,
    IconClock,
    IconAlertCircle,
  } from "@tabler/icons-svelte";

  import { removeTicket } from "$lib/api/tickets.remote";
  import { formatDate } from "$lib/formatters";
  import { STATUS_CONFIG, SEVERITY_CONFIG } from "$lib/tickets";
  import { canEditTicket, canRemoveTicket } from "$lib/authz";

  import { Button } from "$lib/components/ui/button";
  import { Badge } from "$lib/components/ui/badge";
  import * as Card from "$lib/components/ui/card";
  import * as Dialog from "$lib/components/ui/dialog";
  import { Code, H3, H4, P } from "$lib/components/ui/typography";

  const { params, data } = $props();

  const status = () => STATUS_CONFIG[data.ticket.status];
  const StatusIcon = () => status().icon;
  const severity = () => SEVERITY_CONFIG[data.ticket.severity];

  const isEditable = $derived(
    canEditTicket(data.session.roles, data.ticket.ownerId, data.session.sub),
  );
  const isRemovable = $derived(
    canRemoveTicket(data.session.roles, data.ticket.ownerId, data.session.sub),
  );
</script>

<div class="flex flex-col gap-6">
  <div class="flex items-center gap-4">
    <Button variant="ghost" size="icon-sm" href="/" title="Back to Tickets">
      <IconArrowLeft class="size-5" />
    </Button>

    <div class="flex flex-col">
      <H3 class="flex items-center gap-3">
        Ticket Detail <Code>#{params.id.slice(0, 8)}</Code>
      </H3>
    </div>
  </div>

  <div class="grid grid-cols-1 lg:grid-cols-3 gap-8 items-start">
    <div class="lg:col-span-2 space-y-8">
      <Card.Root class="border-border/50 shadow-sm">
        <Card.Header class="pb-6 border-b border-border/50">
          <div class="flex flex-wrap items-center gap-3 mb-4">
            <Badge variant={status().variant} class="px-3 py-1 uppercase text-[10px]">
              <StatusIcon class="mr-1.5 size-3.5" />
              {status().label}
            </Badge>

            <Badge variant={severity().variant} class="px-3 py-1 uppercase text-[10px]">
              {severity().label} SEVERITY
            </Badge>
          </div>

          <Card.Title class="text-4xl text-foreground">
            {data.ticket.title}
          </Card.Title>
        </Card.Header>

        <Card.Content class="pt-8 pb-12">
          <H4 class="flex items-center gap-2">
            <span class="h-px w-8 bg-muted-foreground/30"></span>
            Description
          </H4>

          <P class="whitespace-pre-wrap">
            {data.ticket.description}
          </P>
        </Card.Content>
      </Card.Root>
    </div>

    <div class="space-y-6">
      <Card.Root class="bg-muted/20 border-border/30 backdrop-blur-sm">
        <Card.Header>
          <Card.Title class="text-muted-foreground">Metadata</Card.Title>
        </Card.Header>

        <Card.Content class="space-y-6">
          <div class="flex items-start gap-3">
            <div class="p-1.5 bg-background rounded-md shadow-sm border border-border/50">
              <IconUser class="size-3.5 text-primary" />
            </div>

            <div class="flex flex-col">
              <span class="text-[10px] text-muted-foreground uppercase">Reporter</span>
              <span class="text-[11px] font-medium text-foreground">
                {data.ticket.ownerId}
              </span>
            </div>
          </div>

          <div class="flex items-center gap-3">
            <div class="p-1.5 bg-background rounded-md shadow-sm border border-border/50">
              <IconCalendar class="size-3.5 text-primary" />
            </div>

            <div class="flex flex-col">
              <span class="text-[10px] text-muted-foreground uppercase">Created</span>
              <span class="text-[11px] font-medium text-foreground">
                {formatDate(data.ticket.createdAt)}
              </span>
            </div>
          </div>

          {#if data.ticket.updatedAt !== data.ticket.createdAt}
            <div class="flex items-center gap-3">
              <div
                class="p-1.5 bg-background rounded-md shadow-sm border border-border/50"
              >
                <IconClock class="size-3.5 text-primary" />
              </div>
              <div class="flex flex-col">
                <span class="text-[10px] text-muted-foreground uppercase">Modified</span>
                <span class="text-[11px] font-medium text-foreground">
                  {formatDate(data.ticket.updatedAt)}
                </span>
              </div>
            </div>
          {/if}
        </Card.Content>

        <Card.Footer class="flex flex-col gap-3 pt-6">
          {#if isEditable}
            <Button href="/{data.ticket.id}/edit" class="w-full">
              <IconEdit class="mr-2 size-4" />
              Edit Ticket
            </Button>
          {:else}
            <div class="bg-muted/50 p-3 rounded-lg border border-dashed text-center">
              <p
                class="text-[10px] text-muted-foreground uppercase font-bold tracking-wider"
              >
                Read-only access
              </p>
            </div>
          {/if}

          {#if isRemovable}
            <Dialog.Root>
              <Dialog.Trigger>
                {#snippet child({ props })}
                  <Button
                    {...props}
                    variant="destructive"
                    class="w-full bg-destructive/10 hover:bg-destructive/20 text-destructive border-destructive/20 shadow-sm"
                  >
                    <IconTrash class="mr-2 size-4" />
                    Delete Ticket
                  </Button>
                {/snippet}
              </Dialog.Trigger>

              <Dialog.Content class="sm:max-w-106.25">
                <Dialog.Header>
                  <Dialog.Title class="text-xl font-bold flex items-center gap-2">
                    <IconAlertCircle class="text-destructive size-6" />
                    Confirm Deletion
                  </Dialog.Title>
                  <Dialog.Description class="pt-2">
                    This action is permanent. Are you sure you want to delete <span
                      class="font-bold text-foreground">"{data.ticket.title}"</span
                    >?
                  </Dialog.Description>
                </Dialog.Header>
                <Dialog.Footer class="mt-6">
                  <Dialog.Close>
                    {#snippet child({ props })}
                      <Button {...props} variant="ghost">Cancel</Button>
                    {/snippet}
                  </Dialog.Close>
                  <form {...removeTicket}>
                    <input {...removeTicket.fields.id.as("hidden", params.id)} />
                    <Button
                      type="submit"
                      variant="destructive"
                      aria-busy={!!removeTicket.pending}
                    >
                      Delete Permanently
                    </Button>
                  </form>
                </Dialog.Footer>
              </Dialog.Content>
            </Dialog.Root>
          {/if}
        </Card.Footer>
      </Card.Root>
    </div>
  </div>
</div>
