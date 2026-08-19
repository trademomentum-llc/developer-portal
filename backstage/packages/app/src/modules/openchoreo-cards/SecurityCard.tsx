import { InfoCard, Link } from '@backstage/core-components';
import { useApi, fetchApiRef } from '@backstage/core-plugin-api';
import { useEntity } from '@backstage/plugin-catalog-react';
import { Box, Typography } from '@material-ui/core';
import { useEffect, useState } from 'react';
import { fetchConstraint } from './gatekeeper';

// Shape of the combined security scan artifact committed to platform-config
// by the CI security gates (SEC-PLANE-WAVE0-TECH-001 section 3.3). All fields
// optional: the card degrades honestly on partial data (NFR-03).
interface SecurityArtifact {
  artifact_type?: string;
  git_sha?: string;
  run_id?: string;
  generated_at?: string;
  gate?: {
    severity_threshold?: string[];
    status?: string;
    suppressions?: string[];
  };
  results?: {
    trivy_fs?: { report?: any };
    trivy_image?: { report?: any };
    osv?: { report?: any };
  };
}

// The Wave 0 constraint set (SEC-PLANE-WAVE0-TECH-001 section 5.4), same as
// PolicyCard. `resource` is the CRD plural used in the API path.
const CONSTRAINTS = [
  { kind: 'C1PlatformAddonsMainProtected', resource: 'c1platformaddonsmainprotected', name: 'c1-enforce' },
  { kind: 'C2ScoreSchemaValid', resource: 'c2scoreschemavalid', name: 'c2-enforce' },
  { kind: 'C3InfracostDelta', resource: 'c3infracostdelta', name: 'c3-enforce' },
];

interface ConstraintState {
  kind: string;
  name: string;
  totalViolations: number;
}

// Counts trivy vulnerabilities by Severity across every Result of a report.
function countTrivyBySeverity(report: any): Record<string, number> {
  const counts: Record<string, number> = {};
  for (const r of report?.Results ?? []) {
    for (const v of r?.Vulnerabilities ?? []) {
      const sev = v?.Severity ?? 'UNKNOWN';
      counts[sev] = (counts[sev] ?? 0) + 1;
    }
  }
  return counts;
}

// Counts osv-scanner findings. OSV entries only carry a severity label when
// the source database provides one (database_specific.severity); everything
// else is bucketed UNKNOWN rather than dropped.
function countOsvBySeverity(report: any): Record<string, number> {
  const counts: Record<string, number> = {};
  for (const r of report?.results ?? []) {
    for (const p of r?.packages ?? []) {
      for (const v of p?.vulnerabilities ?? []) {
        const sev = v?.database_specific?.severity ?? 'UNKNOWN';
        counts[sev] = (counts[sev] ?? 0) + 1;
      }
    }
  }
  return counts;
}

// Renders severity buckets in a stable order; an empty map is "no findings".
function formatSeverityCounts(counts: Record<string, number>): string {
  const order = ['CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'UNKNOWN'];
  const parts = order
    .filter(s => counts[s] !== undefined)
    .map(s => `${s}: ${counts[s]}`);
  for (const key of Object.keys(counts)) {
    if (!order.includes(key)) {
      parts.push(`${key}: ${counts[key]}`);
    }
  }
  return parts.length > 0 ? parts.join(', ') : 'no findings';
}

/**
 * SecurityCard
 *
 * FR-04 (SEC-PLANE-WAVE0-TECH-001 section 4): renders the component's latest
 * CI security scan artifact (Trivy + OSV-Scanner gate result, committed to
 * platform-config per section 3) and the live Gatekeeper constraint violation
 * totals (section 5's data path).
 *
 * Every failure mode is an explicit, labeled not-wired state -- no placeholder
 * is ever styled as live data (NFR-03).
 */
export const SecurityCard = () => {
  const { entity } = useEntity();
  const { fetch } = useApi(fetchApiRef);

  const annotations = entity.metadata.annotations ?? {};
  const env = annotations['openchoreo.dev/environment'] || 'development';
  const component =
    annotations['openchoreo.dev/component'] || entity.metadata.name;

  // Raw URL for the human-facing link (same pattern as CostCard's artifact
  // link); the card itself FETCHES through the authenticated proxy below.
  const giteaBase = 'http://localhost:3333';
  const artifactRawUrl = `${giteaBase}/openchoreo/platform-config/raw/branch/main/security-artifacts/${component}/${env}/latest.json`;

  const [artifact, setArtifact] = useState<SecurityArtifact | null>(null);
  const [artifactState, setArtifactState] = useState<
    'loading' | 'absent' | 'error' | 'ok'
  >('loading');
  const [artifactError, setArtifactError] = useState<string | null>(null);
  const [constraints, setConstraints] = useState<ConstraintState[] | null>(
    null,
  );
  const [violationsError, setViolationsError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetch(
      `/api/proxy/gitea-actions/repos/openchoreo/platform-config/contents/security-artifacts/${component}/${env}/latest.json`,
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
        const decoded = JSON.parse(
          atob(String(body.content ?? '').replace(/\s/g, '')),
        ) as SecurityArtifact;
        if (!cancelled) {
          setArtifact(decoded);
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

  useEffect(() => {
    let cancelled = false;
    Promise.all(
      CONSTRAINTS.map(async c => {
        const res = await fetchConstraint(fetch, c.resource, c.name);
        if (!res.ok) {
          throw new Error(`kubernetes proxy returned ${res.status}`);
        }
        const body = await res.json();
        return {
          kind: c.kind,
          name: c.name,
          totalViolations: body?.status?.totalViolations ?? 0,
        };
      }),
    )
      .then(states => {
        if (!cancelled) {
          setConstraints(states);
        }
      })
      .catch(err => {
        if (!cancelled) {
          setViolationsError(err.message);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [fetch]);

  return (
    <InfoCard title="Security" variant="gridItem">
      <Box>
        <Typography variant="body2">
          <strong>Component:</strong> {component} | <strong>Env:</strong> {env}
        </Typography>

        <Box mt={2}>
          <Typography variant="body2">
            <strong>Latest security scan (CI gate artifact):</strong>
          </Typography>
          {artifactState === 'absent' ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              No scan artifact yet -- the security pipeline has not committed
              one for this component
            </Typography>
          ) : artifactState === 'error' ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              <span style={{ color: 'orange' }}>{artifactError}</span>
            </Typography>
          ) : artifact === null ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              loading...
            </Typography>
          ) : (
            <>
              <Typography variant="body2" style={{ marginTop: 4 }}>
                <strong>Gate:</strong> {artifact.gate?.status ?? 'unknown'}{' '}
                (threshold:{' '}
                {artifact.gate?.severity_threshold?.join('/') || 'unknown'})
              </Typography>
              <Typography variant="body2">
                <strong>trivy fs:</strong>{' '}
                {formatSeverityCounts(
                  countTrivyBySeverity(artifact.results?.trivy_fs?.report),
                )}
              </Typography>
              <Typography variant="body2">
                <strong>trivy image:</strong>{' '}
                {formatSeverityCounts(
                  countTrivyBySeverity(artifact.results?.trivy_image?.report),
                )}
              </Typography>
              <Typography variant="body2">
                <strong>osv:</strong>{' '}
                {formatSeverityCounts(
                  countOsvBySeverity(artifact.results?.osv?.report),
                )}
              </Typography>
              <Typography variant="body2">
                <strong>git_sha:</strong>{' '}
                <code>{artifact.git_sha ?? 'unknown'}</code>{' '}
                <strong>generated_at:</strong>{' '}
                <code>{artifact.generated_at ?? 'unknown'}</code>
              </Typography>
            </>
          )}
        </Box>

        <Box mt={2}>
          <Typography variant="body2">
            <strong>Gatekeeper constraint violations:</strong>
          </Typography>
          {violationsError ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              Gatekeeper violations: not wired (kubernetes proxy unavailable:{' '}
              {violationsError})
            </Typography>
          ) : constraints === null ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              loading...
            </Typography>
          ) : (
            constraints.map(c => (
              <Typography variant="body2" key={c.name} style={{ marginTop: 4 }}>
                <code>{c.name}</code> ({c.kind}): {c.totalViolations}{' '}
                violation{c.totalViolations === 1 ? '' : 's'}
              </Typography>
            ))
          )}
        </Box>

        <Box mt={2} display="flex" flexDirection="column" gridGap={4}>
          <Link to={artifactRawUrl}>
            Latest scan artifact (platform-config)
          </Link>
        </Box>

        <Typography variant="caption" style={{ marginTop: 8, opacity: 0.7 }}>
          The scan artifact is assembled on every push by the CI security gates
          (Trivy + OSV-Scanner) and committed to platform-config; it always
          describes the last PASSING push -- a failed gate stops the pipeline
          before the artifact is written. Constraint violations are read live
          from each constraint's .status through the kubernetes proxy.
        </Typography>
      </Box>
    </InfoCard>
  );
};
