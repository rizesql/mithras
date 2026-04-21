import { redirect } from "@sveltejs/kit";
import { createAuthClient, setTokens } from "$lib/auth.server";

export async function GET(event) {
  const code = event.url.searchParams.get("code");
  const state = event.url.searchParams.get("state") || "";
  const authClient = createAuthClient(event);

  // let verifier = getVerifier(event);
  // if (!verifier && state.includes(".")) {
  //   verifier = state.split(".").pop();
  // }
  const verifier = state.includes(".") ? state.split(".").pop() : undefined;

  const tokens = await authClient.exchange(code!, `${event.url.origin}/callback`, verifier);
  if (!tokens.err) {
    setTokens(event, tokens.tokens.access, tokens.tokens.refresh);
    // deleteVerifier(event);
  } else {
    throw tokens.err;
  }

  return redirect(302, `/`);
}
