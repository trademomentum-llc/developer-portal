# Runbook: member provisioning and the paved-path deploy token

Scope: FR-13 (scaffolded projects deploy through the full Score ->
platform-config loop) and FR-14/OQ-09 (admin-provisioned members; Gitea is
the source of truth, org.yaml is generated).

## PLATFORM_CONFIG_TOKEN (org-level Actions secret)

Every repository in the `openchoreo` org inherits the org-level Actions
secret `PLATFORM_CONFIG_TOKEN`. Scaffolded projects use it in their deploy
stage to commit the rendered OpenChoreo Component into
`openchoreo/platform-config/environments/dev/`.

- Verified live on Gitea 1.25.4: org secrets exist at
  `/api/v1/orgs/openchoreo/actions/secrets/*` and are injected into every
  org repo's workflows.
- Seeded once by `scripts/seed-gitea-repos.sh`, which creates a fresh
  `gitea_admin` token named `platform-config-org-secret` scoped to
  `write:repository` and stores it as the org secret.
- Gitea secrets are write-only over the API, so the script is
  create-if-absent. To rotate: DELETE the org secret via the API, then
  re-run `scripts/seed-gitea-repos.sh` (stale same-named tokens are pruned
  automatically before the new one is created).
- If the secret is absent, a scaffolded project's deploy stage still renders
  the Component but skips the platform-config commit with a loud log line
  (honest degradation, not a fake success).

## Provisioning a member

    ./scripts/provision-member.sh create <username> <email> ["Full Name"] [--team <slug>]

- Creates the Gitea user with a random temporary password and
  `must_change_password=true`; the password is printed exactly once -- hand
  it to the member out of band.
- Adds the user to the org team (default `members`, created on first use
  with unit-level write access to code/issues/pulls/actions).
- Regenerates `backstage/examples/org.yaml` from live Gitea state. Never
  hand-edit that file; fix the data in Gitea and re-sync:

    ./scripts/provision-member.sh sync

- Removing a member: remove their org membership first
  (`DELETE /api/v1/orgs/openchoreo/members/<username>`), then delete the
  user (`DELETE /api/v1/admin/users/<username>`), then `sync`. Gitea refuses
  to delete a user who still holds org membership (HTTP 422).

## Removing a scaffolded project

1. Delete the repo: `DELETE /api/v1/repos/openchoreo/<name>`.
2. Delete the catalog entity (and its location) via the Backstage catalog
   API.
3. Delete the rendered Component from
   `platform-config/environments/dev/<name>.yaml` (contents API); the
   platform-config kustomizations prune, and OpenChoreo tears down the pod.
