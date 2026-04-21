import { action, redirect, useSubmission, useSearchParams } from "@solidjs/router";

import { Button, Form, FormAlert, Input, Link } from "../components";

const submit = action(async (formData: FormData) => {
  const payload = Object.fromEntries(formData.entries());

  const res = await fetch("/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  const json = await res.json();
  if (!res.ok) {
    throw new Error(json.detail || "Failed to login");
  }

  throw redirect(json.redirect_url);
}, "login");

export function Login() {
  const submission = useSubmission(submit);
  const [searchParams] = useSearchParams();

  return (
    <Form method="post" action={submit}>
      {submission.error && <FormAlert message={submission.error.message} />}
      {searchParams.message && (
        <div class="h-10 flex items-center px-4 rounded-md bg-success-a3 text-success-11 text-left text-sm gap-2">
          {searchParams.message}
        </div>
      )}

      <Input type="email" name="email" required placeholder="Email" />
      <Input
        type="password"
        name="password"
        required
        placeholder="Password"
        autocomplete="current-password"
      />

      <Button type="submit" disabled={submission.pending}>
        {submission.pending ? "Continuing..." : "Continue"}
      </Button>

      <div class="flex gap-4 text-sm items-center justify-center has-nth-2:justify-between">
        <span>
          Don't have an account? <Link href="/register">Register</Link>
        </span>
        <Link href="/forgot-password">Forgot password?</Link>
      </div>
    </Form>
  );
}
