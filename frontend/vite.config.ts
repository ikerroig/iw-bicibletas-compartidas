import { defineConfig } from "vitest/config";
import solid from "vite-plugin-solid";

export default defineConfig(({ mode }) => ({
  // El módulo virtual @solid-refresh (HMR) no es resoluble en el entorno
  // de tests; se desactiva el modo hot solo en modo test.
  plugins: [solid({ hot: mode !== "test" })],
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["src/test/setup.ts"],
  },
}));
