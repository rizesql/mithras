import { action, redirect, useSubmission } from "@solidjs/router";

import { Button, Form, FormAlert, Input, Link } from "../components";

const submit = action(async (formData: FormData) => {
  const payload = Object.fromEntries(formData.entries());

  const res = await fetch("/register", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  const json = await res.json();
  if (!res.ok) {
    throw new Error(json.detail || "Failed to login");
  }

  throw redirect(json.redirect_url);
}, "register");

export function Register() {
  const submission = useSubmission(submit);

  return (
    <Form method="post" action={submit}>
      {submission.error && <FormAlert message={submission.error.message} />}

      <Input type="text" name="name" required placeholder="Name" />
      <Input type="email" name="email" required placeholder="Email" />
      <Input
        type="password"
        name="password"
        required
        placeholder="Password"
        autocomplete="new-password"
      />

      <Button type="submit">{submission.pending ? "Continuing..." : "Continue"}</Button>

      <div class="flex gap-4 text-sm items-center justify-center has-nth-2:justify-between">
        <span>
          Already have an account? <Link href="/login">Login</Link>
        </span>
      </div>
    </Form>
  );
}
