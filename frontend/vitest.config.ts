import { defineConfig } from 'vitest/config';
import solid from 'vite-plugin-solid';

// Este fichero (no vite.config.ts) es la config de Vitest: incluye los tests
// de las specs-001/002 (tests/) y los colocados de las specs-003/004/007/008
// (src/**). hot:false porque @solid-refresh no es resoluble en el entorno de
// tests (spec-008).
export default defineConfig({
  plugins: [solid({ hot: false })],
  resolve: {
    conditions: ['development', 'browser'],
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./tests/setup.ts', './src/test/setup.ts'],
    include: ['tests/**/*.test.{ts,tsx}', 'src/**/*.test.{ts,tsx}'],
    server: { deps: { inline: [/solid-js/, /@solidjs\/router/] } },
  },
});
