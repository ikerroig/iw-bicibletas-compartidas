import { setupServer } from 'msw/node';
import { handlers } from './handlers';

/** Servidor MSW compartido por setup.ts y los tests (server.use(...)). */
export const server = setupServer(...handlers);
