import { createSubjects } from "./create-client";
import * as v from "valibot";

export const subjects = createSubjects({
  user: v.object({
    sub: v.string(),
    roles: v.array(v.string()),
  }),
});
