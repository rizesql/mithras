import type { JSX } from "solid-js/jsx-runtime";

export namespace Link {
  export interface Props extends JSX.AnchorHTMLAttributes<HTMLAnchorElement> {}
}
export function Link(props: Link.Props) {
  return <a {...props} class="underline underline-offset-2 font-semibold" />;
}
