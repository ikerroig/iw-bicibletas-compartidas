// ÚNICO archivo de rutas compartido (T022): cada spec añade una línea <Route>
// dentro del Layout. No reordenar; añadir al final.
import { Route, Router } from '@solidjs/router';
import Layout from './components/Layout';
import Login from './pages/Login';
import MapaRed from './pages/mapa/MapaRed';
import Desbloqueo from './pages/alquiler/Desbloqueo';
import { Devolucion } from './pages/devolucion';
import { MisViajes } from './pages/viajes';
import { Balanceo } from './pages/balanceo';
import { PanelAnalitica } from './pages/analitica';
import { ReportarAveria } from './pages/averias';
import { PanelIncidencias } from './pages/incidencias';

export default function Rutas() {
  return (
    <Router>
      <Route path="/login" component={Login} />
      <Route path="/" component={Layout}>
        <Route path="/" component={() => <MapaRed />} /> {/* spec-001 */}
        <Route path="/alquiler" component={Desbloqueo} /> {/* spec-002 */}
        <Route path="/devolucion" component={Devolucion} /> {/* spec-003 */}
        <Route path="/viajes" component={MisViajes} /> {/* spec-004 */}
        <Route path="/averias" component={ReportarAveria} /> {/* spec-005 */}
        <Route path="/incidencias" component={PanelIncidencias} /> {/* spec-005 */}
        <Route path="/balanceo" component={Balanceo} /> {/* spec-007 */}
        <Route path="/analitica" component={PanelAnalitica} /> {/* spec-008 */}
      </Route>
    </Router>
  );
}
