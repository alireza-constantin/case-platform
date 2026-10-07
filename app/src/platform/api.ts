export type CaseMetadata = { case_id: string; case_version: string; title: string };

export type Playthrough = {
  playthrough_id: string; case_id: string; case_version: string; revision: number;
  completed_at: string | null; created_at: string; updated_at: string;
  availability: 'available' | 'unavailable';
};
export type Snapshot = { playthrough: Playthrough; view: unknown };
export type ActionInput = { request_id: string; base_revision: number; action_type: string; payload: unknown };
export type ActionResult = Snapshot & { request_id: string; outcome: unknown };
export class ApiError extends Error {
  constructor(public readonly status: number, public readonly code: string, message: string) { super(message); }
}
async function request<T>(path: string, signal: AbortSignal, body?: unknown): Promise<T> {
  const response = await fetch(path, {
    credentials: 'same-origin', signal, cache: 'no-store',
    ...(body === undefined ? {} : { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) }),
  });
  if (!response.ok) {
    const envelope = await response.json().catch(() => null);
    throw new ApiError(response.status, envelope?.error?.code ?? 'connection_error', envelope?.error?.message ?? 'The request could not be completed.');
  }
  return response.json() as Promise<T>;
}
const runPath = (id: string) => `/api/playthroughs/${encodeURIComponent(id)}`;
export const listPlaythroughs = async (signal: AbortSignal) => (await request<{ playthroughs: Playthrough[] }>('/api/playthroughs', signal)).playthroughs;
export const createPlaythrough = (item: CaseMetadata, signal: AbortSignal) => request<Snapshot>('/api/playthroughs', signal, { case_id: item.case_id, case_version: item.case_version });
export const readPlaythrough = (id: string, signal: AbortSignal) => request<Snapshot>(runPath(id), signal);
export const sendAction = (id: string, input: ActionInput, signal: AbortSignal) => request<ActionResult>(`${runPath(id)}/actions`, signal, input);
export async function readAsset(id: string, assetId: string, signal: AbortSignal): Promise<Blob> {
  const response = await fetch(`${runPath(id)}/assets/${encodeURIComponent(assetId)}`, { credentials: 'same-origin', signal, cache: 'no-store' });
  if (!response.ok) throw new Error('This image could not be loaded. Refresh the view and try again.');
  return response.blob();
}

export async function loadCatalogue(signal: AbortSignal): Promise<CaseMetadata[]> {
  const guest = await fetch('/api/guest', {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: '{}', credentials: 'same-origin', signal,
  });
  if (guest.status !== 204) throw new Error('Guest access could not be established. Check the connection and try again.');
  const response = await fetch('/api/cases', { credentials: 'same-origin', signal });
  if (!response.ok) throw new Error('Cases could not be loaded. Check the connection and try again.');
  const catalogue: { cases: CaseMetadata[] } = await response.json();
  return catalogue.cases;
}
