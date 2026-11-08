/** @jest-environment node */
/* eslint-disable no-restricted-imports -- node-env equality test: it builds
   and executes the Go reference binary, so fs/os/path/child_process are
   required and never ship in the browser bundle. */
/**
 * namespace-predictor.test.ts
 *
 * FR-37: asserts that the TypeScript port (namespace-predictor.ts) produces
 * byte-for-byte identical output to the Go reference binary
 * (tools/namespace-predictor/main.go) for the canonical smoke-m3 vectors
 * (truncation and underscore normalization included).
 *
 * The Go binary is built once per test run; Go is a platform prerequisite
 * (smoke-m3.sh treats its absence as a failure too), so a missing toolchain
 * fails this test loudly rather than silently skipping the equality
 * guarantee.
 */
import { execFileSync, execSync } from 'child_process';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import { predictRuntimeNamespace } from './namespace-predictor';

// Same table as tools/namespace-predictor/main_test.go and
// scripts/smoke-m3.sh section 1. Vector 0 is verified against the live
// k3d-openchoreo cluster.
const VECTORS: Array<{
  name: string;
  c: string;
  p: string;
  e: string;
  expected: string;
}> = [
  {
    name: 'canonical (live-verified, must never change)',
    c: 'default',
    p: 'default',
    e: 'development',
    expected: 'dp-default-default-development-f8e58905',
  },
  {
    name: 'hello-m2 development',
    c: 'default',
    p: 'hello-m2',
    e: 'development',
    expected: 'dp-default-hello-m2-development-bd0274a8',
  },
  {
    name: 'part truncation',
    c: 'openchoreo-control',
    p: 'prod-api',
    e: 'production',
    expected: 'dp-openchoreo-co-prod-api-production-bf865e69',
  },
  {
    name: 'underscore normalization',
    c: 'underscore_ns',
    p: 'my_project',
    e: 'prod_env',
    expected: 'dp-underscore-ns-my-project-prod-env-f1cc0757',
  },
  {
    name: '63-char length limit with long parts',
    c: 'long-control-ns',
    p: 'very-long-project-name-that-keeps-going',
    e: 'development',
    expected: 'dp-long-control--very-long-pro-development-121ccff5',
  },
];

// backstage/packages/app/src/modules/openchoreo-cards -> repo root (6 up).
const REPO_ROOT = path.resolve(__dirname, '../../../../../..');
const PREDICTOR_DIR = path.join(REPO_ROOT, 'tools', 'namespace-predictor');
const BIN = path.join(
  os.tmpdir(),
  `namespace-predictor-test-${process.pid}${process.platform === 'win32' ? '.exe' : ''}`,
);

describe('predictRuntimeNamespace (TS port vs Go binary)', () => {
  beforeAll(() => {
    try {
      execSync('go version', { stdio: 'pipe' });
    } catch {
      throw new Error(
        'Go toolchain not found: the predictor equality test requires Go ' +
          '(platform prerequisite, same as smoke-m3.sh)',
      );
    }
    execSync(`go build -o "${BIN}" .`, { cwd: PREDICTOR_DIR, stdio: 'pipe' });
  }, 120000);

  afterAll(() => {
    fs.rmSync(BIN, { force: true });
  });

  for (const v of VECTORS) {
    it(`matches the Go binary: ${v.name}`, () => {
      const goOutput = execFileSync(BIN, [v.c, v.p, v.e], {
        encoding: 'utf8',
      }).trim();
      const tsOutput = predictRuntimeNamespace(v.c, v.p, v.e);
      // Sanity: the Go binary must still produce the pinned expected value
      // (guards against a broken binary silently agreeing with a broken port).
      expect(goOutput).toBe(v.expected);
      expect(tsOutput).toBe(goOutput);
    });
  }
});
