# Observability Tenancy (FR-11)

> Phase 2+3 closure, Lane A. Decided in OQ-05: `openchoreo.project` is the
> per-project filter key; SigNoz community edition suffices at this scale.
> No new instrumentation was added in this phase -- this document records
> where the key already flows and how to filter on it.

## The filter key

`openchoreo.project` is the single tenancy attribute for traces, metrics,
and logs. Companion keys: `openchoreo.component`, `openchoreo.environment`,
`openchoreo.runtime_namespace`.

## Where the attribute comes from (per signal)

- Traces: workload instrumentation sets `openchoreo.*` as OTLP resource
  attributes (see `seed-repos/hello-m2/main.go`: `OPENCHOREO_PROJECT` env
  -> `openchoreo.project`). The standalone collector forwards them
  untouched (`k8s_attributes` runs in passthrough mode for traces).
- Spanmetrics (FR-07): the connector copies span resource attributes onto
  the derived `spanmetrics_*` metric resource, so `openchoreo.project` is
  a queryable label on RED metrics.
- Logs (FR-05): the `filelog/k8s-pods` receiver lifts pod identity from the
  log file path; `k8s_attributes/logs` then promotes the pod labels
  `openchoreo.dev/project` -> `openchoreo.project` and
  `openchoreo.dev/component` -> `openchoreo.component` (label keys use
  slashes, which Kubernetes requires; attribute keys use dots).

## Where the filter is applied

- Dashboards (`observability/dashboards/hello-m2.json`,
  `observability/dashboards/platform.json`): each carries a `project`
  dashboard variable bound to the `openchoreo.project` attribute; panel
  filter expressions use `openchoreo.project IN $project`.
- Portal cards: `ObservabilityLinksCard` builds its live trace query as
  `service.name = '<component>' AND openchoreo.project = '<project>'`
  (the project clause is omitted when the entity lacks the annotation).

## Querying manually

SigNoz v5 query builder filter expression examples:

```
service.name = 'hello-m2' AND openchoreo.project = 'default'
openchoreo.component = 'hello-m2' AND body CONTAINS 'error'
```

ClickHouse (debugging):

```sql
SELECT count(*) FROM signoz_logs.distributed_logs_v2
WHERE resources_string['k8s.namespace.name'] = 'dp-default-default-development-f8e58905'
```

## Limits (community edition)

No row-level access control: any SigNoz user sees all projects' data. The
tenancy key is a filter convenience, not a security boundary. Multi-tenant
isolation is a Phase 4+ (NUC-era) question.
