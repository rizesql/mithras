import type { JSX } from "solid-js/jsx-runtime";

const LOGO = `
  ▇▇      ▇▇     M I T H R A S
  ▇▇▇▇  ▇▇▇▇     Identity Provider
  ▇▇▇▇▇▇▇▇▇▇
  ▇▇  ▇▇  ▇▇     Self-contained authentication
 ▇▇▇▇    ▇▇▇▇
`;

export namespace Layout {
  export interface Props {
    children?: JSX.Element;
  }
}

export function Layout(props: Layout.Props) {
  return (
    <main class="p-4 absolute inset-0 flex items-center justify-center flex-col select-none">
      <div class="w-95 flex flex-col gap-6">
        <div class="block whitespace-pre font-mono text-sm mx-auto">{LOGO}</div>

        {props.children}

        {/*<footer class="mx-auto ">
          <small class="text-xs leading-normal text-surface-11 m-0">
            © {new Date().getFullYear()} Mithras app. All rights reserved.
          </small>
        </footer>*/}
      </div>
    </main>
  );
}
