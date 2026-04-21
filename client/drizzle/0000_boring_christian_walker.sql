CREATE TYPE "public"."severity" AS ENUM('LOW', 'MED', 'HIGH');--> statement-breakpoint
CREATE TYPE "public"."status" AS ENUM('OPEN', 'IN_PROGRESS', 'RESOLVED');--> statement-breakpoint
CREATE TABLE "tickets" (
	"pk" serial PRIMARY KEY NOT NULL,
	"id" uuid,
	"title" text NOT NULL,
	"description" text NOT NULL,
	"severity" "severity" NOT NULL,
	"status" "status" NOT NULL,
	"owner_id" text NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "tickets_id_unique" UNIQUE("id")
);
