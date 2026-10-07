// Real catalogue comparison; change only frontend assembly for the mismatch probe.
async (page) => {
  const assert = (value, message) => { if (!value) throw new Error(message); };
  const mismatchExpected = new URL(page.url()).searchParams.has('expectMismatch');
  await page.reload();
  if (mismatchExpected) {
    await page.getByRole('alert').waitFor({ timeout: 3000 });
    assert((await page.getByRole('alert').textContent()).includes('Case registration mismatch'), 'Mismatched assembled metadata must fail clearly');
    assert(await page.getByRole('button', { name: /^(Start|Continue|Restart) / }).count() === 0, 'Mismatch blocks case execution');
    return { mismatchedAssemblyBlocked: true };
  }
  await page.getByText('Guest access ready', { exact: true }).waitFor();
  const response = await page.request.get(new URL('/api/cases', page.url()).href);
  const catalogue = await response.json();
  assert(catalogue.cases.length === 2 && await page.getByRole('alert').count() === 0, 'Matching real assembly starts normally');
  return { realMatchingRegistrations: catalogue.cases.map(item => ({ case_id: item.case_id, case_version: item.case_version })) };
}
