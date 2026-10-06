import { readdir, readFile } from 'node:fs/promises';
import { dirname, join, relative, resolve, sep } from 'node:path';
import { fileURLToPath } from 'node:url';
import ts from 'typescript';

const root = fileURLToPath(new URL('../src/', import.meta.url));
const violations = [];
async function inspect(directory) {
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const file = join(directory, entry.name);
    if (entry.isDirectory()) { await inspect(file); continue; }
    if (!/\.tsx?$/.test(file)) continue;
    const owner = relative(root, file).split(sep);
    const source = ts.createSourceFile(file, await readFile(file, 'utf8'), ts.ScriptTarget.Latest, true);
    function check(specifier) {
      if (!ts.isStringLiteralLike(specifier)) {
        violations.push(`${relative(root, file)}: computed imports are not part of M0 assembly`); return;
      }
      if (!specifier.text.startsWith('.')) return; // External library; no source aliases configured.
      const target = relative(root, resolve(dirname(file), specifier.text)).split(sep);
      const forbidden = owner[0] === 'platform' && (target[0] === 'cases' || target[0] === 'assembly')
        || owner[0] === 'cases' && (target[0] === 'assembly' || target[0] === 'cases' && target[1] !== owner[1]);
      if (forbidden) violations.push(`${relative(root, file)} → ${specifier.text}`);
    }
    function visit(node) {
      if ((ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) && node.moduleSpecifier) check(node.moduleSpecifier);
      if (ts.isCallExpression(node) && (node.expression.kind === ts.SyntaxKind.ImportKeyword || ts.isIdentifier(node.expression) && node.expression.text === 'require') && node.arguments[0]) check(node.arguments[0]);
      ts.forEachChild(node, visit);
    }
    visit(source);
  }
}
await inspect(root);
if (violations.length) { console.error('Frontend dependency boundary violation:\n' + violations.join('\n')); process.exitCode = 1; }
else console.log('Frontend platform/case/assembly dependency boundaries passed.');
