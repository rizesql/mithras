import { type Handle, redirect } from "@sveltejs/kit";
import { createAuthClient, setTokens } from "$lib/auth.server";
import { subjects } from "$lib/subject";
import type { Role } from "$lib/authz";

export const handle: Handle = async ({ event, resolve }) => {
  if (event.url.pathname === "/callback") {
    return resolve(event);
  }

  const client = createAuthClient(event);

  try {
    const accessToken = event.cookies.get("access_token");
    const refreshToken = event.cookies.get("refresh_token");

    if (accessToken) {
      const verified = await client.verify(subjects, accessToken, {
        refresh: refreshToken,
      });

      if (!verified.err) {
        if (verified.tokens) {
          setTokens(event, verified.tokens.access, verified.tokens.refresh);
        }

        event.locals.session = {
          sub: verified.subject.properties.sub,
          roles: verified.subject.properties.roles as Role[],
        };
        return resolve(event);
      }
    }
  } catch (e) {
    console.error("Verification error:", e);
  }

  const { url, challenge } = await client.authorize(`${event.url.origin}/callback`, "code");

  const finalUrl = new URL(url);
  const state = finalUrl.searchParams.get("state");
  finalUrl.searchParams.set("state", `${state}.${challenge.verifier}`);

  return redirect(302, finalUrl.toString());
};
