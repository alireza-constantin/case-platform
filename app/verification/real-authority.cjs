// All ownership/authority checks hit real Go. Local flags never supply approval.
async (page) => {
  const assert = (value, message) => { if (!value) throw new Error(message); };
  const absolute = path => new URL(path, page.url()).href;
  const ids = await page.evaluate(() => JSON.parse(localStorage.getItem('verification-run-ids')));
  const reads = async id => { const response = await page.request.get(absolute(`/api/playthroughs/${id}`)); assert(response.ok(), 'Owned read failed'); return response.json(); };
  const beforePhone = await reads(ids.freshPhoneId); const beforeTerminal = await reads(ids.freshTerminalId);
  for (const [id, asset] of [[ids.freshPhoneId, ids.phoneAsset], [ids.freshTerminalId, ids.terminalAsset]]) {
    const response = await page.request.get(absolute(`/api/playthroughs/${id}/assets/${asset}`));
    assert(response.status() === 404 && (await response.json()).error.code === 'not_found', 'Locked second run cannot inherit protected permission');
  }
  const forge = async (id, action_type, payload) => page.request.post(absolute(`/api/playthroughs/${id}/actions`), { data: { request_id: crypto.randomUUID(), base_revision: 0, action_type, payload } });
  assert((await forge(ids.freshPhoneId, 'gallery.attempt', { password: 'verification-wrong', gallery_unlocked: true })).status() === 422, 'Forged Phone flag cannot assign case state');
  assert((await forge(ids.freshTerminalId, 'workstation.command', { command: 'cat report.txt', workstation_ready: true })).status() === 422, 'Forged Terminal flag cannot assign case state');
  const crossOrigin = await page.request.post(absolute(`/api/playthroughs/${ids.freshPhoneId}/actions`), { headers: { Origin: 'https://unrelated.invalid' }, data: { request_id: crypto.randomUUID(), base_revision: 0, action_type: 'gallery.open', payload: {} } });
  assert(crossOrigin.status() === 400, 'Clearly cross-origin mutation is denied');
  const malformed = await page.request.post(absolute(`/api/playthroughs/${ids.freshPhoneId}/actions`), { headers: { 'Content-Type': 'application/json' }, data: '{' });
  assert(malformed.status() === 400, 'Malformed action cannot mutate progress');
  const oversized = await forge(ids.freshTerminalId, 'workstation.command', { command: 'x'.repeat(200000) });
  assert(oversized.status() === 400, 'Oversized action is bounded');
  const other = await page.context().browser().newContext();
  try {
    assert((await other.request.get(absolute(`/api/playthroughs/${ids.phoneId}`))).status() === 401, 'Missing guest cannot read owned run');
    assert((await other.request.post(absolute('/api/guest'), { data: {} })).status() === 204, 'Another guest boots independently');
    for (const [id, asset] of [[ids.phoneId, ids.phoneAsset], [ids.terminalId, ids.terminalAsset]]) {
      const read = await other.request.get(absolute(`/api/playthroughs/${id}`));
      const mutate = await other.request.post(absolute(`/api/playthroughs/${id}/actions`), { data: { request_id: crypto.randomUUID(), base_revision: 0, action_type: 'verification-forged', payload: {} } });
      const copied = await other.request.get(absolute(`/api/playthroughs/${id}/assets/${asset}`));
      assert([read, mutate, copied].every(response => response.status() === 404), 'Another guest cannot read, mutate or use copied protected URLs');
      const unknown = await other.request.get(absolute('/api/playthroughs/unknown-verification-id'));
      assert(JSON.stringify(await read.json()) === JSON.stringify(await unknown.json()), 'Not-owned and unknown resources share the safe response');
    }
  } finally { await other.close(); }
  await page.evaluate(() => { localStorage.setItem('gallery_unlocked', 'true'); localStorage.setItem('workstation_ready', 'true'); });
  await page.goto(absolute(`/play/${ids.freshPhoneId}/gallery`));
  await page.getByRole('button', { name: 'Gallery', exact: true }).click();
  await page.getByText('The gallery is locked', { exact: true }).waitFor();
  assert(await page.locator('.phone-gallery img').count() === 0, 'Forged local flags and hidden suffix cannot reveal gallery');
  await page.goto(absolute(`/play/${ids.freshTerminalId}/archive`));
  await page.getByText('Workstation requires attention', { exact: true }).waitFor();
  assert(await page.getByRole('region', { name: 'Protected file' }).count() === 0, 'Forged local flags and hidden suffix cannot reveal Terminal file');
  await page.evaluate(() => { localStorage.removeItem('gallery_unlocked'); localStorage.removeItem('workstation_ready'); });
  const afterPhone = await reads(ids.freshPhoneId); const afterTerminal = await reads(ids.freshTerminalId);
  assert(JSON.stringify(beforePhone) === JSON.stringify(afterPhone) && JSON.stringify(beforeTerminal) === JSON.stringify(afterTerminal), 'Rejected/forged requests leave locked projections and metadata unchanged');
  return { realOwnershipAndCopiedUrlsDenied: true, lockedSecondRunsDenied: true, forgedClientStateDenied: true, rejectedRequestsPreserveRevision: true };
}
