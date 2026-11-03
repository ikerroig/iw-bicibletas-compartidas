import { defineConfig } from 'vite';
import solid from 'vite-plugin-solid';

// El backend Go escucha en :8080; en desarrollo el frontend llama a /api/v1
// en su propio origen y Vite lo reenvía (sin CORS).
export default defineConfig({
  plugins: [solid()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
});
