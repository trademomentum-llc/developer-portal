/**
 * namespace-predictor.ts
 *
 * Pure, zero-dependency TypeScript port of the deterministic OpenChoreo
 * runtime namespace predictor.
 *
 * Reference implementation: tools/namespace-predictor/main.go
 * (PredictRuntimeNamespace)
 *
 * OpenChoreo computes the namespace in
 * internal/controller/project/integrations/kubernetes/namespace_handler.go as:
 *
 *   GenerateK8sNameWithLengthLimit(63, "dp", controlPlaneNs, projectName, environmentName)
 *
 * where GenerateK8sNameWithLengthLimit lives in
 * internal/dataplane/kubernetes/name.go.
 *
 * This file is a byte-for-byte semantic replica of that logic for ASCII
 * inputs (the only inputs Kubernetes object names allow in practice; all
 * openchoreo.dev/* annotation values are ASCII slugs). Known boundary:
 * upstream sanitizeName keeps Unicode letters (`unicode.IsLetter`) while
 * this port maps non-ASCII to '-', and JS/Go Unicode case-folding differ
 * (e.g. U+0130) -- non-ASCII parity is explicitly out of contract.
 *
 * Determinism guarantee: For any identical ASCII (c, p, e) triple the output
 * string is identical between the Go binary and this TS function, pinned by
 * namespace-predictor.test.ts (FR-37), which executes the Go binary.
 *
 * Primary test vector (verified against the live cluster):
 *   predictRuntimeNamespace("default", "default", "development")
 *   => "dp-default-default-development-f8e58905"
 *
 * This module contains no side effects, no network, no secrets.
 * It is safe for synchronous rendering in React entity cards.
 */

// ---------------------------------------------------------------------------
// Minimal, correct, zero-dependency SHA-256 (FIPS 180-2)
// Compact port suitable for browser + Node without external packages.
// ---------------------------------------------------------------------------

const K: number[] = [
  0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
  0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
  0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
  0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
  0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
  0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
  0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
  0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
];

function rightRotate(n: number, x: number): number {
  return (x >>> n) | (x << (32 - n));
}

function sha256(message: string): Uint8Array {
  const msgBuffer = new TextEncoder().encode(message);
  const len = msgBuffer.length;

  const bitLen = len * 8;
  // paddingLen spans the 0x80 terminator byte plus the zero fill, so the
  // total block size is len + paddingLen + 8 and lands on a 64-byte
  // boundary. (An earlier revision added the terminator twice and produced
  // non-SHA-256 digests; the Go/TS equality test pins this.)
  const paddingLen = (len % 64 < 56) ? (56 - (len % 64)) : (120 - (len % 64));
  const padded = new Uint8Array(len + paddingLen + 8);
  padded.set(msgBuffer);
  padded[len] = 0x80;

  const view = new DataView(padded.buffer);
  view.setUint32(padded.length - 4, bitLen >>> 0, false);

  let h0 = 0x6a09e667;
  let h1 = 0xbb67ae85;
  let h2 = 0x3c6ef372;
  let h3 = 0xa54ff53a;
  let h4 = 0x510e527f;
  let h5 = 0x9b05688c;
  let h6 = 0x1f83d9ab;
  let h7 = 0x5be0cd19;

  for (let i = 0; i < padded.length; i += 64) {
    const w = new Array<number>(64);

    for (let j = 0; j < 16; j++) {
      w[j] =
        (padded[i + j * 4] << 24) |
        (padded[i + j * 4 + 1] << 16) |
        (padded[i + j * 4 + 2] << 8) |
        (padded[i + j * 4 + 3]);
    }

    for (let j = 16; j < 64; j++) {
      const s0 =
        rightRotate(7, w[j - 15]) ^ rightRotate(18, w[j - 15]) ^ (w[j - 15] >>> 3);
      const s1 =
        rightRotate(17, w[j - 2]) ^ rightRotate(19, w[j - 2]) ^ (w[j - 2] >>> 10);
      w[j] = (w[j - 16] + s0 + w[j - 7] + s1) >>> 0;
    }

    let a = h0,
      b = h1,
      c = h2,
      d = h3,
      e = h4,
      f = h5,
      g = h6,
      h = h7;

    for (let j = 0; j < 64; j++) {
      const S1 = rightRotate(6, e) ^ rightRotate(11, e) ^ rightRotate(25, e);
      const ch = (e & f) ^ (~e & g);
      const temp1 = (h + S1 + ch + K[j] + w[j]) >>> 0;
      const S0 = rightRotate(2, a) ^ rightRotate(13, a) ^ rightRotate(22, a);
      const maj = (a & b) ^ (a & c) ^ (b & c);
      const temp2 = (S0 + maj) >>> 0;

      h = g;
      g = f;
      f = e;
      e = (d + temp1) >>> 0;
      d = c;
      c = b;
      b = a;
      a = (temp1 + temp2) >>> 0;
    }

    h0 = (h0 + a) >>> 0;
    h1 = (h1 + b) >>> 0;
    h2 = (h2 + c) >>> 0;
    h3 = (h3 + d) >>> 0;
    h4 = (h4 + e) >>> 0;
    h5 = (h5 + f) >>> 0;
    h6 = (h6 + g) >>> 0;
    h7 = (h7 + h) >>> 0;
  }

  const digest = new Uint8Array(32);
  const dv = new DataView(digest.buffer);
  dv.setUint32(0, h0, false);
  dv.setUint32(4, h1, false);
  dv.setUint32(8, h2, false);
  dv.setUint32(12, h3, false);
  dv.setUint32(16, h4, false);
  dv.setUint32(20, h5, false);
  dv.setUint32(24, h6, false);
  dv.setUint32(28, h7, false);
  return digest;
}

function bytesToHex(bytes: Uint8Array, length: number): string {
  let s = '';
  for (let i = 0; i < length; i++) {
    s += bytes[i].toString(16).padStart(2, '0');
  }
  return s;
}

// ---------------------------------------------------------------------------
// OpenChoreo name-generation helpers (mirrors internal/dataplane/kubernetes/name.go)
// ---------------------------------------------------------------------------

const HASH_LENGTH = 8;
const SEPARATOR = '-';

function sanitizeName(name: string): string {
  name = name.toLowerCase();
  let sanitized = '';
  for (const r of name) {
    if (/[a-z0-9\-.]/.test(r)) {
      sanitized += r;
    } else {
      sanitized += '-';
    }
  }
  return sanitized.replace(/^[\-.]+|[\-.]+$/g, '');
}

function ensureDnsSubdomainCompliance(name: string): string {
  return name.replace(/^[^a-z0-9]+/g, '').replace(/[^a-z0-9]+$/g, '');
}

function generateK8sNameWithLengthLimit(limit: number, ...names: string[]): string {
  const cleanedNames = names.map(sanitizeName);

  const fullName = names.join(SEPARATOR);
  const digest = sha256(fullName);
  const hashString = bytesToHex(digest, HASH_LENGTH / 2); // 4 bytes -> 8 hex chars

  const numberOfNames = cleanedNames.length;
  const numberOfSeparatorsInBaseName = numberOfNames - 1;
  let totalSeparatorLength = SEPARATOR.length * numberOfSeparatorsInBaseName;
  totalSeparatorLength += SEPARATOR.length;

  const maxBaseNameLength = limit - HASH_LENGTH - totalSeparatorLength;
  const maxPartLength = Math.floor(maxBaseNameLength / numberOfNames);
  const extraChars = maxBaseNameLength % numberOfNames;

  const truncatedNames: string[] = [];
  for (let i = 0; i < cleanedNames.length; i++) {
    let allocatedLength = maxPartLength;
    if (i < extraChars) {
      allocatedLength++;
    }
    const n = cleanedNames[i];
    truncatedNames.push(n.length > allocatedLength ? n.substring(0, allocatedLength) : n);
  }

  const baseName = truncatedNames.join(SEPARATOR);
  const finalName = `${baseName}${SEPARATOR}${hashString}`;
  return ensureDnsSubdomainCompliance(finalName);
}

// ---------------------------------------------------------------------------
// Public API — exact match to Go reference
// ---------------------------------------------------------------------------

/**
 * predictRuntimeNamespace
 *
 * Deterministic computation of the data-plane runtime namespace for a given
 * (controlPlaneNs, projectName, environmentName) triple.
 *
 * This function is the TypeScript semantic equivalent of the Go implementation
 * in tools/namespace-predictor/main.go.
 *
 * @param controlPlaneNs - Control plane namespace (usually "default")
 * @param projectName    - Logical project name from openchoreo.dev/project
 * @param environmentName- Environment name from the ReleaseBinding (e.g. "development")
 * @returns The fully normalized namespace string (max 63 chars)
 */
export function predictRuntimeNamespace(
  controlPlaneNs: string,
  projectName: string,
  environmentName: string,
): string {
  return generateK8sNameWithLengthLimit(63, 'dp', controlPlaneNs, projectName, environmentName);
}

// ---------------------------------------------------------------------------
// Embedded equivalence test vectors (same table as main_test.go and the
// FR-37 jest test). All expected values are outputs of the Go reference
// binary; vector 0 is additionally verified against the live
// k3d-openchoreo cluster.
// ---------------------------------------------------------------------------

const TEST_VECTORS: Array<{
  c: string;
  p: string;
  e: string;
  expected: string;
}> = [
  { c: 'default', p: 'default', e: 'development', expected: 'dp-default-default-development-f8e58905' },
  { c: 'default', p: 'hello-m2', e: 'development', expected: 'dp-default-hello-m2-development-bd0274a8' },
  { c: 'openchoreo-control', p: 'prod-api', e: 'production', expected: 'dp-openchoreo-co-prod-api-production-bf865e69' },
  { c: 'underscore_ns', p: 'my_project', e: 'prod_env', expected: 'dp-underscore-ns-my-project-prod-env-f1cc0757' },
  { c: 'long-control-ns', p: 'very-long-project-name-that-keeps-going', e: 'development', expected: 'dp-long-control--very-long-pro-development-121ccff5' },
];

export function runSelfTest(): { passed: number; failed: number; details: string[] } {
  let passed = 0;
  let failed = 0;
  const details: string[] = [];

  for (const v of TEST_VECTORS) {
    const actual = predictRuntimeNamespace(v.c, v.p, v.e);
    if (actual === v.expected) {
      passed++;
    } else {
      failed++;
      details.push(`FAIL: (${v.c},${v.p},${v.e}) => ${actual} (expected ${v.expected})`);
    }
  }
  return { passed, failed, details };
}

// Note: every expected value above is the output of the Go reference binary;
// vector #0 is authoritative and verified against the live k3d-openchoreo
// cluster. The FR-37 jest test (namespace-predictor.test.ts) re-executes the
// Go binary and asserts equality with this port for the same table.
