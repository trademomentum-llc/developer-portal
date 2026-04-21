# hello-m2

Demo application that proves the M2 pipeline. Push to `main` triggers CI,
which validates Score, builds the image, pushes to Gitea's OCI registry,
renders a Component, and commits to `platform-config/environments/dev/`.
