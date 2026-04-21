import { getTickets } from "$lib/api/tickets.remote";

export const load = async () => {
  return { tickets: await getTickets() };
};
