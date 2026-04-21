import { form, getRequestEvent } from "$app/server";
import { redirect } from "@sveltejs/kit";

export const logout = form(async () => {
  const evt = getRequestEvent();
  evt.cookies.delete("access_token", { path: "/" });
  evt.cookies.delete("refresh_token", { path: "/" });

  redirect(303, "/");
});
