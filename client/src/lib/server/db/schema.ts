import { randomUUID } from "node:crypto";
import { index, pgEnum, pgTable, serial, text, timestamp, uuid } from "drizzle-orm/pg-core";

export const severityEnum = pgEnum("severity", ["LOW", "MED", "HIGH"]);
export const statusEnum = pgEnum("status", ["OPEN", "IN_PROGRESS", "RESOLVED"]);

export const tickets = pgTable(
  "tickets",
  {
    pk: serial().primaryKey(),
    id: uuid()
      .$defaultFn(() => randomUUID())
      .unique(),

    title: text().notNull(),
    description: text().notNull(),
    severity: severityEnum().notNull(),
    status: statusEnum().notNull(),

    ownerId: text().notNull(),

    createdAt: timestamp().defaultNow().notNull(),
    updatedAt: timestamp()
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (t) => [index("ownerId_idx").on(t.ownerId)],
);
