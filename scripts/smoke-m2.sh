#!/usr/bin/env bash
# scripts/smoke-m2.sh
set -e
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT}/scripts/lib/smoke-json.sh"
smoke_json_parse_args "$@"
smoke_json_begin m2

# FR-34: children append their own per-tool records to the same JSONL file.
if [ -n "${SMOKE_JSON_OUT}" ]; then
    export SMOKE_JSON_OUT
fi

for check in tofu actions flux score infracost gatekeeper openbao; do
    printf "[%s] " "$check"
    if "$ROOT/scripts/smoke-$check.sh"; then
        smoke_json_count pass
    else
        echo "FAIL"
        smoke_json_count fail
        exit 1
    fi
done

# Lane C static checks (FR-13/FR-14/FR-18): the scaffolder template must carry
# the full Score -> deploy loop and per-project TechDocs by construction.
# These are file-contract checks; the live scaffold e2e is the binding proof.
printf "[template-full-loop] "
TEMPLATE_DIR="$ROOT/backstage/examples/template/content"
if [ -f "$TEMPLATE_DIR/score.yaml" ] \
    && grep -q 'port: 3000' "$TEMPLATE_DIR/score.yaml" \
    && grep -q '/healthz' "$TEMPLATE_DIR/index.js" \
    && grep -q 'PORT || 3000' "$TEMPLATE_DIR/index.js" \
    && grep -q '^  deploy:' "$TEMPLATE_DIR/.gitea/workflows/ci.yaml" \
    && grep -q 'score2openchoreo --validate-only' "$TEMPLATE_DIR/.gitea/workflows/ci.yaml" \
    && grep -q 'commit-to-platform-config.sh dev' "$TEMPLATE_DIR/.gitea/workflows/ci.yaml" \
    && grep -q 'secrets.PLATFORM_CONFIG_TOKEN' "$TEMPLATE_DIR/.gitea/workflows/ci.yaml" \
    && grep -q 'Skipping the platform-config commit' "$TEMPLATE_DIR/.gitea/workflows/ci.yaml"; then
    echo "PASS"
    smoke_json_count pass
else
    echo "FAIL"
    smoke_json_count fail
    exit 1
fi

printf "[template-techdocs] "
if [ -f "$TEMPLATE_DIR/docs/index.md" ] \
    && [ -f "$TEMPLATE_DIR/mkdocs.yml" ] \
    && grep -q 'backstage.io/techdocs-ref: dir:.' "$TEMPLATE_DIR/catalog-info.yaml"; then
    echo "PASS"
    smoke_json_count pass
else
    echo "FAIL"
    smoke_json_count fail
    exit 1
fi

printf "[member-provisioning] "
if [ -x "$ROOT/scripts/provision-member.sh" ] \
    && bash -n "$ROOT/scripts/provision-member.sh" \
    && grep -q 'orgs/openchoreo/actions/secrets/PLATFORM_CONFIG_TOKEN' "$ROOT/scripts/seed-gitea-repos.sh"; then
    echo "PASS"
    smoke_json_count pass
else
    echo "FAIL"
    smoke_json_count fail
    exit 1
fi

echo "M2 smoke: all pass"
