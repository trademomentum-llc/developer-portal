import { InfoCard, Link } from '@backstage/core-components';
import { useApi, fetchApiRef } from '@backstage/core-plugin-api';
import { useEntity } from '@backstage/plugin-catalog-react';
import { Box, Button, Typography } from '@material-ui/core';
import { useEffect, useState } from 'react';

// Shape of one entry in Gitea's
// GET /api/v1/repos/{owner}/{repo}/actions/runs response. Verified live
// against Gitea 1.25.4: the endpoint returns
// {workflow_runs: [...], total_count: N}, most-recent-first, and ignores a
// limit query param, so the card caps the rendered list client-side. All
// fields optional: the card degrades honestly on partial data (NFR-03).
interface WorkflowRun {
  id?: number;
  run_number?: number;
  display_title?: string;
  event?: string;
  head_branch?: string;
  head_sha?: string;
  status?: string;
  conclusion?: string | null;
  started_at?: string;
  completed_at?: string;
  path?: string;
}

// The runs list is rendered most-recent-first as returned by the API;
// MAX_RUNS keeps the card compact regardless of total_count.
const MAX_RUNS = 10;

// One-line summary of a run's outcome: the conclusion when the run has
// finished (success/failure/cancelled), otherwise the live status
// (waiting/running/...).
function runOutcome(run: WorkflowRun): string {
  return run.conclusion ?? run.status ?? 'unknown';
}

// FR-35: the workflow dispatched by the in-portal action. `ci.yaml` is the
// M2 pipeline convention (every component repo's main CI workflow; the same
// workflow id smoke-actions.sh dispatches).
const DISPATCH_WORKFLOW = 'ci.yaml';
const DISPATCH_REF = 'main';

/**
 * CiRunsCard
 *
 * Engagement-plane surface: renders the component repo's recent Gitea Actions
 * CI runs (status/conclusion, short sha, event, timestamps) through the
 * authenticated gitea-actions proxy, with a per-run link into the Gitea UI.
 *
 * The repo is derived from the openchoreo.dev/component annotation (falling
 * back to the entity name) under the fixed owner `openchoreo`, matching how
 * the M2 CD loop names repos after their component.
 *
 * Every failure mode is an explicit, labeled not-wired state -- no placeholder
 * is ever styled as live data (NFR-03).
 */
export const CiRunsCard = () => {
  const { entity } = useEntity();
  const { fetch } = useApi(fetchApiRef);

  const annotations = entity.metadata.annotations ?? {};
  const component =
    annotations['openchoreo.dev/component'] || entity.metadata.name;
  const repoOwner = 'openchoreo';

  // Human-facing UI links; the card itself FETCHES through the authenticated
  // proxy below (same pattern as SecurityCard).
  const giteaBase = 'http://localhost:3333';
  const actionsUrl = `${giteaBase}/${repoOwner}/${component}/actions`;

  const [runs, setRuns] = useState<WorkflowRun[] | null>(null);
  const [totalCount, setTotalCount] = useState<number>(0);
  const [runsState, setRunsState] = useState<'loading' | 'absent' | 'error'>(
    'loading',
  );
  const [runsError, setRunsError] = useState<string | null>(null);

  // FR-35 dispatch control state. refreshIndex re-triggers the runs fetch
  // after a successful dispatch so the new run surfaces without a reload.
  const [dispatchState, setDispatchState] = useState<
    'idle' | 'dispatching' | 'dispatched' | 'error'
  >('idle');
  const [dispatchError, setDispatchError] = useState<string | null>(null);
  const [refreshIndex, setRefreshIndex] = useState(0);

  useEffect(() => {
    let cancelled = false;
    fetch(
      `/api/proxy/gitea-actions/repos/${repoOwner}/${component}/actions/runs`,
    )
      .then(async res => {
        if (!res.ok) {
          throw new Error(`Gitea proxy returned ${res.status}`);
        }
        const body = await res.json();
        const list = (body?.workflow_runs ?? []) as WorkflowRun[];
        if (!cancelled) {
          setTotalCount(body?.total_count ?? list.length);
          if (list.length === 0) {
            setRunsState('absent');
          } else {
            setRuns(list.slice(0, MAX_RUNS));
          }
        }
      })
      .catch(err => {
        if (!cancelled) {
          setRunsError(err.message);
          setRunsState('error');
        }
      });
    return () => {
      cancelled = true;
    };
  }, [fetch, component, refreshIndex]);

  // FR-35: in-portal workflow dispatch. Actuation is the documented OQ-15
  // exception for forge dispatch: the POST rides the same authenticated
  // gitea-actions proxy as the reads (the backend attaches the token), and
  // in production the entity surface sits behind the Wave-0 RBAC policy
  // (SecurityRbacPolicy), same as every other card on this page.
  const dispatchWorkflow = () => {
    setDispatchState('dispatching');
    setDispatchError(null);
    fetch(
      `/api/proxy/gitea-actions/repos/${repoOwner}/${component}/actions/workflows/${DISPATCH_WORKFLOW}/dispatches`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref: DISPATCH_REF }),
      },
    )
      .then(async res => {
        if (!res.ok) {
          throw new Error(`Gitea proxy returned ${res.status}`);
        }
        setDispatchState('dispatched');
        // Give Gitea a moment to register the run before refreshing.
        setTimeout(() => setRefreshIndex(i => i + 1), 3000);
      })
      .catch(err => {
        setDispatchError(err.message);
        setDispatchState('error');
      });
  };

  return (
    <InfoCard title="CI Runs" variant="gridItem">
      <Box>
        <Typography variant="body2">
          <strong>Component:</strong> {component} | <strong>Repo:</strong>{' '}
          {repoOwner}/{component}
        </Typography>

        <Box mt={2}>
          <Typography variant="body2">
            <strong>Recent CI runs (Gitea Actions):</strong>
          </Typography>
          {runsState === 'absent' ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              No CI runs recorded for this component
            </Typography>
          ) : runsState === 'error' ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              <span style={{ color: 'orange' }}>
                CI runs: not wired (gitea-actions proxy unavailable:{' '}
                {runsError})
              </span>
            </Typography>
          ) : runs === null ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              loading...
            </Typography>
          ) : (
            <>
              <Typography variant="body2" style={{ marginTop: 4 }}>
                showing {runs.length} of {totalCount} run
                {totalCount === 1 ? '' : 's'}
              </Typography>
              {runs.map(run => (
                <Box key={run.id ?? run.run_number} mt={1}>
                  <Typography variant="body2">
                    <Link
                      to={`${actionsUrl}/runs/${run.run_number ?? run.id}`}
                    >
                      #{run.run_number ?? run.id}
                    </Link>{' '}
                    <strong>{runOutcome(run)}</strong> ({run.status ?? 'unknown'}
                    ) <code>{(run.head_sha ?? '').slice(0, 7) || 'unknown'}</code>{' '}
                    {run.event ?? 'unknown'}
                    {run.head_branch ? ` on ${run.head_branch}` : ''}
                  </Typography>
                  <Typography variant="caption" style={{ opacity: 0.8 }}>
                    started {run.started_at ?? 'unknown'} | completed{' '}
                    {run.completed_at ?? 'n/a'}
                  </Typography>
                </Box>
              ))}
            </>
          )}
        </Box>

        <Box mt={2}>
          <Typography variant="body2">
            <strong>Dispatch workflow:</strong>
          </Typography>
          <Box mt={1}>
            <Button
              size="small"
              variant="outlined"
              color="primary"
              disabled={dispatchState === 'dispatching'}
              onClick={dispatchWorkflow}
            >
              {dispatchState === 'dispatching'
                ? 'Dispatching...'
                : `Dispatch ${DISPATCH_WORKFLOW} on ${DISPATCH_REF}`}
            </Button>
          </Box>
          {dispatchState === 'dispatched' ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              Dispatched {DISPATCH_WORKFLOW} on {DISPATCH_REF} -- the run list
              refreshes automatically
            </Typography>
          ) : null}
          {dispatchState === 'error' ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              <span style={{ color: 'orange' }}>
                Dispatch failed (not wired): {dispatchError}
              </span>
            </Typography>
          ) : null}
        </Box>

        <Box mt={2} display="flex" flexDirection="column" gridGap={4}>
          <Link to={actionsUrl}>All runs in Gitea Actions</Link>
        </Box>

        <Typography variant="caption" style={{ marginTop: 8, opacity: 0.7 }}>
          Live data from the Gitea Actions API (repos/{repoOwner}/{component}
          /actions/runs) through the authenticated gitea-actions proxy. Each
          entry links to the run in the Gitea UI; the run list reflects the
          repo's full CI history, most recent first. The dispatch action POSTs
          workflow_dispatch through the same proxy (backend-attached token);
          in production this surface is behind the Wave-0 RBAC policy.
        </Typography>
      </Box>
    </InfoCard>
  );
};
