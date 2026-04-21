import { action, useSubmission, useSearchParams, redirect } from "@solidjs/router";

import { Button, Form, FormAlert, Input, Link } from "../components";

const submit = action(async (formData: FormData) => {
  const payload = Object.fromEntries(formData.entries());

  const res = await fetch("/reset-password", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  const json = await res.json();
  if (!res.ok) {
    throw new Error(json.detail || "Failed to reset password");
  }

  throw redirect("/login?message=Password reset successful");
}, "reset-password");

export function ResetPassword() {
  const submission = useSubmission(submit);
  const [searchParams] = useSearchParams();

  return (
    <Form method="post" action={submit}>
      {submission.error && <FormAlert message={submission.error.message} />}

      <input type="hidden" name="token" value={searchParams.token || ""} />

      <Input
        type="password"
        name="password"
        required
        placeholder="New Password"
        autocomplete="new-password"
      />

      <Button type="submit" disabled={submission.pending || !searchParams.token}>
        {submission.pending ? "Resetting..." : "Reset Password"}
      </Button>

      <div class="flex gap-4 text-sm items-center justify-center">
        <span>
          Remember your password? <Link href="/login">Login</Link>
        </span>
      </div>
    </Form>
  );
}
