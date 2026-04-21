import { getTicket } from "$lib/api/tickets.remote";

export const load = async (evt) => {
  return { ticket: await getTicket(evt.params.id) };
};
