import { render } from 'solid-js/web';
import 'leaflet/dist/leaflet.css';
import './index.css';
import Rutas from './routes';

render(() => <Rutas />, document.getElementById('root')!);
