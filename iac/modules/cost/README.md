# M4 Cost Visibility Module

OpenTofu module that deploys a minimal Prometheus + OpenCost stack for actual cluster cost allocation.

## Usage

```bash
cd iac
tofu init -reconfigure
tofu plan -target=module.cost
tofu apply -target=module.cost -auto-approve
```

## Resources

- `helm_release.prometheus` in namespace `opencost`
- `helm_release.opencost` in namespace `opencost`

## Notes

- Prometheus is configured without persistence and with short retention for local development.
- OpenCost is pointed at the in-cluster Prometheus service.
