import { action, useSubmission } from "@solidjs/router";

import { Button, Form, FormAlert, Input, Link } from "../components";

const submit = action(async (formData: FormData) => {
  const payload = Object.fromEntries(formData.entries());

  const res = await fetch("/forgot-password", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  const json = await res.json();
  if (!res.ok) {
    throw new Error(json.detail || "Failed to process request");
  }

  return { ok: true };
}, "forgot-password");

export function ForgotPassword() {
  const submission = useSubmission(submit);

  return (
    <Form method="post" action={submit}>
      {submission.error && <FormAlert message={submission.error.message} />}
      {submission.result?.ok && (
        <div class="h-10 flex items-center px-4 rounded-md bg-accent-a3 text-accent-11 text-left text-sm gap-2">
          Password reset link has been sent to your email.
        </div>
      )}

      <Input type="email" name="email" required placeholder="Email" />

      <Button type="submit" disabled={submission.pending}>
        {submission.pending ? "Sending..." : "Send Reset Link"}
      </Button>

      <div class="flex gap-4 text-sm items-center justify-center has-nth-2:justify-between">
        <span>
          Remember your password? <Link href="/login">Login</Link>
        </span>
      </div>
    </Form>
  );
}
