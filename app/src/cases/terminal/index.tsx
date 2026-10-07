import { useEffect, useRef, useState, useSyncExternalStore } from 'react';
import { createRoot } from 'react-dom/client';
import type { CaseContext, CaseImplementation } from '../../platform/runtime';
import type { CaseSdk } from '../../platform/sdk';
import './terminal.css';

type TerminalView = {
  cwd: string; entries: { name: string; kind: string }[]; workstation_ready: boolean;
  protected_file?: { asset_id: string; name: string };
};
type CommandOutcome = { lines: string[]; ok: boolean };
type TranscriptEntry = { command: string; cwd: string; lines: string[]; ok: boolean };
type ProtectedFile = { asset_id: string; name: string };
function FileReader({ sdk, file }: { sdk: CaseSdk; file: ProtectedFile }) {
  const [content, setContent] = useState<string | null>(null);
  const [error, setError] = useState(false);
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    const controller = new AbortController(); setContent(null); setError(false);
    sdk.asset(file.asset_id, controller.signal).then(blob => blob.text()).then(text => {
      if (!controller.signal.aborted) setContent(text);
    }).catch(() => { if (!controller.signal.aborted) setError(true); });
    return () => controller.abort();
  }, [sdk, file, attempt]);
  return <section className="terminal-file" aria-label="Protected file"><header><span>FILE VIEW</span><h2>{file.name}</h2></header>
    {error ? <><p role="alert">The file could not be loaded. Check the connection and try again.</p><button onClick={() => setAttempt(value => value + 1)}>Load file again</button></> : content === null ? <p>Loading file…</p> : <pre>{content}</pre>}
  </section>;
}
function Terminal({ sdk, exit }: CaseContext) {
  const snapshot = useSyncExternalStore(sdk.subscribe, sdk.getSnapshot);
  const view = snapshot.view as TerminalView;
  const [command, setCommand] = useState('');
  const [transcript, setTranscript] = useState<TranscriptEntry[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [pending, setPending] = useState<{ command: string; cwd: string } | null>(null);
  const [file, setFile] = useState<ProtectedFile | null>(null);
  const output = useRef<HTMLDivElement>(null);
  const commandInput = useRef<HTMLInputElement>(null);
  const uncertain = sdk.hasUncertainAction();
  const needsRefresh = sdk.needsRefresh();
  const disabled = busy || uncertain || needsRefresh;
  const completed = snapshot.playthrough.completed_at !== null;
  useEffect(() => { if (output.current) output.current.scrollTop = output.current.scrollHeight; }, [transcript]);
  useEffect(() => {
    const element = output.current!;
    const observer = new ResizeObserver(() => { element.scrollTop = element.scrollHeight; });
    observer.observe(element);
    return () => observer.disconnect();
  }, []);
  useEffect(() => { if (!disabled) commandInput.current?.focus({ preventScroll: true }); }, [disabled]);
  async function run(input: string, retry = false) {
    const execution = retry && pending ? pending : { command: input, cwd: view.cwd };
    setBusy(true); setError(''); setNotice(''); setPending(execution);
    try {
      const { result, superseded } = await (retry ? sdk.retry() : sdk.submit('workstation.command', { command: execution.command }));
      setPending(null);
      if (superseded) setNotice('The earlier command was received. Your newer workstation state is still shown.');
      else {
        const outcome = result.outcome as CommandOutcome;
        setTranscript(items => [...items, { ...execution, lines: outcome.lines, ok: outcome.ok }]);
        setCommand('');
        const projected = (result.view as TerminalView).protected_file;
        if (outcome.ok && projected && execution.command.trim() === `cat ${projected.name}`) setFile({ ...projected });
      }
    } catch (cause) {
      if (sdk.hasUncertainAction()) setError('The response was lost. This command may have reached the server. Refresh the view or retry the original command.');
      else { setPending(null); setError(cause instanceof Error ? cause.message : 'Could not send the command. Refresh the view before trying again.'); }
    } finally { setBusy(false); }
  }
  async function refresh() {
    setBusy(true); setError('');
    try { await sdk.refresh(); setNotice('Current workstation state loaded.'); }
    catch { setError('Could not refresh the workstation. Check the connection and try again.'); }
    finally { setBusy(false); }
  }
  // A reader is cosmetic. Only the current server projection can keep its
  // asset available; local transcript/navigation never grants permission.
  const visibleFile = file && view.workstation_ready && view.protected_file?.asset_id === file.asset_id ? file : null;
  return <main className="terminal-case">
    <header className="terminal-heading"><div><p>Terminal Demo / simulated workstation</p><h1>Archive workstation</h1></div><button onClick={exit}>Back to cases</button></header>
    {completed ? <div className="terminal-completion"><h2>Case complete</h2><p>The workstation recorded your file read. You can continue exploring.</p></div> : null}
    <section className="terminal-workbench" aria-label="Archive workstation">
      <aside className="terminal-directory"><p className="terminal-label">Current directory</p><h2>{view.cwd}</h2><ul>{view.entries.map(entry => <li key={entry.name}><span>{entry.name}</span><small>{entry.kind}</small></li>)}</ul>
        <p className="terminal-readiness">{view.workstation_ready ? 'Workstation ready' : 'Workstation requires attention'}</p>
        {view.workstation_ready && view.protected_file ? <button className="terminal-read-file" disabled={disabled} onClick={() => void run(`cat ${view.protected_file!.name}`)}>Read {view.protected_file.name}</button> : null}
      </aside>
      <div className="terminal-console"><div className="terminal-console-title"><span>COMMAND SESSION</span><span>SIMULATED</span></div><div ref={output} className="terminal-output" role="log" aria-label="Command output" aria-live="polite" tabIndex={0}><p className="terminal-welcome">Start with help or read the workstation notes.</p>{transcript.map((entry, index) => <div className="terminal-command-entry" key={index}><p className="terminal-prompt">{entry.cwd} &gt; {entry.command}</p><pre className={entry.ok ? '' : 'terminal-rejected'}>{entry.lines.join('\n')}</pre></div>)}</div>
        <form className="terminal-command-form" onSubmit={event => { event.preventDefault(); void run(command); }}><label htmlFor="workstation-command">Command</label><div><span aria-hidden="true">{view.cwd} &gt;</span><input ref={commandInput} id="workstation-command" value={command} autoComplete="off" spellCheck={false} onChange={event => setCommand(event.target.value)} disabled={disabled} /><button disabled={disabled || !command.trim()}>Run command</button></div></form>
        <div className="terminal-feedback" aria-live="polite">{error ? <p className="terminal-error" role="alert">{error}</p> : null}{notice ? <p>{notice}</p> : null}{busy ? <p>Connecting…</p> : null}
          {uncertain ? <button disabled={busy} onClick={() => void run('', true)}>Retry original command</button> : null}
          <button disabled={busy} onClick={() => void refresh()}>Refresh workstation</button>
        </div>
      </div>
    </section>
    {visibleFile ? <FileReader sdk={sdk} file={visibleFile} /> : null}
  </main>;
}
export const terminal: CaseImplementation = {
  case_id: 'terminal-demo', case_version: 'm0-v1',
  mount(element, context) { const root = createRoot(element); root.render(<Terminal {...context} />); return () => root.unmount(); },
};
