// Explicit synthetic API boundary check. Never imported by application code.
async (page) => {
  const assert = (condition, message) => { if (!condition) throw new Error(message); };
  await page.unroute('**/api/**');
  const summary = { playthrough_id: 'synthetic-run', case_id: 'phone-demo', case_version: 'm0-v1', revision: 0, completed_at: null, created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z', availability: 'available' };
  let snapshot = { playthrough: summary, view: { messages: [{ id: 'synthetic-clue', sender: 'Demo contact', text: 'A synthetic public clue for this browser check.' }], gallery_unlocked: false } };
  let assetReads = 0;
  const inputs = [];
  await page.route('**/api/**', async route => {
    const req = route.request(); const path = new URL(req.url()).pathname;
    if (path === '/api/guest') return route.fulfill({ status: 204 });
    if (path.endsWith('/assets/synthetic-image')) {
      assetReads++;
      return route.fulfill({ contentType: 'image/svg+xml', body: '<svg xmlns="http://www.w3.org/2000/svg" width="260" height="200"><rect width="260" height="200" fill="#325b76"/><text x="20" y="90" fill="white">Synthetic image</text></svg>' });
    }
    let body = snapshot; let status = 200;
    if (path === '/api/cases') body = { cases: [{ case_id: 'phone-demo', case_version: 'm0-v1', title: 'Phone Demo' }] };
    if (path === '/api/playthroughs' && req.method() === 'GET') body = { playthroughs: [] };
    if (path === '/api/playthroughs' && req.method() === 'POST') status = 201;
    if (path.endsWith('/actions')) {
      const input = req.postDataJSON(); inputs.push(input);
      const accepted = input.payload.password === 'synthetic-browser-input';
      if (input.action_type === 'gallery.attempt' && accepted) snapshot = { ...snapshot, playthrough: { ...summary, revision: 1 }, view: { ...snapshot.view, gallery_unlocked: true, gallery_assets: [{ asset_id: 'synthetic-image', label: 'Synthetic revealed image' }] } };
      if (input.action_type === 'gallery.open') snapshot = { ...snapshot, playthrough: { ...summary, revision: 2, completed_at: '2026-01-01T00:01:00Z' } };
      body = { ...snapshot, request_id: input.request_id, outcome: input.action_type === 'gallery.open' ? { opened: true } : { accepted } };
    }
    return route.fulfill({ status, json: body });
  });
  await page.goto('http://localhost:5183');
  await page.getByRole('button', { name: 'Start Phone Demo', exact: true }).click({ timeout: 3000 });
  await page.getByRole('button', { name: 'Messages', exact: true }).click();
  assert(await page.getByText('A synthetic public clue for this browser check.').isVisible(), 'Public projected message must be visible');
  await page.getByRole('button', { name: 'Phone home' }).click();
  await page.getByRole('button', { name: 'Gallery', exact: true }).click({ timeout: 3000 });
  assert(assetReads === 0, 'Locked navigation must never fetch protected content');
  await page.getByLabel('Gallery password').fill('synthetic-wrong-input');
  await page.getByRole('button', { name: 'Unlock gallery', exact: true }).click();
  await page.getByText('That password did not unlock the gallery.').waitFor();
  assert(assetReads === 0, 'Rejected attempt must leave protected content unfetched');
  await page.getByLabel('Gallery password').fill('synthetic-browser-input');
  await page.getByRole('button', { name: 'Unlock gallery', exact: true }).click();
  await page.getByRole('button', { name: 'Open gallery', exact: true }).click();
  await page.getByRole('heading', { name: 'Case complete', exact: true }).waitFor();
  await page.getByRole('img', { name: 'Synthetic revealed image' }).waitFor();
  assert(assetReads === 1, 'Gallery content must use the playthrough asset endpoint');
  assert(inputs.length === 3 && inputs.every(input => typeof input.request_id === 'string' && input.request_id.length > 0), 'Actions require distinct request identities');
  assert(new Set(inputs.map(input => input.request_id)).size === 3 && inputs[2].base_revision === 1, 'Later action must use latest revision');
  const imageUrl = await page.getByRole('img', { name: 'Synthetic revealed image' }).getAttribute('src');
  await page.getByRole('button', { name: 'Back to cases', exact: true }).click();
  await page.getByRole('button', { name: 'Start Phone Demo', exact: true }).waitFor();
  assert(await page.getByRole('img', { name: 'Synthetic revealed image' }).count() === 0, 'Unmount removes the protected image from the case root');
  const revoked = await page.evaluate(async url => { try { await fetch(url); return false; } catch { return true; } }, imageUrl);
  assert(revoked, 'Unmount must release the temporary protected asset URL');
  console.log('PASS synthetic message → rejected attempt → unlock → protected gallery → completion');
}
