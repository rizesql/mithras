import type { Role } from "$lib/authz";

// for information about these interfaces
declare global {
  namespace App {
    // interface Error {}
    interface Locals {
      session: {
        sub: string;
        roles: Role[];
      };
    }
    // interface PageData {}
    // interface PageState {}
    // interface Platform {}
  }
}
