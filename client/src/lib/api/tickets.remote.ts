import { error, redirect } from "@sveltejs/kit";
import * as v from "valibot";
import { eq } from "drizzle-orm";

import { form, getRequestEvent, query } from "$app/server";
import { db } from "$lib/server/db";
import { tickets } from "$lib/server/db/schema";
import { createTicketSchema, updateTicketSchema } from "$lib/tickets";
import { canAccessTicket, canReadAll, canEditTicket, canRemoveTicket } from "$lib/authz";

function auth() {
  const evt = getRequestEvent();
  if (!evt.locals.session) {
    redirect(307, "/");
  }

  return evt.locals.session;
}

export const getTickets = query(async () => {
  const session = auth();
  return await db.query.tickets.findMany({
    where: (it, op) => (canReadAll(session.roles) ? undefined : op.eq(it.ownerId, session.sub)),
    columns: { pk: false },
  });
});

export const getTicket = query(v.string(), async (id) => {
  const session = auth();
  const ticket = await db.query.tickets.findFirst({
    where: (it, op) => op.eq(it.id, id),
    columns: { pk: false },
  });

  if (!ticket) {
    error(404, "Ticket not found");
  }

  if (!canAccessTicket(session.roles, ticket.ownerId, session.sub)) {
    error(403, "Forbidden");
  }

  return ticket;
});

export const createTicket = form(createTicketSchema, async (ticket) => {
  const session = auth();
  const res = await db
    .insert(tickets)
    .values({ ...ticket, status: "OPEN", ownerId: session.sub })
    .returning({ id: tickets.id });

  redirect(303, `/${res[0].id!}`);
});

export const updateTicket = form(updateTicketSchema, async (ticket) => {
  const session = auth();

  const existing = await getTicket(ticket.id);

  if (session.roles.includes("USER") && ticket.status === "RESOLVED") {
    error(400, "You cannot resolve a ticket");
  }

  if (!canEditTicket(session.roles, existing.ownerId, session.sub)) {
    error(403, "Forbidden");
  }

  await db.update(tickets).set(ticket).where(eq(tickets.id, ticket.id));

  redirect(303, `/${ticket.id}`);
});

export const removeTicket = form(v.object({ id: v.string() }), async ({ id }) => {
  const session = auth();
  const existing = await getTicket(id);

  if (!canRemoveTicket(session.roles, existing.ownerId, session.sub)) {
    error(403, "Forbidden");
  }

  await db.delete(tickets).where(eq(tickets.id, id));

  redirect(303, `/`);
});
