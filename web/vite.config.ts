import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig(({ mode }) => {
  const isDemo = mode === "demo";

  return {
    plugins: [react()],
    base: isDemo ? "/demo/" : "/",
    server: {
      port: 3000,
      proxy: isDemo
        ? undefined
        : {
            "/api": {
              target: "http://localhost:8080",
              changeOrigin: true,
              ws: true,
            },
          },
    },
    define: {
      __DEMO_MODE__: JSON.stringify(isDemo),
    },
    build: {
      outDir: isDemo ? "dist-demo" : "dist",
    },
  };
});