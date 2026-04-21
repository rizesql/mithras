import * as v from "valibot";

export const severitySchema = v.picklist(["LOW", "MED", "HIGH"]);
export const statusSchema = v.picklist(["OPEN", "IN_PROGRESS", "RESOLVED"]);

export const createTicketSchema = v.object({
  title: v.pipe(v.string(), v.nonEmpty("Title required")),
  description: v.pipe(v.string(), v.nonEmpty("Description required")),
  severity: severitySchema,
});

export const updateTicketSchema = v.object({
  id: v.string(),
  title: v.pipe(v.string(), v.nonEmpty("Title required")),
  description: v.pipe(v.string(), v.nonEmpty("Description required")),
  severity: severitySchema,
  status: statusSchema,
});
