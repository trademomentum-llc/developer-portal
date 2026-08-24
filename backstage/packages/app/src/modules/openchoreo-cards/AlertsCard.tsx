import { InfoCard, Link } from '@backstage/core-components';
import { useApi, fetchApiRef } from '@backstage/core-plugin-api';
import { useEntity } from '@backstage/plugin-catalog-react';
import { Box, Typography } from '@material-ui/core';
import { useEffect, useState } from 'react';

// OQ-02 (Phase 2+3 closure): the portal is the pull surface for alerts in the
// local phase. This card lists the alert rules defined in SigNoz and their
// state, read through the /api/proxy/signoz backend proxy. Push channels
// (e.g. Gitea issues via a receiver) are a documented NUC-era upgrade.
//
// NFR-04: every state is honest -- unreachable proxy renders a not-wired
// notice, zero rules renders "no rules defined", never a fake green.

type AlertRule = {
  id: string;
  name: string;
  state: string;
};

function parseRules(resp: any): AlertRule[] {
  const items = resp?.data;
  if (!Array.isArray(items)) {
    throw new Error('unexpected SigNoz rules response shape');
  }
  return items.map((r: any) => ({
    id: String(r?.id ?? ''),
    name: String(r?.name ?? r?.alert ?? '(unnamed rule)'),
    state: String(r?.state ?? 'unknown'),
  }));
}

export const AlertsCard = () => {
  const { entity } = useEntity();
  const { fetch } = useApi(fetchApiRef);
  const serviceName = entity.metadata.name;

  const signozBase = 'http://localhost:3301';

  const [rules, setRules] = useState<AlertRule[] | null>(null);
  const [rulesError, setRulesError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetch('/api/proxy/signoz/v2/rules')
      .then(async res => {
        if (!res.ok) {
          throw new Error(`SigNoz proxy returned ${res.status}`);
        }
        const parsed = parseRules(await res.json());
        if (!cancelled) {
          setRules(parsed);
        }
      })
      .catch(err => {
        if (!cancelled) {
          setRulesError(err.message);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [fetch]);

  const firing = rules?.filter(r => r.state === 'firing') ?? [];

  return (
    <InfoCard title="Alerts (SigNoz pull)" variant="gridItem">
      <Box>
        {rulesError ? (
          <Typography variant="body2" style={{ color: 'orange' }}>
            Alert rules not wired through the portal proxy ({rulesError}).
            Open SigNoz directly when the 3301 port-forward is up.
          </Typography>
        ) : rules === null ? (
          <Typography variant="body2">loading...</Typography>
        ) : rules.length === 0 ? (
          <Typography variant="body2" style={{ opacity: 0.8 }}>
            No alert rules defined in SigNoz. Rules are codified in-repo and
            seeded by install-m3.sh (OQ-02); none matched this instance yet.
          </Typography>
        ) : (
          <>
            <Typography variant="body2" gutterBottom>
              {firing.length > 0
                ? `${firing.length} of ${rules.length} rules firing`
                : `${rules.length} rules defined, none firing`}
            </Typography>
            <Box component="ul" style={{ margin: 0, paddingLeft: 18 }}>
              {rules.map(r => (
                <li key={r.id || r.name}>
                  <Typography variant="caption">
                    {r.name} — {r.state}
                  </Typography>
                </li>
              ))}
            </Box>
          </>
        )}

        <Box mt={1}>
          <Link to={`${signozBase}/alerts`}>Manage alerts in SigNoz</Link>
        </Box>
        <Typography variant="caption" style={{ marginTop: 8, opacity: 0.7 }}>
          Pull surface for {serviceName} and platform alerts; evaluation happens
          in SigNoz, no portal-side actuator.
        </Typography>
      </Box>
    </InfoCard>
  );
};
