import { ApiError, readAsset, readPlaythrough, sendAction, type ActionInput, type ActionResult, type Snapshot } from './api';

// Transport and lifecycle only. Views, payloads, and outcomes stay opaque here.
export function createCaseSdk(initial: Snapshot) {
  const lifetime = new AbortController();
  const listeners = new Set<() => void>();
  let current = initial;
  let pending: ActionInput | null = null;
  let sending = false;
  let requiresRefresh = false;
  const id = initial.playthrough.playthrough_id;
  function install(next: Snapshot) {
    if (lifetime.signal.aborted) return;
    if (next.playthrough.playthrough_id !== id || next.playthrough.case_id !== initial.playthrough.case_id || next.playthrough.case_version !== initial.playthrough.case_version) throw new Error('The returned playthrough does not match this case.');
    if (next.playthrough.revision >= current.playthrough.revision) {
      current = next; listeners.forEach(listener => listener());
    }
  }
  async function deliver(input: ActionInput): Promise<{ result: ActionResult; superseded: boolean }> {
    if (sending || lifetime.signal.aborted) throw new Error('This request is not available.');
    sending = true;
    try {
      const result = await sendAction(id, input, lifetime.signal);
      if (result.request_id !== input.request_id) throw new Error('The action response could not be matched. Refresh the view before retrying.');
      const superseded = result.playthrough.revision < current.playthrough.revision;
      install(result); pending = null;
      return { result, superseded };
    } catch (cause) {
      // A definite rejection can be reconciled into a new action. A lost/500
      // response may have committed; retain the entire original input for retry.
      if (cause instanceof ApiError && cause.status < 500) {
        pending = null;
        if (cause.code === 'revision_conflict') requiresRefresh = true;
      }
      throw cause;
    } finally { sending = false; }
  }
  return {
    getSnapshot: () => current,
    subscribe(listener: () => void) { listeners.add(listener); return () => { listeners.delete(listener); }; },
    async refresh() { const next = await readPlaythrough(id, lifetime.signal); install(next); requiresRefresh = false; },
    submit(action_type: string, payload: unknown) {
      if (pending) throw new Error('Resolve the previous request before sending another.');
      if (requiresRefresh) throw new Error('Refresh the view before sending another request.');
      pending = { request_id: crypto.randomUUID(), base_revision: current.playthrough.revision, action_type, payload: JSON.parse(JSON.stringify(payload)) };
      return deliver(pending);
    },
    retry() { if (!pending) throw new Error('There is no uncertain request to retry.'); return deliver(pending); },
    hasUncertainAction: () => pending !== null,
    needsRefresh: () => requiresRefresh,
    asset: (assetId: string, signal: AbortSignal) => readAsset(id, assetId, AbortSignal.any([signal, lifetime.signal])),
    dispose() { lifetime.abort(); listeners.clear(); pending = null; },
  };
}
export type CaseSdk = ReturnType<typeof createCaseSdk>;
