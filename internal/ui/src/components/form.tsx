import type { JSX } from "solid-js/jsx-runtime";

export namespace Form {
  export interface Props extends JSX.FormHTMLAttributes<HTMLFormElement> {}
}
export function Form(props: Form.Props) {
  return <form {...props} class="max-w-full flex flex-col gap-4 m-0" />;
}

export namespace FormAlert {
  export interface Props extends JSX.HTMLAttributes<HTMLDivElement> {
    message: string;
  }
}
export function FormAlert(props: FormAlert.Props) {
  return (
    <div
      class="h-10 flex items-center px-4 rounded-md bg-danger-a3 text-danger-11 text-left text-sm gap-2"
      {...props}
    >
      <span data-slot="message">{props.message}</span>
    </div>
  );
}
