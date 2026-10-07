// Explicit synthetic transport: browser recovery/lifecycle behavior, not real authority.
async (page) => {
  const assert = (condition, message) => { if (!condition) throw new Error(message); };
  await page.unroute('**/api/**');
  const errors = []; const capture = msg => { if (msg.type() === 'error' && msg.text().includes('React')) errors.push(msg.text()); };
  page.on('console', capture);
  const summary = { playthrough_id: 'synthetic-original', case_id: 'phone-demo', case_version: 'm0-v1', revision: 0, completed_at: null, created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z', availability: 'available' };
  const locked = { playthrough: summary, view: { messages: [{ id: 'clue', sender: 'Demo', text: 'Original projected message' }], gallery_unlocked: false } };
  let current = locked; let runs = [summary]; let originalInput; let actions = 0; let creations = 0;
  await page.route('**/api/**', async route => {
    const req = route.request(); const path = new URL(req.url()).pathname;
    if (path === '/api/guest') return route.fulfill({ status: 204 });
    let body = current; let status = 200;
    if (path === '/api/cases') body = { cases: [{ case_id: 'phone-demo', case_version: 'm0-v1', title: 'Phone Demo' }, { case_id: 'terminal-demo', case_version: 'm0-v1', title: 'Terminal Demo' }] };
    if (path === '/api/playthroughs' && req.method() === 'GET') body = { playthroughs: runs };
    if (path === '/api/playthroughs' && req.method() === 'POST') {
      creations++; current = { ...locked, playthrough: { ...summary, playthrough_id: 'synthetic-fresh' } }; runs = [current.playthrough, ...runs]; body = current; status = 201;
    }
    if (path.endsWith('/actions')) {
      actions++; const input = req.postDataJSON();
      if (actions === 1) {
        originalInput = input;
        // Model an uncertain committed response, then another writer's newer view.
        current = { playthrough: { ...summary, revision: 2, completed_at: '2026-01-01T00:01:00Z' }, view: { messages: [{ id: 'new', sender: 'Demo', text: 'Newer projected message' }], gallery_unlocked: true, gallery_assets: [] } };
        runs = [current.playthrough]; return route.abort('failed');
      }
      assert(JSON.stringify(input) === JSON.stringify(originalInput), 'Uncertain retry must preserve request ID, original input and base revision');
      body = { playthrough: { ...summary, revision: 1 }, view: { ...locked.view, gallery_unlocked: true }, request_id: input.request_id, outcome: { accepted: true } };
    }
    return route.fulfill({ status, json: body });
  });
  await page.goto(new URL(page.url()).origin);
  await page.getByRole('button', { name: 'Continue Phone Demo', exact: true }).click();
  await page.getByRole('button', { name: 'Gallery', exact: true }).click();
  await page.getByLabel('Gallery password').fill('synthetic-original-input');
  await page.getByRole('button', { name: 'Unlock gallery', exact: true }).click();
  await page.getByRole('button', { name: 'Retry original request', exact: true }).waitFor();
  assert(await page.getByText('The gallery is locked', { exact: true }).isVisible(), 'Lost response cannot create local success');
  assert(actions === 1, 'No automatic action retry or offline queue');
  await page.getByRole('button', { name: 'Refresh view', exact: true }).click();
  await page.getByRole('heading', { name: 'Case complete', exact: true }).waitFor();
  await page.getByRole('button', { name: 'Retry original request', exact: true }).click();
  await page.getByText('The earlier request was received. Your newer progress is still shown.').waitFor();
  assert(await page.getByRole('heading', { name: 'Case complete', exact: true }).isVisible(), 'Older replay must preserve completion from newer revision');
  await page.getByRole('button', { name: 'Phone home' }).click();
  await page.getByRole('button', { name: 'Messages', exact: true }).click();
  assert(await page.getByText('Newer projected message', { exact: true }).isVisible(), 'Older replay must not overwrite projected view');
  await page.reload();
  await page.getByRole('heading', { name: 'Case complete', exact: true }).waitFor();
  await page.getByRole('button', { name: 'Messages', exact: true }).click();
  assert(await page.getByText('Newer projected message', { exact: true }).isVisible(), 'Resume must read the API projection');
  await page.getByRole('button', { name: 'Back to cases', exact: true }).click();
  await page.getByRole('button', { name: 'Restart Phone Demo', exact: true }).click();
  await page.getByRole('button', { name: 'Gallery', exact: true }).click();
  assert(page.url().endsWith('/play/synthetic-fresh') && creations === 1 && runs.length === 2, 'Restart must request a new run identity');
  assert(await page.getByText('The gallery is locked', { exact: true }).isVisible(), 'Restart must not inherit local unlock or navigation');
  await page.getByRole('button', { name: 'Back to cases', exact: true }).click();
  await page.getByRole('button', { name: 'Continue Phone Demo', exact: true }).waitFor();
  assert(errors.length === 0, 'Case switching must unmount roots without React lifecycle errors: ' + errors.join('; '));
  page.off('console', capture);
  console.log('PASS uncertain retry, newer revision preservation, projected reload, fresh identity and unmount');
}
