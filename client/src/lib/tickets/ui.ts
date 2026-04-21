import {
  IconCircleCheck,
  IconClock,
  IconCircleDashed,
  IconAlertTriangle,
  IconAlertCircle,
  IconInfoCircle,
} from "@tabler/icons-svelte";

export const SEVERITY_CONFIG = {
  HIGH: {
    label: "High",
    variant: "destructive" as const,
    icon: IconAlertTriangle,
    color: "text-destructive",
  },
  MED: {
    label: "Medium",
    variant: "default" as const,
    icon: IconAlertCircle,
    color: "text-primary",
  },
  LOW: {
    label: "Low",
    variant: "secondary" as const,
    icon: IconInfoCircle,
    color: "text-muted-foreground",
  },
} as const;

export const STATUS_CONFIG = {
  OPEN: {
    label: "Open",
    variant: "outline" as const,
    icon: IconCircleDashed,
  },
  IN_PROGRESS: {
    label: "In Progress",
    variant: "default" as const,
    icon: IconClock,
  },
  RESOLVED: {
    label: "Resolved",
    variant: "secondary" as const,
    icon: IconCircleCheck,
  },
} as const;
