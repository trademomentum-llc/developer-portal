# Gitea Actions runner module

Creates `gitea-runners` namespace, provisions an ExternalSecret pulling the
runner registration token from openbao, and installs the `act-runner` helm
chart. Runner is configured with the `ubuntu-latest` label per project
convention (runner label is self-hosted, label name matches operator muscle
memory).

| Input | Purpose |
|---|---|
| gitea_url | In-cluster Gitea URL the runner registers against |
