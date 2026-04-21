import adapter from "@sveltejs/adapter-auto";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

/** @type {import('@sveltejs/kit').Config} */
export default {
  preprocess: [vitePreprocess()],

  compilerOptions: {
    experimental: {
      async: true,
    },

    runes: ({ filename }) => (filename.split(/[/\\]/).includes("node_modules") ? undefined : true),
  },

  kit: {
    experimental: {
      remoteFunctions: true,
    },

    adapter: adapter(),

    typescript: {
      config: (config) => ({
        ...config,
        include: [...config.include, "../drizzle.config.ts"],
      }),
    },
  },
};
