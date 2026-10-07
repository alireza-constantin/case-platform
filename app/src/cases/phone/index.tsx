import { useEffect, useState, useSyncExternalStore } from 'react';
import { createRoot } from 'react-dom/client';
import type { CaseContext, CaseImplementation } from '../../platform/runtime';
import type { CaseSdk } from '../../platform/sdk';
import './phone.css';

type PhoneView = { messages: { id: string; sender: string; text: string }[]; gallery_unlocked: boolean; gallery_assets?: { asset_id: string; label: string }[] };
function GalleryImage({ sdk, asset }: { sdk: CaseSdk; asset: { asset_id: string; label: string } }) {
  const [url, setUrl] = useState('');
  const [error, setError] = useState('');
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    const controller = new AbortController();
    let objectUrl = '';
    setUrl(''); setError('');
    sdk.asset(asset.asset_id, controller.signal).then(blob => {
      if (!controller.signal.aborted) { objectUrl = URL.createObjectURL(blob); setUrl(objectUrl); }
    }).catch(cause => { if (!controller.signal.aborted) setError(cause instanceof Error ? cause.message : 'Image unavailable.'); });
    return () => { controller.abort(); if (objectUrl) URL.revokeObjectURL(objectUrl); };
  }, [sdk, asset.asset_id, attempt]);
  return <figure>{url ? <img src={url} alt={asset.label} /> : error ? <><p role="alert">{error}</p><button onClick={() => setAttempt(value => value + 1)}>Load image again</button></> : <p>Loading image…</p>}<figcaption>{asset.label}</figcaption></figure>;
}
function Phone({ sdk, exit }: CaseContext) {
  const snapshot = useSyncExternalStore(sdk.subscribe, sdk.getSnapshot);
  const view = snapshot.view as PhoneView;
  const [screen, setScreen] = useState<'home' | 'messages' | 'gallery'>('home');
  const [password, setPassword] = useState('');
  const [notice, setNotice] = useState('');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [opened, setOpened] = useState(false);
  const uncertain = sdk.hasUncertainAction();
  const needsRefresh = sdk.needsRefresh();
  const completed = snapshot.playthrough.completed_at !== null;
  async function act(kind: 'attempt' | 'open' | 'retry') {
    setBusy(true); setError(''); setNotice('');
    try {
      const { result, superseded } = await (kind === 'retry' ? sdk.retry() : kind === 'attempt' ? sdk.submit('gallery.attempt', { password }) : sdk.submit('gallery.open', {}));
      if (superseded) { setNotice('The earlier request was received. Your newer progress is still shown.'); }
      else {
        const outcome = result.outcome as { accepted?: boolean; opened?: boolean };
        if (outcome.accepted === false) setNotice('That password did not unlock the gallery.');
        if (outcome.accepted === true) { setPassword(''); setNotice('Gallery unlocked.'); }
        if (outcome.opened === true) setOpened(true);
      }
    } catch (cause) {
      setError(sdk.hasUncertainAction() ? 'The response was lost. This request may have reached the server. Refresh the view or retry the original request.' : cause instanceof Error ? cause.message : 'The request failed. Refresh the view before trying again.');
    } finally { setBusy(false); }
  }
  async function refresh() {
    setBusy(true); setError('');
    try { await sdk.refresh(); setNotice('Current progress loaded.'); }
    catch { setError('Could not refresh progress. Check the connection and try again.'); }
    finally { setBusy(false); }
  }
  const showAssets = screen === 'gallery' && view.gallery_unlocked && (opened || completed);
  return <main className="phone-case">
    <aside className="phone-intro"><p className="phone-eyebrow">Phone Demo</p><h1>A phone.<br /> A missing piece.</h1><p>Read the messages. Find the connection. See what the gallery holds.</p>
      {completed ? <div className="phone-completion"><h2>Case complete</h2><p>You opened the gallery. You can keep exploring this phone.</p></div> : null}
      <button onClick={exit}>Back to cases</button>
    </aside>
    <section className="phone-device" aria-label="Recovered phone">
      <div className="phone-status"><span>09:41</span><span>● ● ● ▰</span></div>
      <header><button onClick={() => setScreen('home')} aria-label="Phone home">‹</button><h2>{screen === 'messages' ? 'Messages' : screen === 'gallery' ? 'Gallery' : 'Recovered phone'}</h2></header>
      <div className="phone-content">{screen === 'home' ? <><p className="phone-owner">Look a little closer.</p><nav>
        <button onClick={() => setScreen('messages')}><span aria-hidden="true">▤</span>Messages</button>
        <button onClick={() => setScreen('gallery')}><span aria-hidden="true">▧</span>Gallery</button>
      </nav><p className="phone-hint">The messages might hold the key.</p></> : screen === 'messages' ? <ul className="phone-messages">{view.messages.map(message => <li key={message.id}><h3>{message.sender}</h3><p>{message.text}</p></li>)}</ul> : <div className="phone-gallery">
        {!view.gallery_unlocked ? <><div className="phone-lock" aria-hidden="true">◇</div><h3>The gallery is locked</h3><p>Read the messages for a clue to its password.</p><form onSubmit={event => { event.preventDefault(); void act('attempt'); }}>
          <label htmlFor="gallery-password">Gallery password</label><input id="gallery-password" type="password" autoComplete="off" value={password} onChange={event => setPassword(event.target.value)} disabled={busy || uncertain} />
          <button disabled={busy || uncertain || needsRefresh || password.length === 0}>Unlock gallery</button>
        </form></> : showAssets ? <>{view.gallery_assets?.map(asset => <GalleryImage key={asset.asset_id} sdk={sdk} asset={asset} />)}{!view.gallery_assets?.length ? <p>No gallery images in the current view.</p> : null}</> : <><h3>Gallery unlocked</h3><p>The phone is ready to show its gallery.</p><button disabled={busy || uncertain || needsRefresh} onClick={() => void act('open')}>Open gallery</button></>}
      </div>}
        <div className="phone-feedback" aria-live="polite">{notice ? <p>{notice}</p> : null}{error ? <p role="alert">{error}</p> : null}{busy ? <p>Connecting…</p> : null}
          {uncertain ? <button disabled={busy} onClick={() => void act('retry')}>Retry original request</button> : null}
          <button className="phone-refresh" disabled={busy} onClick={() => void refresh()}>Refresh view</button>
        </div>
      </div><div className="phone-home-indicator" />
    </section>
  </main>;
}
export const phone: CaseImplementation = {
  case_id: 'phone-demo', case_version: 'm0-v1',
  mount(element, context) {
    const root = createRoot(element); root.render(<Phone {...context} />);
    return () => root.unmount();
  },
};
