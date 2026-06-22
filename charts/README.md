# google-keep-clone

A Helm chart for deploying [Google Keep Clone](https://github.com/flaccid/google-keep-clone) — a Go backend with a Next.js frontend and PostgreSQL database, secured by OAuth2 Proxy.

## Quick Start

```bash
# Install with default values (dev credentials)
helm install google-keep-clone ./charts

# Override ingress host and OAuth2 credentials
helm install google-keep-clone ./charts \
  --set ingress.host=keep.example.com \
  --set secrets.create.oauth2Proxy.clientId=your-client-id \
  --set secrets.create.oauth2Proxy.clientSecret=your-client-secret \
  --set secrets.create.oauth2Proxy.cookieSecret=$(openssl rand -base64 32)
```

## Requirements

- Kubernetes 1.28+
- Helm 3.8+
- Ingress controller (e.g. ingress-nginx) for `ingress.className: nginx`
- Cert-manager (recommended) for TLS certificate provisioning via `ingress.tlsSecretName`

## Architecture

```
              ┌──────────────┐
              │  Ingress     │  keep.example.com
              │  (nginx)     │
              └──────┬───────┘
                     │
          ┌──────────┼──────────┐
          ▼          ▼          ▼
   ┌──────────┐ ┌────────┐ ┌──────┐
   │ Frontend │ │ Backend│ │OAuth2│
   │:3000     │ │:8080   │ │:4180 │
   └──────────┘ └───┬────┘ └──────┘
                    │
              ┌─────▼──────┐
              │ PostgreSQL │
              │:5432       │
              └────────────┘
```

Ingress routes:
- `/` → frontend (auth-protected via OAuth2 Proxy annotation)
- `/v1/*` → backend (auth-protected)
- `/oauth2/*` → OAuth2 Proxy (unauthenticated — handles the OAuth callback flow)

## Configuration

### Global

| Parameter | Default | Description |
|-----------|---------|-------------|
| `global.namespace` | `google-keep-clone` | Target namespace for all resources |

### Images

| Parameter | Default | Description |
|-----------|---------|-------------|
| `images.backend.repository` | `flaccid/google-keep-clone-backend` | Backend image repository |
| `images.backend.tag` | `latest` | Backend image tag |
| `images.backend.pullPolicy` | `Always` | Backend image pull policy |
| `images.frontend.repository` | `flaccid/google-keep-clone-frontend` | Frontend image repository |
| `images.frontend.tag` | `latest` | Frontend image tag |
| `images.frontend.pullPolicy` | `Always` | Frontend image pull policy |
| `images.postgres.repository` | `postgres` | PostgreSQL image repository |
| `images.postgres.tag` | `16-alpine` | PostgreSQL image tag |
| `images.postgres.pullPolicy` | `IfNotPresent` | PostgreSQL image pull policy |
| `images.oauth2Proxy.repository` | `quay.io/oauth2-proxy/oauth2-proxy` | OAuth2 Proxy image repository |
| `images.oauth2Proxy.tag` | `latest` | OAuth2 Proxy image tag |
| `images.oauth2Proxy.pullPolicy` | `IfNotPresent` | OAuth2 Proxy image pull policy |

### Ingress

| Parameter | Default | Description |
|-----------|---------|-------------|
| `ingress.host` | `keep.example.com` | Ingress hostname |
| `ingress.className` | `nginx` | Ingress class name |
| `ingress.tlsSecretName` | `keep-tls` | TLS certificate secret name |

### Backend

| Parameter | Default | Description |
|-----------|---------|-------------|
| `backend.replicas` | `2` | Number of replicas |
| `backend.runAsUser` | `10001` | Container securityContext user ID |
| `backend.runAsGroup` | `10001` | Container securityContext group ID |
| `backend.resources.requests.cpu` | `100m` | CPU request |
| `backend.resources.requests.memory` | `128Mi` | Memory request |
| `backend.resources.limits.cpu` | `500m` | CPU limit |
| `backend.resources.limits.memory` | `256Mi` | Memory limit |
| `backend.hpa.enabled` | `true` | Enable HPA |
| `backend.hpa.minReplicas` | `2` | Minimum HPA replicas |
| `backend.hpa.maxReplicas` | `10` | Maximum HPA replicas |
| `backend.hpa.cpuUtilization` | `70` | CPU utilization target (%) |

The backend also runs a `wait-for-postgres` init container that blocks until PostgreSQL is ready. It reads `POSTGRES_USER`, `POSTGRES_DB`, and `PGPASSWORD` from the postgres secret to authenticate the health check.

Attachments are stored on an `emptyDir` volume mounted at `/data/attachments`.
The `ATTACHMENT_STORE_DIR` env var is hardcoded to this path.

### Frontend

| Parameter | Default | Description |
|-----------|---------|-------------|
| `frontend.replicas` | `2` | Number of replicas |
| `frontend.runAsUser` | `10002` | Container securityContext user ID |
| `frontend.runAsGroup` | `10002` | Container securityContext group ID |
| `frontend.resources.requests.cpu` | `100m` | CPU request |
| `frontend.resources.requests.memory` | `128Mi` | Memory request |
| `frontend.resources.limits.cpu` | `500m` | CPU limit |
| `frontend.resources.limits.memory` | `256Mi` | Memory limit |
| `frontend.hpa.enabled` | `true` | Enable HPA |
| `frontend.hpa.minReplicas` | `2` | Minimum HPA replicas |
| `frontend.hpa.maxReplicas` | `10` | Maximum HPA replicas |
| `frontend.hpa.cpuUtilization` | `70` | CPU utilization target (%) |

### PostgreSQL

| Parameter | Default | Description |
|-----------|---------|-------------|
| `postgres.replicas` | `1` | Number of replicas (should remain 1 for PVC-backed statefulset) |
| `postgres.storageSize` | `1Gi` | Persistent volume claim storage size |
| `postgres.runAsUser` | `999` | Container securityContext user ID (postgres user) |
| `postgres.runAsGroup` | `999` | Container securityContext group ID |
| `postgres.fsGroup` | `999` | Pod fsGroup for PVC access |

### OAuth2 Proxy

| Parameter | Default | Description |
|-----------|---------|-------------|
| `oauth2Proxy.replicas` | `1` | Number of replicas |
| `oauth2Proxy.runAsUser` | `65534` | Container securityContext user ID (nobody) |
| `oauth2Proxy.runAsGroup` | `65534` | Container securityContext group ID |
| `oauth2Proxy.resources.requests.cpu` | `50m` | CPU request |
| `oauth2Proxy.resources.requests.memory` | `64Mi` | Memory request |
| `oauth2Proxy.resources.limits.cpu` | `200m` | CPU limit |
| `oauth2Proxy.resources.limits.memory` | `256Mi` | Memory limit |
| `oauth2Proxy.provider` | `google` | OAuth2 provider |
| `oauth2Proxy.redirectUrl` | `https://keep.example.com/oauth2/callback` | OAuth callback URL |
| `oauth2Proxy.cookieSecure` | `true` | Require secure cookies |
| `oauth2Proxy.httpAddress` | `0.0.0.0:4180` | HTTP listen address |
| `oauth2Proxy.authorizedEmails` | `[admin@example.com, redacted@example.com]` | List of authorized email addresses |

## Secrets

The chart supports two modes:

### Creating secrets from values

Set `secrets.create.*` values to have the chart generate Secret resources using `stringData` (values are not base64-encoded in the values file):

```yaml
secrets:
  create:
    backend:
      databaseUrl: "postgres://user:pass@postgres:5432/db?sslmode=disable"
    postgres:
      username: myuser
      password: mypass
      database: mydb
    oauth2Proxy:
      clientId: "xxxxxxxx.apps.googleusercontent.com"
      clientSecret: "GOCSPX-..."
      cookieSecret: "32-char-random-string"
```

### Using existing secrets

Set `secrets.existing.*` to the name of a pre-created Kubernetes Secret, and the chart will reference it via `secretKeyRef` instead:

```yaml
secrets:
  existing:
    backend: my-backend-secret
    postgres: my-postgres-secret
    oauth2Proxy: my-oauth2-secret
```

The expected keys in each secret are:
- **backend secret**: `database-url`
- **postgres secret**: `username`, `password`, `database`
- **oauth2-proxy secret**: `client-id`, `client-secret`, `cookie-secret`

If both `secrets.existing.*` and `secrets.create.*` are set, the `create` values take precedence.

## Conditional Resources

| Resource | Condition | Default |
|----------|-----------|---------|
| `HorizontalPodAutoscaler` (backend) | `backend.hpa.enabled` | `true` |
| `HorizontalPodAutoscaler` (frontend) | `frontend.hpa.enabled` | `true` |
| `Secret` (backend) | `secrets.create.backend.databaseUrl` is non-empty | created |
| `Secret` (postgres) | Any of `secrets.create.postgres.*` is non-empty | created |
| `Secret` (oauth2-proxy) | Any of `secrets.create.oauth2Proxy.*` is non-empty | not created (empty defaults) |
| `NetworkPolicy` (all 7) | `networkPolicy.enabled` | `true` |
| `PodDisruptionBudget` (all 3) | `podDisruptionBudget.enabled` | `true` |

## Network Policies

When `networkPolicy.enabled: true`, the following policies are applied:

| Policy | Effect |
|--------|--------|
| `default-deny-all` | Blocks all ingress and egress by default |
| `allow-backend-to-postgres` | PostgreSQL accepts connections from backend on port 5432 |
| `allow-frontend-to-backend` | Backend accepts connections from frontend on port 8080 |
| `allow-ingress-to-frontend` | Frontend accepts connections from any namespace on port 3000 |
| `allow-ingress-to-oauth2` | OAuth2 Proxy accepts connections from any namespace on port 4180 |
| `allow-backend-to-oauth2-auth` | OAuth2 Proxy accepts connections from backend on port 4180 |
| `allow-dns-egress` | All pods can reach kube-dns on UDP 53 |

## Pod Disruption Budgets

When `podDisruptionBudget.enabled: true`:

| PDB | Rule |
|-----|------|
| `backend` | `minAvailable: 1` |
| `frontend` | `minAvailable: 1` |
| `postgres` | `maxUnavailable: 0` |

## Security Contexts

All workloads run with restricted security contexts:

| Setting | Backend | Frontend | PostgreSQL | OAuth2 Proxy |
|---------|---------|----------|------------|--------------|
| `runAsNonRoot` | true | true | true | true |
| `runAsUser` | 10001 | 10002 | 999 | 65534 |
| `runAsGroup` | 10001 | 10002 | 999 | 65534 |
| `fsGroup` | — | — | 999 | — |
| `seccompProfile` | RuntimeDefault | RuntimeDefault | RuntimeDefault | RuntimeDefault |
| `allowPrivilegeEscalation` | false | false | false | false |
| `capabilities.drop` | ALL | ALL | ALL | ALL |
| `readOnlyRootFilesystem` | true | true | — | true |

## Service Accounts

Four dedicated `ServiceAccount` resources are created: `backend`, `frontend`, `postgres`, `oauth2-proxy`. Each deployment references its corresponding service account. No cluster-level RBAC is configured — this is intended for a single-namespace deployment with no Kubernetes API access requirements.

## Template Structure

```
charts/
├── Chart.yaml
├── values.yaml
├── README.md
└── templates/
    ├── _helpers.tpl                     # Label and selector helpers
    ├── namespace.yaml                   # Target namespace
    ├── service-accounts.yaml            # 4 service accounts
    ├── ingress.yaml                     # Main + OAuth2 ingress
    ├── network-policy.yaml              # 7 network policies
    ├── pod-disruption-budgets.yaml      # 3 PDBs
    ├── backend/
    │   ├── deployment.yaml              # Backend app + init container
    │   ├── service.yaml                 # ClusterIP on :8080
    │   ├── hpa.yaml                     # CPU-based autoscaling
    │   └── secret.yaml                  # Database URL
    ├── frontend/
    │   ├── deployment.yaml              # Frontend app
    │   ├── service.yaml                 # ClusterIP on :3000
    │   └── hpa.yaml                     # CPU-based autoscaling
    ├── postgres/
    │   ├── statefulset.yaml             # Single-replica statefulset
    │   ├── service.yaml                 # Headless/ClusterIP on :5432
    │   └── secret.yaml                  # DB credentials
    └── oauth2-proxy/
        ├── deployment.yaml              # OAuth2 Proxy
        ├── service.yaml                 # ClusterIP on :80→4180
        ├── secret.yaml                  # OAuth client credentials
        └── emails-configmap.yaml        # Authorized email list
```

## Uninstall

```bash
helm uninstall google-keep-clone
kubectl delete pvc -l app.kubernetes.io/instance=google-keep-clone  # postgres PVC persists by default
```
