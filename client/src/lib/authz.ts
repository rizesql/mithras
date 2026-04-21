export type Role = "USER" | "MANAGER" | "ANALYST";

export function canReadAll(roles: Role[]) {
  return roles.some((role) => role === "MANAGER" || role === "ANALYST");
}

export function canAccessTicket(roles: Role[], ownerId: string, userId: string) {
  if (roles.some((role) => role === "MANAGER" || role === "ANALYST")) {
    return true;
  }

  return ownerId === userId;
}

export function canEditTicket(roles: Role[], ownerId: string, userId: string) {
  if (roles.some((role) => role === "MANAGER")) {
    return true;
  }

  return ownerId === userId;
}

export function canRemoveTicket(roles: Role[], ownerId: string, userId: string) {
  if (roles.some((role) => role === "MANAGER")) {
    return true;
  }

  return ownerId === userId;
}
