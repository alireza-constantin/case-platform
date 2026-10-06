export type CaseMetadata = { case_id: string; case_version: string; title: string };

// Only the two bootstrap operations are implemented in this ticket.
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
