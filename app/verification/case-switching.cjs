// Explicit synthetic Phone → Terminal → Phone lifecycle and projected resume.
async (page) => {
  const assert = (condition, message) => { if (!condition) throw new Error(message); };
  await page.unroute('**/api/**');
  const runtimeErrors = []; const capture = message => { if (message.type() === 'error' && message.text().includes('React')) runtimeErrors.push(message.text()); };
  page.on('console', capture);
  const metadata = [{ case_id: 'phone-demo', case_version: 'm0-v1', title: 'Phone Demo' }, { case_id: 'terminal-demo', case_version: 'm0-v1', title: 'Terminal Demo' }];
  const summary = { case_version: 'm0-v1', revision: 2, completed_at: '2026-01-01T00:01:00Z', created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z', availability: 'available' };
  const phone = { playthrough: { ...summary, case_id: 'phone-demo', playthrough_id: 'synthetic-phone' }, view: { messages: [{ id: 'clue', sender: 'Demo contact', text: 'Phone committed projected clue.' }], gallery_unlocked: true, gallery_assets: [] } };
  let terminal = { playthrough: { ...summary, case_id: 'terminal-demo', playthrough_id: 'synthetic-terminal', completed_at: null }, view: { cwd: '/archive', entries: [{ name: 'report.txt', kind: 'file' }], workstation_ready: true, protected_file: { name: 'report.txt', asset_id: 'synthetic-terminal-report' } } };
  let releaseAsset; const heldAsset = new Promise(resolve => { releaseAsset = resolve; });
  let assetFailed; const canceledAsset = new Promise(resolve => { assetFailed = resolve; });
  const failure = request => { if (request.url().endsWith('/assets/synthetic-terminal-report')) assetFailed(); };
  page.on('requestfailed', failure);
  await page.route('**/api/**', async route => {
    const req = route.request(); const path = new URL(req.url()).pathname;
    if (path === '/api/guest') return route.fulfill({ status: 204 });
    if (path.endsWith('/assets/synthetic-terminal-report')) {
      await heldAsset; return route.fulfill({ contentType: 'text/plain', body: 'Synthetic delayed file body.' }).catch(() => {});
    }
    let body = path.includes('synthetic-phone') ? phone : terminal;
    if (path === '/api/cases') body = { cases: metadata };
    if (path === '/api/playthroughs') body = { playthroughs: [phone.playthrough, terminal.playthrough] };
    if (path.endsWith('/actions')) {
      const input = req.postDataJSON();
      if (input.payload.command === 'cat report.txt') terminal = { ...terminal, playthrough: { ...terminal.playthrough, revision: 3, completed_at: summary.completed_at } };
      body = { ...terminal, request_id: input.request_id, outcome: { ok: true, lines: [input.payload.command === 'cat report.txt' ? 'Synthetic read accepted.' : 'Synthetic transient terminal output.'] } };
    }
    return route.fulfill({ json: body });
  });
  await page.goto('http://localhost:5185');
  await page.getByRole('button', { name: 'Continue Phone Demo', exact: true }).click();
  await page.getByRole('heading', { name: 'Case complete', exact: true }).waitFor();
  await page.getByRole('button', { name: 'Messages', exact: true }).click();
  await page.getByText('Phone committed projected clue.', { exact: true }).waitFor();
  await page.getByRole('button', { name: 'Back to cases', exact: true }).click();
  await page.getByRole('button', { name: 'Continue Terminal Demo', exact: true }).click();
  await page.getByLabel('Command', { exact: true }).fill('synthetic-temporary-command');
  await page.getByRole('button', { name: 'Run command', exact: true }).click();
  await page.getByText('Synthetic transient terminal output.', { exact: true }).waitFor();
  const assetRequested = page.waitForRequest('**/assets/synthetic-terminal-report');
  await page.getByRole('button', { name: 'Read report.txt', exact: true }).click();
  await assetRequested;
  await page.getByRole('button', { name: 'Back to cases', exact: true }).click();
  await page.getByRole('button', { name: 'Continue Phone Demo', exact: true }).click();
  await Promise.race([canceledAsset, new Promise((_, reject) => setTimeout(() => reject(new Error('Leaving Terminal must cancel its pending protected fetch')), 3000))]);
  releaseAsset();
  await page.getByRole('button', { name: 'Messages', exact: true }).waitFor();
  assert(await page.getByRole('button', { name: 'Messages', exact: true }).isVisible(), 'Resumed Phone starts in its own clean local navigation');
  await page.getByRole('heading', { name: 'Case complete', exact: true }).waitFor();
  await page.getByRole('button', { name: 'Messages', exact: true }).click();
  assert(await page.getByText('Phone committed projected clue.', { exact: true }).isVisible(), 'Committed Phone projection survives Terminal switching');
  assert(await page.getByRole('log', { name: 'Command output' }).count() === 0 && await page.getByText('Synthetic delayed file body.', { exact: true }).count() === 0, 'No Terminal reader/output leaks into Phone');
  await page.getByRole('button', { name: 'Back to cases', exact: true }).click();
  await page.getByRole('button', { name: 'Continue Terminal Demo', exact: true }).click();
  await page.getByRole('heading', { name: 'Case complete', exact: true }).waitFor();
  assert(await page.getByRole('heading', { name: '/archive', exact: true }).isVisible() && await page.getByText('Synthetic transient terminal output.', { exact: true }).count() === 0, 'Terminal resumes committed server state with a clean transient transcript');
  await page.reload();
  await page.getByRole('heading', { name: 'Case complete', exact: true }).waitFor();
  assert(await page.getByRole('heading', { name: '/archive', exact: true }).isVisible(), 'Reload reads Terminal projected resume state');
  assert(runtimeErrors.length === 0, 'No React lifecycle errors while switching: ' + runtimeErrors.join('; '));
  page.off('console', capture); page.off('requestfailed', failure);
  console.log('PASS synthetic Phone/Terminal switching, pending asset cancellation and projected resumes');
}
