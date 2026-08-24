import { InfoCard, Link } from '@backstage/core-components';
import { useApi, fetchApiRef } from '@backstage/core-plugin-api';
import { useEntity } from '@backstage/plugin-catalog-react';
import { Box, Typography } from '@material-ui/core';
import { useEffect, useState } from 'react';
import { predictRuntimeNamespace } from './namespace-predictor';

// FR-09: raw span row returned by the SigNoz v5 query_range API (requestType raw).
type TraceRow = {
  traceId: string;
  spanName: string;
  durationMs: number | null;
  hasError: boolean;
  timestamp: string;
};

// The v5 raw response nests rows under data.data.results[0].rows; each row
// carries the selected fields in a `data` map plus a `timestamp`. Parse
// defensively: any shape drift degrades to an honest not-wired message.
function parseTraceRows(resp: any): TraceRow[] {
  const results = resp?.data?.data?.results;
  if (!Array.isArray(results)) {
    throw new Error('unexpected SigNoz response shape');
  }
  const first = results.find((r: any) => Array.isArray(r?.rows));
  if (!first) {
    return [];
  }
  return first.rows.map((row: any) => {
    const d = row?.data ?? row ?? {};
    const durationNano = Number(d.duration_nano);
    const ts = row?.timestamp ?? d.timestamp;
    // SigNoz emits nanosecond-precision ISO strings; Date handles them, but
    // fall back to a truncated raw string rather than rendering Invalid Date.
    let when = '';
    if (ts) {
      const parsed = new Date(ts);
      when = Number.isNaN(parsed.getTime()) ? String(ts).slice(0, 19).replace('T', ' ') : parsed.toLocaleString();
    }
    return {
      traceId: String(d.trace_id ?? ''),
      spanName: String(d.name ?? '(unnamed span)'),
      durationMs: Number.isFinite(durationNano) ? durationNano / 1e6 : null,
      hasError: d.has_error === true || d.has_error === 'true',
      timestamp: when,
    };
  });
}

// Robust card for multi-angle observability links plus a live trace list.
// Uses the /api/proxy/signoz backend proxy (FR-09) and the runtime namespace
// from annotations (or predictor).
export const ObservabilityLinksCard = () => {
  const { entity } = useEntity();
  const { fetch } = useApi(fetchApiRef);

  const annotations = entity.metadata.annotations ?? {};
  const serviceName = entity.metadata.name;
  const env = annotations['openchoreo.dev/environment'] || 'development';

  // Single source of truth: the pure deterministic predictor (identical to Go reference).
  // The template annotation is retained only as a hint; the computed value is authoritative.
  const projectForNs = annotations['openchoreo.dev/project'] || 'unknown';
  const controlNs = annotations['openchoreo.dev/control-plane-namespace'] || 'default';
  const runtimeNs = annotations['openchoreo.dev/runtime-namespace-template'] || predictRuntimeNamespace(controlNs, projectForNs, env);

  // Dev-host SigNoz via the managed port-forward (start-backstage.sh, 3301 -> svc/signoz:8080);
  // team/ingress alternative: https://signoz.local via the Envoy gateway.
  const signozBase = 'http://localhost:3301'; // Configurable via app-config in future

  const [traces, setTraces] = useState<TraceRow[] | null>(null);
  const [traceError, setTraceError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    const end = Date.now();
    const start = end - 24 * 60 * 60 * 1000;
    // FR-11: per-tenancy filter is service.name + openchoreo.project. The
    // project clause is only added when the entity actually carries the
    // annotation, so unannotated entities degrade to a service-only view.
    const project = annotations['openchoreo.dev/project'];
    const filter = project
      ? `service.name = '${serviceName}' AND openchoreo.project = '${project}'`
      : `service.name = '${serviceName}'`;
    const body = {
      schemaVersion: 'v1',
      start,
      end,
      requestType: 'raw',
      compositeQuery: {
        queries: [
          {
            type: 'builder_query',
            spec: {
              name: 'A',
              signal: 'traces',
              filter: { expression: filter },
              selectFields: [
                { name: 'name', fieldContext: 'span' },
                { name: 'duration_nano', fieldContext: 'span' },
                { name: 'trace_id', fieldContext: 'span' },
                { name: 'has_error', fieldContext: 'span' },
              ],
              order: [
                { key: { name: 'timestamp', fieldContext: 'span' }, direction: 'desc' },
              ],
              limit: 10,
            },
          },
        ],
      },
    };
    fetch('/api/proxy/signoz/v5/query_range', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
      .then(async res => {
        if (!res.ok) {
          throw new Error(`SigNoz proxy returned ${res.status}`);
        }
        const rows = parseTraceRows(await res.json());
        if (!cancelled) {
          setTraces(rows);
        }
      })
      .catch(err => {
        if (!cancelled) {
          setTraceError(err.message);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [fetch, serviceName, annotations]);

  return (
    <InfoCard title="Observability (Multi-Angle)" variant="gridItem">
      <Box>
        <Typography variant="body2" gutterBottom>
          Service: <strong>{serviceName}</strong> | Env: <strong>{env}</strong>
        </Typography>

        <Typography variant="body2" style={{ marginBottom: 4 }}>
          Runtime namespace pattern: <code>{runtimeNs}</code>
        </Typography>

        <Typography variant="subtitle2" style={{ marginTop: 8 }}>
          Recent traces (last 24h, max 10)
        </Typography>
        {traceError ? (
          <Typography variant="body2" style={{ color: 'orange' }}>
            SigNoz not wired through the portal proxy ({traceError}). Deep links
            below still work when the 3301 port-forward is up.
          </Typography>
        ) : traces === null ? (
          <Typography variant="body2">loading...</Typography>
        ) : traces.length === 0 ? (
          <Typography variant="body2" style={{ opacity: 0.8 }}>
            No spans for service.name={serviceName} in the last 24h.
          </Typography>
        ) : (
          <Box component="ul" style={{ margin: 0, paddingLeft: 18 }}>
            {traces.map((t, i) => (
              <li key={`${t.traceId}-${i}`}>
                <Typography variant="caption">
                  {t.traceId ? (
                    <Link to={`${signozBase}/trace/${t.traceId}`}>{t.spanName}</Link>
                  ) : (
                    t.spanName
                  )}
                  {t.durationMs !== null ? ` — ${t.durationMs.toFixed(1)} ms` : ''}
                  {t.hasError ? ' — ERROR' : ''}
                  {t.timestamp ? ` — ${t.timestamp}` : ''}
                </Typography>
              </li>
            ))}
          </Box>
        )}

        <Box display="flex" flexDirection="column" gridGap={4} mt={1}>
          <Link to={`${signozBase}/traces?service=${serviceName}&env=${env}`}>
            Traces in SigNoz
          </Link>
          <Link to={`${signozBase}/metrics?service=${serviceName}`}>
            Metrics in SigNoz
          </Link>
          <Link to={`${signozBase}/logs?service=${serviceName}`}>
            Logs in SigNoz
          </Link>
          <Typography variant="caption" style={{ marginTop: 8, opacity: 0.7 }}>
            (Links will be enriched with exact runtime namespace once M3 collector + predictor integration lands)
          </Typography>
        </Box>
      </Box>
    </InfoCard>
  );
};
