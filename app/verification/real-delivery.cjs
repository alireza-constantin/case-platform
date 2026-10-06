// Drop only browser delivery after a REAL Go commit; never invent a response.
async (page) => {
  const assert = (value, message) => { if (!value) throw new Error(message); };
  await page.unrouteAll({ behavior: 'wait' });
  const absolute = path => new URL(path, page.url()).href;
  const ids = await page.evaluate(() => JSON.parse(localStorage.getItem('verification-run-ids')));
  const newRun = await page.request.post(absolute('/api/playthroughs'), { data: { case_id: 'phone-demo', case_version: 'm0-v1' } });
  assert(newRun.status() === 201, 'Delivery scenario requires a fresh owned run');
  ids.freshPhoneId = (await newRun.json()).playthrough.playthrough_id;
  await page.goto(absolute(`/play/${ids.freshPhoneId}`));
  await page.getByRole('button', { name: 'Gallery', exact: true }).click();
  await page.evaluate(() => { const input = document.createElement('input'); input.type = 'file'; input.id = 'verification-private-input'; input.hidden = true; document.body.append(input); });
  await page.locator('#verification-private-input').setInputFiles('output/ticket7-private-inputs.json');
  const privateInputs = JSON.parse(await page.locator('#verification-private-input').evaluate(input => input.files[0].text()));
  await page.locator('#verification-private-input').evaluate(input => input.remove());
  const path = `/api/playthroughs/${ids.freshPhoneId}/actions`;
  let original; let committed; let routeFailure = ''; let dropped = false; let deliveries = 0;
  await page.route(`**${path}`, async route => {
    deliveries++;
    if (!dropped) {
      dropped = true; original = route.request().postDataJSON();
      try {
        const response = await route.fetch(); assert(response.ok(), 'Actual Go action must commit before response loss');
        committed = await response.json();
      } catch (cause) { routeFailure = cause.message; }
      await route.abort('failed');
    } else {
      assert(JSON.stringify(route.request().postDataJSON()) === JSON.stringify(original), 'Browser retry retains exact ID/base/tag/input');
      await route.continue();
    }
  });
  await page.getByLabel('Gallery password').fill('verification-response-loss-wrong');
  await page.getByRole('button', { name: 'Unlock gallery', exact: true }).click();
  await page.getByRole('button', { name: 'Retry original request', exact: true }).waitFor();
  await page.getByRole('alert').waitFor(); // Wait for failed delivery, not merely an in-flight pending button.
  assert(committed?.playthrough?.revision === 1 && committed.outcome.accepted === false && deliveries === 1, `Lost real response check failed (requests=${deliveries}, revision=${committed?.playthrough?.revision}, transport=${routeFailure})`);
  assert(await page.getByText('The gallery is locked', { exact: true }).isVisible(), 'Lost response produces no fake progression');
  const act = async (id, base_revision, action_type, payload, request_id = crypto.randomUUID()) => page.request.post(absolute(`/api/playthroughs/${id}/actions`), { data: { request_id, base_revision, action_type, payload } });
  const unlockResponse = await act(ids.freshPhoneId, 1, 'gallery.attempt', { password: privateInputs.phonePassword }); assert(unlockResponse.ok(), 'A real second writer unlocks at current revision');
  const unlock = await unlockResponse.json();
  const completeResponse = await act(ids.freshPhoneId, unlock.playthrough.revision, 'gallery.open', {}); assert(completeResponse.ok(), 'Real second writer opens gallery');
  const newer = await completeResponse.json();
  await page.getByRole('button', { name: 'Refresh view', exact: true }).click();
  await page.getByRole('heading', { name: 'Case complete', exact: true }).waitFor();
  const replayResponse = page.waitForResponse(response => new URL(response.url()).pathname === path && response.request().method() === 'POST');
  await page.getByRole('button', { name: 'Retry original request', exact: true }).click();
  const replay = await (await replayResponse).json();
  await page.getByText('The earlier request was received. Your newer progress is still shown.', { exact: true }).waitFor();
  assert(JSON.stringify(replay) === JSON.stringify(committed) && deliveries === 2, 'Real receipt replay returns the old recorded response');
  assert(await page.getByRole('heading', { name: 'Case complete', exact: true }).isVisible() && await page.getByText('The gallery is locked', { exact: true }).count() === 0, 'Frontend preserves newer completed/unlocked state over old locked replay');
  await page.unroute(`**${path}`);
  const currentResponse = await page.request.get(absolute(`/api/playthroughs/${ids.freshPhoneId}`));
  const current = await currentResponse.json();
  assert(current.playthrough.revision === newer.playthrough.revision && current.playthrough.completed_at === newer.playthrough.completed_at, 'Replay does not repeat transition/completion');
  const noop = await (await act(ids.freshPhoneId, current.playthrough.revision, 'gallery.open', {})).json();
  assert(noop.playthrough.revision === current.playthrough.revision && noop.playthrough.completed_at === current.playthrough.completed_at, 'Post-completion case action is allowed and preserves first completion');
  const create = async () => {
    const response = await page.request.post(absolute('/api/playthroughs'), { data: { case_id: 'phone-demo', case_version: 'm0-v1' } });
    assert(response.status() === 201, 'Owned race run creation failed'); return (await response.json()).playthrough.playthrough_id;
  };
  const distinctId = await create();
  const distinct = await Promise.all(Array.from({ length: 4 }, () => act(distinctId, 0, 'gallery.attempt', { password: 'verification-race-wrong' })));
  assert(distinct.filter(response => response.status() === 200).length === 1 && distinct.filter(response => response.status() === 409).length === 3, 'Only one distinct write commits from a shared base revision');
  const duplicateId = await create(); const requestId = crypto.randomUUID();
  const duplicates = await Promise.all(Array.from({ length: 4 }, () => act(duplicateId, 0, 'gallery.attempt', { password: 'verification-duplicate-wrong' }, requestId)));
  const recorded = await duplicates[0].json();
  assert(duplicates.every(response => response.ok()) && recorded.playthrough.revision === 1, 'Concurrent identical delivery commits only one transition');
  for (const response of duplicates.slice(1)) assert(JSON.stringify(await response.json()) === JSON.stringify(recorded), 'Concurrent duplicate response matches committed receipt');
  const conflict = await act(duplicateId, 0, 'gallery.attempt', { password: 'verification-changed-input' }, requestId);
  assert(conflict.status() === 409 && (await conflict.json()).error.code === 'request_id_conflict', 'Changed input under committed ID is rejected');
  return { lostRealResponseReplayed: true, newerFrontendRevisionPreserved: current.playthrough.revision, concurrentDistinct: { committed: 1, stale: 3 }, concurrentIdentical: { responses: 4, revision: 1 }, changedInputConflict: true };
}
