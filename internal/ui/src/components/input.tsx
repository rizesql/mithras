import type { JSX } from "solid-js/jsx-runtime";

export namespace Input {
  export interface Props extends JSX.InputHTMLAttributes<HTMLInputElement> {}
}
export function Input(props: Input.Props) {
  return (
    <input
      {...props}
      class="w-full h-10 px-2 border bg-surface-3 rounded-sm text-sm outline-none focus:border-accent user-invalid:not:focus:border-danger-9"
    />
  );
}
