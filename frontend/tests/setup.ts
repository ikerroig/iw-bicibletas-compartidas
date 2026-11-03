// Configuración global de Vitest (T023): jest-dom, MSW y limpieza entre tests.
import '@testing-library/jest-dom/vitest';
import { cleanup } from '@solidjs/testing-library';
import { afterAll, afterEach, beforeAll, vi } from 'vitest';
import { server } from './server';

/**
 * Solo tests: el AbortSignal de jsdom no es una instancia del AbortSignal de
 * Node, y el fetch nativo (que intercepta MSW) lo rechaza con TypeError. Este
 * envoltorio quita la señal y reproduce su efecto: rechaza con AbortError al
 * abortar. En el navegador real no hace falta (señal y fetch son del mismo DOM).
 */
function fetchCompatibleConJsdom(fetchReal: typeof fetch): typeof fetch {
  return (input, init) => {
    const signal = init?.signal;
    if (!signal) return fetchReal(input, init);
    const { signal: _omitida, ...resto } = init;
    void _omitida;
    return new Promise<Response>((resolve, reject) => {
      const abortar = () => reject(new DOMException('Petición cancelada', 'AbortError'));
      if (signal.aborted) return abortar();
      signal.addEventListener('abort', abortar, { once: true });
      fetchReal(input, resto).then(resolve, reject);
    });
  };
}

beforeAll(() => {
  server.listen({ onUnhandledRequest: 'error' });
  // Después de listen: MSW ya ha sustituido fetch; envolvemos su versión.
  globalThis.fetch = fetchCompatibleConJsdom(globalThis.fetch);
});
afterEach(() => {
  cleanup();
  server.resetHandlers();
  vi.useRealTimers();
  try {
    localStorage.clear();
  } catch {
    /* sin almacenamiento */
  }
});
afterAll(() => server.close());
