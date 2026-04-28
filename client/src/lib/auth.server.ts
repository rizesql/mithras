import type { RequestEvent } from "@sveltejs/kit";
import { createClient } from "./create-client";

export function createAuthClient(event: RequestEvent) {
  return createClient({
    clientID: "client-example",
    issuer: "http://localhost:8080",
    fetch: event.fetch,
  });
}

export function setTokens(event: RequestEvent, access: string, refresh: string) {
  event.cookies.set("refresh_token", refresh, {
    httpOnly: false,
    sameSite: "lax",
    path: "/",
    maxAge: 34560000,
    secure: false,
  });

  event.cookies.set("access_token", access, {
    httpOnly: false,
    sameSite: "lax",
    path: "/",
    maxAge: 34560000,
    secure: false,
  });
}
