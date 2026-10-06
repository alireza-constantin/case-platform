// Local verification input only; never imported by the app or served publicly.
// Keep the generated file under ignored output and never print its contents.
import { readFile, readdir, mkdir, writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
const root = new URL('../../', import.meta.url);
const phone = await readFile(new URL('server/internal/cases/phone/phone.go', root), 'utf8');
const terminal = await readFile(new URL('server/internal/cases/terminal/terminal.go', root), 'utf8');
const phonePassword = phone.match(/\*input\.Password == "([^"]+)"/)?.[1];
const hintBlock = terminal.match(/var publicInstructions = \[\]string\{([^\n]+)\}/)?.[1];
if (!phonePassword || !hintBlock) throw new Error('Server verification definitions changed; update the local input loader.');
const action = terminal.slice(terminal.indexOf('func (Module) Act'), terminal.indexOf('func (Module) Asset'));
const terminalCommands = [...action.matchAll(/case ("[^"]+"(?:, "[^"]+")*):([\s\S]*?)(?=\n\tcase |\n\tdefault:)/g)]
  .filter(match => /s\.(PowerOn|LinkVerified|ArchiveMounted) = true/.test(match[2]))
  .map(match => JSON.parse(match[1].match(/"[^"]+"/)[0]));
if (terminalCommands.length < 3) throw new Error('Could not locate the discoverable Terminal sequence.');
const output = new URL('output/ticket7-private-inputs.json', root);
await mkdir(new URL('output/', root), { recursive: true });
await writeFile(output, JSON.stringify({ phonePassword, terminalCommands }));
console.log(`Private verification inputs prepared at ${fileURLToPath(output)}; values withheld.`);
if (process.argv.includes('--verify-bundle')) {
  const protectedBodies = [phone, terminal].map(source => {
    const literal = source.match(/Bytes:\s*\[\]byte\((`[\s\S]*?`|"(?:\\.|[^"\\])*")\)/)?.[1];
    if (!literal) throw new Error('Protected server definition changed; update disclosure verification.');
    return literal.startsWith('`') ? literal.slice(1, -1) : JSON.parse(literal);
  });
  const withheld = [phonePassword, ...terminalCommands, ...protectedBodies];
  async function inspect(directory) {
    for (const entry of await readdir(directory, { withFileTypes: true })) {
      const target = new URL(entry.name + (entry.isDirectory() ? '/' : ''), directory);
      if (entry.isDirectory()) await inspect(target);
      else {
        const text = await readFile(target, 'utf8');
        if (withheld.some(value => text.includes(value) || text.includes(JSON.stringify(value).slice(1, -1)))) throw new Error('Server-only material appeared in frontend source/build. Values withheld.');
      }
    }
  }
  await inspect(new URL('app/src/', root));
  await inspect(new URL('app/dist/', root));
  console.log('Frontend source and production build withhold actual answer, maintenance commands and protected content.');
}
