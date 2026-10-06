import { Launcher } from '../platform/Launcher';
import { phone } from '../cases/phone';

const implementations = [phone];
export function App() {
  return <Launcher implementations={implementations} />;
}
