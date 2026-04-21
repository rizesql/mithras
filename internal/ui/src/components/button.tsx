import type { JSX } from "solid-js/jsx-runtime";

export namespace Button {
  export interface Props extends JSX.ButtonHTMLAttributes<HTMLButtonElement> {}
}
export function Button(props: Button.Props) {
  return (
    <button
      {...props}
      class="h-10 cursor-pointer border-0 font-medium text-md leading-none rounded-sm flex gap-3 items-center justify-center bg-accent-9 text-surface-12"
    />
  );
}
