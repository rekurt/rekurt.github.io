import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import vue from "@vitejs/plugin-vue";
import { resolve } from "node:path";
export default defineConfig({
  base: "/article-demos/",
  plugins: [react(), vue()],
  build: {
    outDir: "../../public/article-demos",
    emptyOutDir: true,
    rollupOptions: {
      input: {
        depth: resolve("depth/index.html"),
        core: resolve("openkline/core/index.html"),
        react: resolve("openkline/react/index.html"),
        vue: resolve("openkline/vue/index.html"),
      },
    },
  },
});
