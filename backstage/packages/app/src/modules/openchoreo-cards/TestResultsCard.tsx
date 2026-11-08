import { InfoCard, Link } from '@backstage/core-components';
import { useApi, fetchApiRef } from '@backstage/core-plugin-api';
import { useEntity } from '@backstage/plugin-catalog-react';
import { Box, Typography } from '@material-ui/core';
import { useEffect, useState } from 'react';

// One JSON Lines record as emitted by the smoke suites via SMOKE_JSON_OUT /
// --json (scripts/lib/smoke-json.sh, FR-34) and committed by CI to
// platform-config test-artifacts/<app>/<env>/latest.jsonl. All fields
// optional: the card degrades honestly on partial data (NFR-03).
interface SuiteResult {
  suite?: string;
  passed?: number;
  failed?: number;
  skipped?: number;
  ts?: string;
  git_sha?: string;
}

// Parses the JSONL artifact body; malformed lines are skipped rather than
// failing the whole card (a partially written artifact stays readable).
function parseSuiteResults(text: string): SuiteResult[] {
  const results: SuiteResult[] = [];
  for (const line of text.split('\n')) {
    const trimmed = line.trim();
    if (trimmed.length === 0) {
      continue;
    }
    try {
      results.push(JSON.parse(trimmed) as SuiteResult);
    } catch {
      // skip malformed line
    }
  }
  return results;
}

// Per-suite outcome summary, e.g. "3 passed, 0 failed, 1 skipped".
function suiteOutcome(r: SuiteResult): string {
  return `${r.passed ?? 0} passed, ${r.failed ?? 0} failed, ${
    r.skipped ?? 0
  } skipped`;
}

/**
 * TestResultsCard
 *
 * FR-36: renders the component's latest committed smoke-test artifact
 * (per-suite pass/fail/skip counts) from platform-config, fetched through the
 * authenticated gitea-actions proxy. The artifact is written by CI via
 * scripts/ci/commit-test-artifact.sh after running the smoke suites with
 * SMOKE_JSON_OUT.
 *
 * Every failure mode is an explicit, labeled not-wired state -- no placeholder
 * is ever styled as live data (NFR-03/NFR-04).
 */
export const TestResultsCard = () => {
  const { entity } = useEntity();
  const { fetch } = useApi(fetchApiRef);

  const annotations = entity.metadata.annotations ?? {};
  const env = annotations['openchoreo.dev/environment'] || 'development';
  const component =
    annotations['openchoreo.dev/component'] || entity.metadata.name;

  // Raw URL for the human-facing link (same pattern as CostCard/SecurityCard);
  // the card itself FETCHES through the authenticated proxy below.
  const giteaBase = 'http://localhost:3333';
  const artifactRawUrl = `${giteaBase}/openchoreo/platform-config/raw/branch/main/test-artifacts/${component}/${env}/latest.jsonl`;

  const [results, setResults] = useState<SuiteResult[] | null>(null);
  const [artifactState, setArtifactState] = useState<
    'loading' | 'absent' | 'error' | 'ok'
  >('loading');
  const [artifactError, setArtifactError] = useState<string | null>(null);
  const [artifactMeta, setArtifactMeta] = useState<{
    ts?: string;
    gitSha?: string;
  } | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetch(
      `/api/proxy/gitea-actions/repos/openchoreo/platform-config/contents/test-artifacts/${component}/${env}/latest.jsonl`,
    )
      .then(async res => {
        if (res.status === 404) {
          if (!cancelled) {
            setArtifactState('absent');
          }
          return;
        }
        if (!res.ok) {
          throw new Error(`Gitea proxy returned ${res.status}`);
        }
        const body = await res.json();
        // The Gitea contents API returns the file base64-encoded in .content.
        const text = atob(String(body.content ?? '').replace(/\s/g, ''));
        const parsed = parseSuiteResults(text);
        if (!cancelled) {
          if (parsed.length === 0) {
            setArtifactState('absent');
            return;
          }
          setResults(parsed);
          // Provenance of the artifact run: identical across records, so the
          // last record's ts/git_sha describes the run as a whole.
          const last = parsed[parsed.length - 1];
          setArtifactMeta({ ts: last.ts, gitSha: last.git_sha });
          setArtifactState('ok');
        }
      })
      .catch(err => {
        if (!cancelled) {
          setArtifactError(err.message);
          setArtifactState('error');
        }
      });
    return () => {
      cancelled = true;
    };
  }, [fetch, component, env]);

  return (
    <InfoCard title="Test Results" variant="gridItem">
      <Box>
        <Typography variant="body2">
          <strong>Component:</strong> {component} | <strong>Env:</strong> {env}
        </Typography>

        <Box mt={2}>
          <Typography variant="body2">
            <strong>Latest smoke results (CI artifact):</strong>
          </Typography>
          {artifactState === 'absent' ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              No test artifact yet -- CI has not committed smoke results for
              this component (FR-34 wiring lands with the pipeline step)
            </Typography>
          ) : artifactState === 'error' ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              <span style={{ color: 'orange' }}>
                Test results: not wired (gitea-actions proxy unavailable:{' '}
                {artifactError})
              </span>
            </Typography>
          ) : results === null ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              loading...
            </Typography>
          ) : (
            <>
              {results.map(r => (
                <Typography
                  variant="body2"
                  key={r.suite ?? 'unknown'}
                  style={{ marginTop: 4 }}
                >
                  <code>{r.suite ?? 'unknown'}</code>:{' '}
                  <strong>{suiteOutcome(r)}</strong>
                </Typography>
              ))}
              <Typography variant="caption" style={{ marginTop: 4, opacity: 0.8 }}>
                run at {artifactMeta?.ts ?? 'unknown'} | git_sha{' '}
                <code>{(artifactMeta?.gitSha ?? '').slice(0, 7) || 'unknown'}</code>
              </Typography>
            </>
          )}
        </Box>

        <Box mt={2} display="flex" flexDirection="column" gridGap={4}>
          <Link to={artifactRawUrl}>Latest test artifact (platform-config)</Link>
        </Box>

        <Typography variant="caption" style={{ marginTop: 8, opacity: 0.7 }}>
          Smoke suites emit machine-readable records (SMOKE_JSON_OUT, FR-34);
          CI commits them to platform-config under test-artifacts/{component}
          /{env}/latest.jsonl. The card reads that artifact through the
          authenticated gitea-actions proxy.
        </Typography>
      </Box>
    </InfoCard>
  );
};
