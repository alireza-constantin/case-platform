// Explicit synthetic Terminal API boundary. Not application authority or bundled data.
async (page) => {
  const assert = (condition, message) => { if (!condition) throw new Error(message); };
  await page.unroute('**/api/**');
  const metadata = [{ case_id: 'phone-demo', case_version: 'm0-v1', title: 'Phone Demo' }, { case_id: 'terminal-demo', case_version: 'm0-v1', title: 'Terminal Demo' }];
  const summary = { playthrough_id: 'synthetic-terminal', case_id: 'terminal-demo', case_version: 'm0-v1', revision: 0, completed_at: null, created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z', availability: 'available' };
  let snapshot = { playthrough: summary, view: { cwd: '/', entries: [{ name: 'readme.txt', kind: 'file' }], workstation_ready: false } };
  const commands = []; let assetReads = 0;
  await page.route('**/api/**', async route => {
    const req = route.request(); const path = new URL(req.url()).pathname;
    if (path === '/api/guest') return route.fulfill({ status: 204 });
    if (path.endsWith('/assets/synthetic-report')) { assetReads++; return route.fulfill({ contentType: 'text/plain', body: 'Synthetic protected file body for browser verification only.\nSecond synthetic line.' }); }
    let body = snapshot; let status = 200;
    if (path === '/api/cases') body = { cases: metadata };
    if (path === '/api/playthroughs' && req.method() === 'GET') body = { playthroughs: [] };
    if (path === '/api/playthroughs' && req.method() === 'POST') status = 201;
    if (path.endsWith('/actions')) {
      const input = req.postDataJSON(); commands.push(input);
      const command = input.payload.command;
      let outcome = { lines: ['Synthetic command hint.'], ok: true };
      if (command === 'cat report.txt' && !snapshot.view.protected_file) outcome = { lines: ['Synthetic access denied.'], ok: false };
      if (command === 'synthetic-prerequisite-one') { snapshot = { ...snapshot, playthrough: { ...summary, revision: 1 } }; outcome.lines = ['Synthetic first prerequisite recorded.']; }
      if (command === 'synthetic-prerequisite-two') { snapshot = { ...snapshot, playthrough: { ...summary, revision: 2 }, view: { ...snapshot.view, workstation_ready: true, entries: [{ name: 'archive', kind: 'directory' }] } }; outcome.lines = ['Synthetic workstation is ready.']; }
      if (command === 'cd archive') { snapshot = { ...snapshot, playthrough: { ...summary, revision: 3 }, view: { ...snapshot.view, cwd: '/archive', entries: [{ name: 'report.txt', kind: 'file' }], protected_file: { asset_id: 'synthetic-report', name: 'report.txt' } } }; outcome.lines = ['Synthetic archive selected.']; }
      if (command === 'cat report.txt' && snapshot.view.protected_file) { snapshot = { ...snapshot, playthrough: { ...summary, revision: 4, completed_at: '2026-01-01T00:01:00Z' } }; outcome.lines = ['Synthetic file read approved.']; }
      body = { ...snapshot, request_id: input.request_id, outcome };
    }
    return route.fulfill({ status, json: body });
  });
  const send = async (command, returnedLine) => {
    await page.getByLabel('Command', { exact: true }).fill(command);
    await page.getByRole('button', { name: 'Run command', exact: true }).click();
    await page.getByText(returnedLine, { exact: true }).waitFor();
  };
  await page.goto('http://localhost:5185');
  await page.getByRole('button', { name: 'Start Terminal Demo', exact: true }).click({ timeout: 3000 });
  await send('help', 'Synthetic command hint.');
  await send('cat report.txt', 'Synthetic access denied.');
  assert(assetReads === 0 && await page.getByRole('button', { name: 'Read report.txt', exact: true }).count() === 0, 'Denied command/locked projection must not fetch or expose protected file');
  await send('synthetic-prerequisite-one', 'Synthetic first prerequisite recorded.');
  await send('synthetic-prerequisite-two', 'Synthetic workstation is ready.');
  assert(await page.getByText('Workstation ready', { exact: true }).isVisible(), 'Readiness must follow projected view');
  await send('cd archive', 'Synthetic archive selected.');
  assert(assetReads === 0, 'Navigation/file availability alone does not locally perform the protected read');
  await page.getByRole('button', { name: 'Read report.txt', exact: true }).click({ timeout: 3000 });
  await page.getByRole('heading', { name: 'Case complete', exact: true }).waitFor();
  await page.getByText('Synthetic protected file body for browser verification only.', { exact: false }).waitFor();
  assert(assetReads === 1, 'Protected text uses the playthrough-scoped asset transport');
  assert(commands.every(input => input.action_type === 'workstation.command' && typeof input.request_id === 'string') && new Set(commands.map(input => input.request_id)).size === commands.length, 'Commands reuse the neutral envelope with distinct request identities');
  assert(commands.at(-1).base_revision === 3 && commands.at(-1).payload.command === 'cat report.txt', 'Projected-file read submits simulated command with latest base revision');
  const latestLine = await page.getByText('Synthetic file read approved.', { exact: true }).boundingBox();
  const consoleBox = await page.getByRole('log', { name: 'Command output' }).boundingBox();
  assert(latestLine.y >= consoleBox.y && latestLine.y + latestLine.height <= consoleBox.y + consoleBox.height, 'Latest command output must stay visible inside the console viewport');
  console.log('PASS synthetic commands, denied access, projected workstation/file, protected bytes and completion');
}
