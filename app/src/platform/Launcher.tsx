import { useEffect, useState } from 'react';
import { loadCatalogue, type CaseMetadata } from './api';
import './launcher.css';

type Status = { kind: 'loading' } | { kind: 'ready'; cases: CaseMetadata[] } | { kind: 'error'; message: string };

export function Launcher() {
  const [status, setStatus] = useState<Status>({ kind: 'loading' });
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    const controller = new AbortController();
    setStatus({ kind: 'loading' });
    loadCatalogue(controller.signal).then(
      cases => { if (!controller.signal.aborted) setStatus({ kind: 'ready', cases }); },
      error => { if (!controller.signal.aborted) setStatus({ kind: 'error', message: error instanceof Error ? error.message : 'Connection unavailable. Try again.' }); },
    );
    return () => controller.abort();
  }, [attempt]);

  return <main className="launcher">
    <header><p className="launcher-label">Case platform</p><h1>Your investigations</h1><p>Choose a case to begin or return to an investigation.</p></header>
    <section aria-label="Case catalogue" aria-live="polite" aria-busy={status.kind === 'loading'}>
      {status.kind === 'loading' ? <p>Connecting and loading cases…</p> : null}
      {status.kind === 'error' ? <><p role="alert">{status.message}</p><button onClick={() => setAttempt(value => value + 1)}>Try again</button></> : null}
      {status.kind === 'ready' ? <>
        <p className="launcher-status">Guest access ready</p>
        {status.cases.length === 0 ? <><h2>No cases available yet</h2><p>Your guest access is ready. Return when an investigation is available.</p></> : <ul>{status.cases.map(item => <li key={`${item.case_id}/${item.case_version}`}>{item.title}</li>)}</ul>}
      </> : null}
    </section>
  </main>;
}
