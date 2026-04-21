import tailwindcss from "@tailwindcss/vite";
import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite-plus";

export default defineConfig({
  fmt: {
    sortImports: true,
    sortTailwindcss: true,
  },
  lint: { options: { typeAware: true, typeCheck: true } },

  plugins: [tailwindcss(), sveltekit()],
});
