// Explicit synthetic API check for uncertain creation and definite action conflicts.
async (page) => {
  const assert = (condition, message) => { if (!condition) throw new Error(message); };
  await page.unroute('**/api/**');
  const summary = { playthrough_id: 'synthetic-created', case_id: 'phone-demo', case_version: 'm0-v1', revision: 0, completed_at: null, created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z', availability: 'available' };
  let snapshot = { playthrough: summary, view: { messages: [], gallery_unlocked: false } };
  let runs = []; let creations = 0; const inputs = [];
  await page.route('**/api/**', async route => {
    const req = route.request(); const path = new URL(req.url()).pathname;
    if (path === '/api/guest') return route.fulfill({ status: 204 });
    let body = snapshot;
    if (path === '/api/cases') body = { cases: [{ case_id: 'phone-demo', case_version: 'm0-v1', title: 'Phone Demo' }, { case_id: 'terminal-demo', case_version: 'm0-v1', title: 'Terminal Demo' }] };
    if (path === '/api/playthroughs' && req.method() === 'GET') body = { playthroughs: runs };
    if (path === '/api/playthroughs' && req.method() === 'POST') { creations++; runs = [summary]; return route.abort('failed'); }
    if (path.endsWith('/actions')) {
      const input = req.postDataJSON(); inputs.push(input);
      if (inputs.length === 1) {
        snapshot = { ...snapshot, playthrough: { ...summary, revision: 1 } };
        return route.fulfill({ status: 409, json: { error: { code: 'revision_conflict', message: 'The view changed. Refresh it before another attempt.' } } });
      }
      body = { ...snapshot, request_id: input.request_id, outcome: { accepted: false } };
    }
    return route.fulfill({ json: body });
  });
  await page.goto(new URL(page.url()).origin);
  await page.getByRole('button', { name: 'Start Phone Demo', exact: true }).click();
  await page.getByRole('alert').waitFor();
  assert(await page.getByRole('button', { name: 'Start Phone Demo', exact: true }).isDisabled(), 'Uncertain creation must require own-run reconciliation before another creation');
  assert(creations === 1, 'No blind creation retry');
  await page.getByRole('button', { name: 'Reload case list', exact: true }).click();
  await page.getByRole('button', { name: 'Continue Phone Demo', exact: true }).click();
  await page.getByRole('button', { name: 'Gallery', exact: true }).click();
  await page.getByLabel('Gallery password').fill('synthetic-conflicted-input');
  await page.getByRole('button', { name: 'Unlock gallery', exact: true }).click();
  await page.getByText('The view changed. Refresh it before another attempt.').waitFor();
  assert(await page.getByRole('button', { name: 'Retry original request', exact: true }).count() === 0 && inputs.length === 1, 'Definite conflict must not become an uncertain or automatic retry');
  await page.getByRole('button', { name: 'Refresh view', exact: true }).click();
  await page.getByText('Current progress loaded.').waitFor();
  await page.getByRole('button', { name: 'Unlock gallery', exact: true }).click();
  await page.getByText('That password did not unlock the gallery.').waitFor();
  assert(inputs[1].base_revision === 1 && inputs[1].request_id !== inputs[0].request_id, 'Reconciled action needs new ID and current revision');
  console.log('PASS uncertain creation reconciliation and definite revision conflict');
}
