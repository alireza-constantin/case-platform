import { useEffect, useRef, useState } from 'react';
import { ApiError, createPlaythrough, listPlaythroughs, loadCatalogue, readPlaythrough, type CaseMetadata, type Playthrough, type Snapshot } from './api';
import type { CaseImplementation } from './runtime';
import { createCaseSdk } from './sdk';
import './launcher.css';

type Catalogue = { cases: CaseMetadata[]; runs: Playthrough[] };
function CaseMount({ implementation, snapshot, exit }: { implementation: CaseImplementation; snapshot: Snapshot; exit: () => void }) {
  const root = useRef<HTMLDivElement>(null);
  useEffect(() => {
    // Each mount owns a distinct host. Defer the nested React root's cleanup
    // until the outer root's commit finishes, without reusing the old host.
    const host = document.createElement('div'); root.current!.replaceChildren(host);
    const sdk = createCaseSdk(snapshot);
    const unmount = implementation.mount(host, { initial: snapshot, sdk, exit });
    return () => { sdk.dispose(); host.remove(); queueMicrotask(unmount); };
  }, [implementation, snapshot, exit]);
  return <div ref={root} className="case-root" />;
}

export function Launcher({ implementations }: { implementations: CaseImplementation[] }) {
  const [catalogue, setCatalogue] = useState<Catalogue | null>(null);
  const [snapshot, setSnapshot] = useState<Snapshot | null>(null);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(true);
  const [creationUncertain, setCreationUncertain] = useState(false);
  const [attempt, setAttempt] = useState(0);
  const operation = useRef<AbortController | null>(null);
  useEffect(() => {
    const controller = new AbortController(); operation.current = controller;
    setBusy(true); setError(''); setCatalogue(null);
    (async () => {
      const cases = await loadCatalogue(controller.signal);
      // Assembly agreement uses only opaque metadata, independently of either
      // side's case runtime. Historical unavailable pins are not the catalogue.
      const key = (item: { case_id: string; case_version: string }) => JSON.stringify([item.case_id, item.case_version]);
      const frontend = new Set(implementations.map(key));
      const backend = new Set(cases.map(key));
      if (frontend.size !== implementations.length || backend.size !== cases.length || frontend.size !== backend.size || [...frontend].some(id => !backend.has(id))) {
        throw new Error('Case registration mismatch. Frontend and API case IDs and versions must agree before playing.');
      }
      const runs = await listPlaythroughs(controller.signal);
      if (!controller.signal.aborted) { setCatalogue({ cases, runs }); setCreationUncertain(false); }
      const match = window.location.pathname.match(/^\/play\/([^/]+)/);
      if (match) {
        const current = await readPlaythrough(decodeURIComponent(match[1]), controller.signal);
        if (!controller.signal.aborted) setSnapshot(current);
      }
    })().catch(cause => { if (!controller.signal.aborted) setError(cause instanceof Error ? cause.message : 'Connection unavailable.'); })
      .finally(() => { if (!controller.signal.aborted) setBusy(false); });
    return () => controller.abort();
  }, [attempt, implementations]);
  useEffect(() => {
    const navigate = () => { setSnapshot(null); setAttempt(value => value + 1); };
    window.addEventListener('popstate', navigate);
    return () => window.removeEventListener('popstate', navigate);
  }, []);
  useEffect(() => () => operation.current?.abort(), []);
  async function open(item: CaseMetadata, run?: Playthrough) {
    operation.current?.abort();
    const controller = new AbortController(); operation.current = controller;
    setBusy(true); setError('');
    try {
      const next = run ? await readPlaythrough(run.playthrough_id, controller.signal) : await createPlaythrough(item, controller.signal);
      if (!controller.signal.aborted) {
        window.history.pushState(null, '', `/play/${encodeURIComponent(next.playthrough.playthrough_id)}`);
        setSnapshot(next);
      }
    } catch (cause) {
      if (!controller.signal.aborted) {
        if (!run && (!(cause instanceof ApiError) || cause.status >= 500)) {
          setCreationUncertain(true); setError('The response was lost. A new playthrough may have been created. Reload the case list to check before starting again.');
        } else setError(cause instanceof Error ? cause.message : 'Connection unavailable. Reload the case list to try again.');
      }
    } finally { if (!controller.signal.aborted) setBusy(false); }
  }
  const exit = useRef(() => {
    operation.current?.abort();
    window.history.pushState(null, '', '/'); setSnapshot(null); setAttempt(value => value + 1);
  }).current;
  if (snapshot) {
    const implementation = implementations.find(item => item.case_id === snapshot.playthrough.case_id && item.case_version === snapshot.playthrough.case_version);
    if (implementation && snapshot.playthrough.availability === 'available') return <CaseMount implementation={implementation} snapshot={snapshot} exit={exit} />;
    return <main className="launcher"><h1>This case version is unavailable</h1><p>Your playthrough is retained. Return to choose a fresh run.</p><button onClick={exit}>Back to cases</button></main>;
  }
  return <main className="launcher">
    <header><p className="launcher-label">Case platform</p><h1>Your investigations</h1><p>Choose a case to begin or return to an investigation.</p></header>
    <section aria-label="Case catalogue" aria-live="polite" aria-busy={busy}>
      {busy ? <p>Connecting and loading…</p> : null}
      {error ? <><p role="alert">{error}</p><button disabled={busy} onClick={() => { window.history.replaceState(null, '', '/'); setAttempt(value => value + 1); }}>Reload case list</button></> : null}
      {catalogue ? <>
        <p className="launcher-status">Guest access ready</p>
        {catalogue.cases.length === 0 ? <><h2>No cases available yet</h2><p>Your guest access is ready. Return when an investigation is available.</p></> : <ul>{catalogue.cases.map(item => {
          const run = catalogue.runs.find(candidate => candidate.case_id === item.case_id);
          const available = implementations.some(candidate => candidate.case_id === item.case_id && candidate.case_version === item.case_version);
          return <li key={`${item.case_id}/${item.case_version}`}><h2>{item.title}</h2>
            {!available ? <p>This version is not available in this application.</p> : <div className="launcher-actions">
              {run ? <button disabled={busy || run.availability !== 'available'} onClick={() => void open(item, run)}>Continue {item.title}</button> : null}
              <button disabled={busy || creationUncertain} onClick={() => void open(item)}>{run ? 'Restart' : 'Start'} {item.title}</button>
            </div>}
            {run?.completed_at ? <p>Completed playthrough available</p> : null}
            {run?.availability === 'unavailable' ? <p>The previous version is unavailable. Its progress is retained.</p> : null}
          </li>;
        })}</ul>}
      </> : null}
    </section>
  </main>;
}
