// Two phases: loaded Terminal with owned Go stopped, then explicit retry after Go restarts.
async (page) => {
  const assert = (value, message) => { if (!value) throw new Error(message); };
  const absolute = path => new URL(path, page.url()).href;
  const phase = await page.evaluate(() => localStorage.getItem('verification-connection-phase'));
  const id = new URL(page.url()).pathname.split('/')[2];
  const path = `/api/playthroughs/${id}/actions`;
  if (phase !== 'recovery') {
    let input; let requests = 0;
    const observed = request => { if (new URL(request.url()).pathname === path) { requests++; input = request.postDataJSON(); } };
    page.on('request', observed);
    await page.getByLabel('Command', { exact: true }).fill('cat report.txt');
    await page.getByRole('button', { name: 'Run command', exact: true }).click();
    await page.getByRole('alert').waitFor();
    await page.getByRole('button', { name: 'Retry original command', exact: true }).waitFor();
    await page.waitForTimeout(700);
    assert(requests === 1 && await page.getByText('Workstation requires attention', { exact: true }).isVisible() && await page.getByRole('heading', { name: 'Case complete', exact: true }).count() === 0, 'Stopped API gives uncertainty without fake success or an offline action queue');
    page.off('request', observed);
    await page.evaluate(input => { localStorage.setItem('verification-connection-input', JSON.stringify(input)); localStorage.setItem('verification-connection-phase', 'recovery'); }, input);
    return { stoppedApiActionFailedCleanly: true, automaticRetries: 0 };
  }
  const expected = await page.evaluate(() => JSON.parse(localStorage.getItem('verification-connection-input')));
  await page.route(`**${path}`, async route => { assert(JSON.stringify(route.request().postDataJSON()) === JSON.stringify(expected), 'Connection recovery retry keeps exact original envelope'); await route.continue(); });
  const responsePromise = page.waitForResponse(response => new URL(response.url()).pathname === path && response.request().method() === 'POST');
  await page.getByRole('button', { name: 'Retry original command', exact: true }).click();
  const response = await responsePromise; const result = await response.json();
  assert(response.ok() && result.outcome.ok === false && result.playthrough.revision === 0 && result.playthrough.completed_at === null, 'Explicit retry uses real unchanged locked state after connection returns');
  await page.unroute(`**${path}`);
  await page.evaluate(() => { localStorage.removeItem('verification-connection-input'); localStorage.removeItem('verification-connection-phase'); });
  return { explicitOriginalRetryAfterApiRestart: true, realLockedRevision: result.playthrough.revision };
}
