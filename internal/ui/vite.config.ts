import { defineConfig } from "vite-plus";
import solidPlugin from "vite-plugin-solid";
import devtools from "solid-devtools/vite";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  fmt: {
    sortImports: true,
    sortTailwindcss: true,
  },
  lint: { options: { typeAware: true, typeCheck: true } },
  plugins: [devtools(), solidPlugin(), tailwindcss()],
});
