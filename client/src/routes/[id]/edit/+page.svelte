<script lang="ts">
  import { IconArrowLeft, IconCheck } from "@tabler/icons-svelte";

  import { updateTicket } from "$lib/api/tickets.remote";
  import { severitySchema, statusSchema } from "$lib/tickets";

  import { Button } from "$lib/components/ui/button";
  import * as Field from "$lib/components/ui/field";
  import * as NativeSelect from "$lib/components/ui/native-select";
  import { Input } from "$lib/components/ui/input";
  import { Textarea } from "$lib/components/ui/textarea";
  import { H3 } from "$lib/components/ui/typography";

  const { params, data } = $props();
</script>

<div class="w-full max-w-xl mx-auto py-8">
  <div class="mb-8 flex items-center gap-4">
    <Button
      variant="ghost"
      size="icon-sm"
      href={`/${params.id}`}
      title="Back to Ticket Detail"
    >
      <IconArrowLeft class="size-5" />
    </Button>

    <H3>Edit Ticket</H3>
  </div>

  <form {...updateTicket}>
    <input {...updateTicket.fields.id.as("hidden", params.id)} />

    <Field.Group>
      <Field.Set>
        <Field.Legend>Vulnerability Details</Field.Legend>
        <Field.Description
          >Provide comprehensive information about the security issue you've discovered.</Field.Description
        >
      </Field.Set>

      <Field.Separator />

      <Field.Group>
        <Field.Group>
          <Field.Field>
            <Field.Label for="title">Title</Field.Label>
            <Input
              id="title"
              placeholder="e.g., SQL Injection in Login Form"
              {...updateTicket.fields.title.as("text", data.ticket.title)}
            />
            <Field.Error errors={updateTicket.fields.title.issues()} />
          </Field.Field>

          <Field.Field>
            <Field.Label for="description">Description</Field.Label>
            <Textarea
              id="description"
              placeholder="Describe the vulnerability, steps to reproduce, and potential impact..."
              class="min-h-48"
              {...updateTicket.fields.description.as("text", data.ticket.description)}
            />
            <Field.Error errors={updateTicket.fields.description.issues()} />
          </Field.Field>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
            <Field.Field>
              <Field.Label for="severity">Severity Rating</Field.Label>
              <NativeSelect.Root
                id="severity"
                class="w-full"
                {...updateTicket.fields.severity.as("select", data.ticket.severity)}
              >
                {#each severitySchema.options as opt}
                  <NativeSelect.Option value={opt}>{opt}</NativeSelect.Option>
                {/each}
              </NativeSelect.Root>
              <Field.Description>
                Assess the potential impact of this report.
              </Field.Description>
              <Field.Error errors={updateTicket.fields.severity.issues()} />
            </Field.Field>

            <Field.Field>
              <Field.Label for="status">Status</Field.Label>
              <NativeSelect.Root
                id="status"
                class="w-full"
                {...updateTicket.fields.status.as("select", data.ticket.status)}
              >
                {#each statusSchema.options as opt}
                  <NativeSelect.Option value={opt}>{opt}</NativeSelect.Option>
                {/each}
              </NativeSelect.Root>
              <Field.Error errors={updateTicket.fields.status.issues()} />
            </Field.Field>
          </div>
        </Field.Group>

        <Field.Separator />

        <div class="flex items-center gap-3">
          <Button
            type="submit"
            aria-busy={!!updateTicket.pending}
            class="flex-1 sm:flex-none sm:min-w-32"
          >
            <IconCheck class="mr-2 size-4" />
            Submit Report
          </Button>

          <Button variant="ghost" href={`/${params.id}`} class="flex-1 sm:flex-none">
            Cancel
          </Button>
        </div>
      </Field.Group>
    </Field.Group>
  </form>
</div>
