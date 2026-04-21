<script lang="ts">
  import { IconArrowLeft, IconCheck } from "@tabler/icons-svelte";

  import { createTicket } from "$lib/api/tickets.remote";
  import { severitySchema } from "$lib/tickets";

  import { Button } from "$lib/components/ui/button";
  import * as Field from "$lib/components/ui/field";
  import * as NativeSelect from "$lib/components/ui/native-select";
  import { Input } from "$lib/components/ui/input";
  import { Textarea } from "$lib/components/ui/textarea";
  import { H3 } from "$lib/components/ui/typography";
</script>

<div class="w-full max-w-xl mx-auto py-8">
  <div class="mb-8 flex items-center gap-4">
    <Button variant="ghost" size="icon-sm" href="/" title="Back to Dashboard">
      <IconArrowLeft class="size-5" />
    </Button>

    <H3>Report Vulnerability</H3>
  </div>

  <form {...createTicket}>
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
              {...createTicket.fields.title.as("text")}
            />
            <Field.Error errors={createTicket.fields.title.issues()} />
          </Field.Field>

          <Field.Field>
            <Field.Label for="description">Description</Field.Label>
            <Textarea
              id="description"
              placeholder="Describe the vulnerability, steps to reproduce, and potential impact..."
              class="min-h-48"
              {...createTicket.fields.description.as("text")}
            />
            <Field.Error errors={createTicket.fields.description.issues()} />
          </Field.Field>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
            <Field.Field>
              <Field.Label for="severity">Severity Rating</Field.Label>
              <NativeSelect.Root
                id="severity"
                class="w-full"
                {...createTicket.fields.severity.as("select")}
              >
                {#each severitySchema.options as opt}
                  <NativeSelect.Option value={opt}>{opt}</NativeSelect.Option>
                {/each}
              </NativeSelect.Root>
              <Field.Description
                >Assess the potential impact of this report.</Field.Description
              >
              <Field.Error errors={createTicket.fields.severity.issues()} />
            </Field.Field>
          </div>
        </Field.Group>

        <Field.Separator />

        <div class="flex items-center gap-3">
          <Button
            type="submit"
            aria-busy={!!createTicket.pending}
            class="flex-1 sm:flex-none sm:min-w-32"
          >
            <IconCheck class="mr-2 size-4" />
            Submit Report
          </Button>

          <Button variant="ghost" href="/" class="flex-1 sm:flex-none">Cancel</Button>
        </div>
      </Field.Group>
    </Field.Group>
  </form>
</div>
