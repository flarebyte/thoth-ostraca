import { expect, test } from 'bun:test';
import * as path from 'node:path';
import { fileURLToPath } from 'node:url';
import { buildBinary, projectRoot, runThoth, saveOutputs } from './helpers';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

test('thoth version prints dev', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const run = runThoth(bin, ['version'], root);
  // Save outputs for inspection; temp/ is git-ignored
  saveOutputs(root, 'version', run);
  expect(run.status).toBe(0);
  expect(run.stdout).toBe('thoth dev\n');
});
