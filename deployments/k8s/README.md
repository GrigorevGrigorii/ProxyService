# Kubernetes manifests (Kustomize)

This folder mirrors the workloads defined in `deployments/aws-ecs-task-definitions`:

- `proxy-api-server`
- `admin-api-server`
- `background-scheduler`
- `background-worker`

Layout:

- `base/`: shared Deployments + Services (no environment-specific config)
- `overlays/cloud/`: cloud settings (RDS-style env + ALB-backed ingress for `proxy-service.sylu.net`)
- `overlays/local/`: local settings (Minikube + local Postgres/Redis via `host.minikube.internal`)

## Notes

- This setup provides a Kubernetes-based cloud deployment path as an alternative to AWS ECS task definitions.
- Cloud correctness has not been fully verified yet because the AWS free tier is over.

## Apply

- Cloud:

```sh
kubectl apply -k deployments/k8s/overlays/cloud
```

- Local (Minikube):

```sh
kubectl apply -k deployments/k8s/overlays/local
```

## What you must customize

- Secrets:
  - Replace `overlays/*/secrets.example.yaml` with a real secret manifest (recommended: keep it out of git).
- Cloud ALB specifics:
  - Ensure AWS Load Balancer Controller is installed.
  - Add TLS/certificate annotations in `overlays/cloud/ingress.yaml` if you need HTTPS.

## Local access options (Minikube)

- Ingress (requires ingress controller):
  - Enable NGINX ingress in Minikube: `minikube addons enable ingress` 
  - Start minicube tunelling: `minikube tunnel`
  - Map host to local IP in `/etc/hosts`:
    - `127.0.0.1 proxy-service.local`
  - Open:
    - `http://proxy-service.local/api/proxy/ping`
    - `http://proxy-service.local/api/admin/ping`
