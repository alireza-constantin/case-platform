// Explicit synthetic Terminal recovery at the API/browser boundary.
async (page) => {
  const assert = (condition, message) => { if (!condition) throw new Error(message); };
  await page.unroute('**/api/**');
  const summary = { playthrough_id: 'synthetic-recovery', case_id: 'terminal-demo', case_version: 'm0-v1', revision: 0, completed_at: null, created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z', availability: 'available' };
  const locked = { playthrough: summary, view: { cwd: '/', entries: [{ name: 'readme.txt', kind: 'file' }], workstation_ready: false } };
  let current = locked; let original; const inputs = []; let assetReads = 0;
  await page.route('**/api/**', async route => {
    const req = route.request(); const path = new URL(req.url()).pathname;
    if (path === '/api/guest') return route.fulfill({ status: 204 });
    if (path.includes('/assets/')) { assetReads++; return route.fulfill({ body: 'Synthetic file' }); }
    let body = current;
    if (path === '/api/cases') body = { cases: [{ case_id: 'terminal-demo', case_version: 'm0-v1', title: 'Terminal Demo' }, { case_id: 'phone-demo', case_version: 'm0-v1', title: 'Phone Demo' }] };
    if (path === '/api/playthroughs') body = { playthroughs: [current.playthrough] };
    if (path.endsWith('/actions')) {
      const input = req.postDataJSON(); inputs.push(input);
      if (inputs.length === 1) {
        original = input; current = { playthrough: { ...summary, revision: 3, completed_at: '2026-01-01T00:01:00Z' }, view: { cwd: '/synthetic-newer', entries: [{ name: 'newer-note.txt', kind: 'file' }], workstation_ready: true, protected_file: { name: 'report.txt', asset_id: 'synthetic-report' } } };
        return route.abort('failed');
      }
      if (inputs.length === 2) {
        assert(JSON.stringify(input) === JSON.stringify(original), 'Retry retains exact original request identity and command');
        body = { ...locked, playthrough: { ...summary, revision: 1 }, request_id: input.request_id, outcome: { ok: true, lines: ['Older synthetic command outcome'] } };
      } else if (inputs.length === 3) {
        current = { ...current, playthrough: { ...current.playthrough, revision: 4 } };
        return route.fulfill({ status: 409, json: { error: { code: 'revision_conflict', message: 'Synthetic view changed. Refresh before another command.' } } });
      } else body = { ...current, request_id: input.request_id, outcome: { ok: false, lines: ['Synthetic command rejected.'] } };
    }
    return route.fulfill({ json: body });
  });
  await page.goto('http://localhost:5185');
  await page.getByRole('button', { name: 'Continue Terminal Demo', exact: true }).click();
  await page.getByLabel('Command', { exact: true }).fill('synthetic-original-command');
  await page.getByRole('button', { name: 'Run command', exact: true }).click();
  await page.getByRole('button', { name: 'Retry original command', exact: true }).waitFor();
  assert(await page.getByText('Workstation requires attention', { exact: true }).isVisible() && inputs.length === 1, 'Lost response leaves existing view and never auto-retries');
  await page.getByRole('button', { name: 'Refresh workstation', exact: true }).click();
  await page.getByRole('heading', { name: 'Case complete', exact: true }).waitFor();
  await page.getByRole('button', { name: 'Retry original command', exact: true }).click();
  await page.getByText('The earlier command was received. Your newer workstation state is still shown.', { exact: true }).waitFor();
  assert(await page.getByRole('heading', { name: '/synthetic-newer', exact: true }).isVisible() && await page.getByRole('heading', { name: 'Case complete', exact: true }).isVisible(), 'Older replay cannot roll newer view or completion backward');
  assert(await page.getByText('Older synthetic command outcome', { exact: true }).count() === 0 && assetReads === 0, 'Superseded outcome cannot create old output or a protected read');
  await page.getByLabel('Command', { exact: true }).fill('synthetic-conflict-command');
  await page.getByRole('button', { name: 'Run command', exact: true }).click();
  await page.getByText('Synthetic view changed. Refresh before another command.', { exact: true }).waitFor();
  assert(await page.getByRole('button', { name: 'Retry original command', exact: true }).count() === 0 && await page.getByRole('button', { name: 'Run command', exact: true }).isDisabled(), 'Definite conflict requires refresh instead of uncertain retry');
  await page.getByRole('button', { name: 'Refresh workstation', exact: true }).click();
  await page.getByText('Current workstation state loaded.', { exact: true }).waitFor();
  await page.getByRole('button', { name: 'Run command', exact: true }).click();
  await page.getByText('Synthetic command rejected.', { exact: true }).waitFor();
  assert(inputs[3].base_revision === 4 && inputs[3].request_id !== inputs[2].request_id, 'Reconciled command uses a new ID and current revision');
  console.log('PASS synthetic uncertain original command, newer projection preservation and definite conflict');
}
