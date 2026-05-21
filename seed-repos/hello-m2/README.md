# hello-m2

Demo application that proves the M2 pipeline. Push to `main` triggers CI,
which validates Score, builds the image, pushes to the in-cluster local
registry, renders OpenChoreo resources, and commits to
`platform-config/environments/dev/`.
