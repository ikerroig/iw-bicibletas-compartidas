// ÚNICO archivo de rutas compartido (T022): cada spec añade una línea <Route>
// dentro del Layout. No reordenar; añadir al final.
import { Route, Router } from '@solidjs/router';
import Layout from './components/Layout';
import Login from './pages/Login';
import MapaRed from './pages/mapa/MapaRed';
import Desbloqueo from './pages/alquiler/Desbloqueo';

export default function Rutas() {
  return (
    <Router>
      <Route path="/login" component={Login} />
      <Route path="/" component={Layout}>
        <Route path="/" component={() => <MapaRed />} /> {/* spec-001 */}
        <Route path="/alquiler" component={Desbloqueo} /> {/* spec-002 */}
      </Route>
    </Router>
  );
}
