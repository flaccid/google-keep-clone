# Getting Started

This guide walks through deploying Google Keep Clone on **Kubernetes** (primary) and with **Docker Compose** (alternative). For local development workflow, see the README.

---

## Prerequisites

| Required | Kubernetes path | Docker Compose path |
|----------|----------------|-------------------|
| Git | ✅ | ✅ |
| Docker | ✅ | ✅ |
| kind / minikube | ✅ | — |
| Helm 3.8+ | ✅ | — |
| kubectl | ✅ | — |
| Ingress controller | ✅ (ingress-nginx) | — |
| OAuth2 credentials | ✅ (Google OAuth) | — |

---

# ☸️ Kubernetes (recommended)

## 1. Cluster

Create a local cluster with kind:

```bash
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 80
        hostPort: 80
      - containerPort: 443
        hostPort: 443
EOF
```

Or use minikube:

```bash
minikube start
minikube addons enable ingress
```

## 2. Install ingress-nginx

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
# Wait for it to be ready
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=120s
```

For minikube, use `minikube addons enable ingress` instead.

## 3. Configure OAuth2 credentials

Go to the [Google Cloud Console](https://console.cloud.google.com/apis/credentials), create an OAuth 2.0 Client ID (Web application), and set the authorized redirect URI to:

```
http://localhost/oauth2/callback
```

Save the **Client ID** and **Client Secret**.

## 4. Install the chart

```bash
# Generate a secure cookie secret
COOKIE_SECRET=$(openssl rand -base64 32 | tr -d '\n' | base64)

# Install
helm install google-keep-clone ./charts \
  --set ingress.host=localhost \
  --set ingress.className=nginx \
  --set ingress.tlsSecretName= \
  --set oauth2Proxy.redirectUrl=http://localhost/oauth2/callback \
  --set oauth2Proxy.cookieSecure=false \
  --set secrets.create.oauth2Proxy.clientId=YOUR_CLIENT_ID \
  --set secrets.create.oauth2Proxy.clientSecret=YOUR_CLIENT_SECRET \
  --set secrets.create.oauth2Proxy.cookieSecret="$COOKIE_SECRET"
```

> **Note:** `ingress.tlsSecretName` is set to empty because local clusters don't have TLS. For production, set it to your cert-manager secret name and keep `oauth2Proxy.cookieSecure: true`.

## 5. Access

Add an entry to `/etc/hosts` (for kind clusters):

```
127.0.0.1 localhost
```

Visit **http://localhost** in your browser. You'll be redirected to Google OAuth — sign in with an email listed in `oauth2Proxy.authorizedEmails`.

### Verify the deployment

```bash
# Check all pods are running
kubectl get pods -n google-keep-clone -w

# Port-forward if ingress isn't available (alternative access)
kubectl port-forward -n google-keep-clone svc/frontend 3000:3000
# Then visit http://localhost:3000
```

## Customisation

### Image tags

```bash
helm upgrade google-keep-clone ./charts \
  --set images.backend.tag=abc1234 \
  --set images.frontend.tag=abc1234
```

### Resource scaling

```bash
helm upgrade google-keep-clone ./charts \
  --set backend.replicas=3 \
  --set backend.hpa.minReplicas=3 \
  --set frontend.replicas=3
```

### Postgres storage

```bash
helm upgrade google-keep-clone ./charts \
  --set postgres.storageSize=10Gi
```

### Disable network policies or PDBs

```bash
helm upgrade google-keep-clone ./charts \
  --set networkPolicy.enabled=false \
  --set podDisruptionBudget.enabled=false
```

### Use existing secrets

```bash
helm upgrade google-keep-clone ./charts \
  --set secrets.existing.backend=my-backend-secret \
  --set secrets.existing.postgres=my-postgres-secret \
  --set secrets.existing.oauth2Proxy=my-oauth2-secret \
  --set secrets.create.backend.databaseUrl= \
  --set secrets.create.postgres.username= \
  --set secrets.create.postgres.password= \
  --set secrets.create.postgres.database= \
  --set secrets.create.oauth2Proxy.clientId= \
  --set secrets.create.oauth2Proxy.clientSecret= \
  --set secrets.create.oauth2Proxy.cookieSecret=
```

## Cleanup

```bash
helm uninstall google-keep-clone
kubectl delete pvc -l app.kubernetes.io/instance=google-keep-clone  # deletes postgres data
kind delete cluster  # or: minikube delete
```

---

# 🐳 Docker Compose

The Docker Compose setup is ideal for a quick local preview without Kubernetes. It runs the same images but **without authentication** (no OAuth2 proxy).

## 1. Start the services

```bash
docker compose up
```

This starts:
- **PostgreSQL 16** on `:5432`
- **Backend** (Go) on `:8080` — auto-migrates the DB schema on startup
- **Frontend** (Next.js) on `:3000` — proxies API calls to the backend via rewrites

## 2. Access

Visit **http://localhost:3000** in your browser.

No authentication is required — you have full access to all notes. The API is only reachable through the Next.js proxy on `localhost:3000`, so it's not exposed directly.

## 3. Rebuild images

```bash
docker compose build
docker compose up -d
```

## 4. Check logs

```bash
docker compose logs -f backend   # Backend logs
docker compose logs -f frontend  # Frontend logs
docker compose logs -f postgres  # Database logs
```

## 5. Stop and clean up

```bash
docker compose down          # Stop containers
docker compose down -v      # Stop + delete postgres volume
```

---

# Using the app

Once the app is running (either path), here are the key things you can do:

| Action | How |
|--------|-----|
| **Create a note** | Click "Take a note..." in the header, type a title and body, click Close |
| **Create a checklist** | Click the checklist icon in the editor, press Enter for new items |
| **Change colour** | Click the palette icon on a card, pick a solid colour or gradient theme |
| **Pin a note** | Click the pin icon on a card |
| **Archive a note** | Click the archive icon on a card |
| **Trash a note** | Click the trash icon on a card |
| **Search** | Type in the header search bar (300ms debounce) |
| **Manage labels** | Click "Edit labels" in the sidebar, or use the sidebar labels list |
| **Add a label** | Open a note, click the tag icon, check labels from the dropdown |
| **Rename a label** | Go to Edit labels, click the pencil icon on a label |
| **Toggle dark mode** | Click the settings icon in the header, toggle Dark theme |
| **Collapse sidebar** | Click the hamburger icon in the header |
| **OpenAPI docs** | Visit http://localhost:3000/openapi (or http://localhost/openapi) |

---

# API reference

All endpoints are at `/v1/...` and proxied through Next.js (no CORS required). Full details in the README, or visit `/openapi` for the interactive Swagger UI.

| Method | Path | Purpose |
|--------|------|---------|
| `POST` | `/v1/notes` | Create a note |
| `GET` | `/v1/notes` | List notes (`?search=term&filter=archived=true&pageSize=20&pageToken=abc`) |
| `GET` | `/v1/notes/{id}` | Get a note |
| `PATCH` | `/v1/notes/{id}` | Update a note (title, body, color, labels) |
| `DELETE` | `/v1/notes/{id}` | Permanently delete |
| `POST` | `/v1/notes/{id}:pin` | Pin |
| `POST` | `/v1/notes/{id}:unpin` | Unpin |
| `POST` | `/v1/notes/{id}:archive` | Archive |
| `POST` | `/v1/notes/{id}:unarchive` | Unarchive |
| `POST` | `/v1/notes/{id}:trash` | Trash |
| `POST` | `/v1/notes/{id}:restore` | Restore from trash |
| `GET` | `/v1/labels` | List labels |
| `POST` | `/v1/labels` | Create a label |
| `PATCH` | `/v1/labels/{id}` | Rename a label |
| `DELETE` | `/v1/labels/{id}` | Delete a label |
| `POST` | `/v1/notes/{id}/permissions:batchCreate` | Share a note |
| `POST` | `/v1/notes/{id}/permissions:batchDelete` | Unshare a note |
| `POST` | `/v1/notes/{id}/attachments` | Upload a file |
| `GET` | `/v1/notes/{id}/attachments/{aid}` | Download an attachment |

---

# Troubleshooting

### Backend can't connect to PostgreSQL

**Kubernetes:** Check the init container succeeded:

```bash
kubectl logs -n google-keep-clone deployment/backend -c wait-for-postgres
```

**Docker Compose:**

```bash
docker compose logs backend
```

Make sure the `postgres` service is healthy before `backend` starts (Docker Compose handles this with `depends_on: condition: service_healthy`).

### "401 Unauthorized" on Kubernetes

Your Google account email isn't in the authorised list. Either:

```bash
# Add your email
kubectl edit configmap -n google-keep-clone oauth2-proxy-emails
```

Or reinstall with:

```bash
--set oauth2Proxy.authorizedEmails={your@email.com}
```

### Frontend shows blank page or API errors

Check the browser console for network errors. The Next.js rewrites should proxy `/v1/*` to the backend:

```
Kubernetes:  http://backend:8080  (cluster DNS)
Docker Compose:  http://backend:8080  (Docker DNS)
```

Verify the backend is reachable from the frontend pod/container:

```bash
# Kubernetes
kubectl exec -n google-keep-clone deploy/frontend -- wget -qO- http://backend:8080/v1/notes

# Docker Compose
docker compose exec frontend wget -qO- http://backend:8080/v1/notes
```

### Helm install fails with "namespace not found"

The chart creates the namespace automatically. If you deleted it manually:

```bash
kubectl create namespace google-keep-clone
```

### Kind: ingress not working

Ensure the ingress-nginx controller is running and the extra port mappings were set when creating the cluster:

```bash
kubectl get pods -n ingress-nginx
curl -H "Host: localhost" http://localhost
```

If ingress still doesn't work, use port-forward as a fallback:

```bash
kubectl port-forward -n google-keep-clone svc/frontend 3000:3000
```
