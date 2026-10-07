import { Launcher } from '../platform/Launcher';
import { phone } from '../cases/phone';
import { terminal } from '../cases/terminal';

const implementations = [phone, terminal];
export function App() {
  return <Launcher implementations={implementations} />;
}
