import { createSignal, Show } from 'solid-js';
import { useNavigate } from '@solidjs/router';
import { iniciarSesion } from '../stores/sesion';
import { ApiError } from '../services/api';

export default function Login() {
  const navegar = useNavigate();
  const [email, setEmail] = createSignal('');
  const [password, setPassword] = createSignal('');
  const [error, setError] = createSignal<string | null>(null);
  const [enviando, setEnviando] = createSignal(false);

  const enviar = async (ev: SubmitEvent) => {
    ev.preventDefault();
    setError(null);
    setEnviando(true);
    try {
      await iniciarSesion(email(), password());
      navegar('/', { replace: true });
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'No se pudo iniciar sesión');
    } finally {
      setEnviando(false);
    }
  };

  return (
    <main class="flex min-h-screen items-center justify-center bg-slate-50 px-4">
      <form
        onSubmit={enviar}
        class="w-full max-w-sm space-y-4 rounded-xl bg-white p-6 shadow"
        aria-labelledby="titulo-login"
      >
        <h1 id="titulo-login" class="text-xl font-semibold">
          Bicicletas compartidas
        </h1>
        <label class="block text-sm">
          Email
          <input
            type="email"
            required
            autocomplete="username"
            class="mt-1 w-full rounded border border-slate-300 px-3 py-2"
            value={email()}
            onInput={(e) => setEmail(e.currentTarget.value)}
          />
        </label>
        <label class="block text-sm">
          Contraseña
          <input
            type="password"
            required
            autocomplete="current-password"
            class="mt-1 w-full rounded border border-slate-300 px-3 py-2"
            value={password()}
            onInput={(e) => setPassword(e.currentTarget.value)}
          />
        </label>
        <Show when={error()}>
          <p role="alert" class="text-sm text-red-700">
            {error()}
          </p>
        </Show>
        <button
          type="submit"
          disabled={enviando()}
          class="w-full rounded bg-emerald-700 py-2 font-medium text-white hover:bg-emerald-800 disabled:opacity-60"
        >
          {enviando() ? 'Entrando…' : 'Entrar'}
        </button>
      </form>
    </main>
  );
}
