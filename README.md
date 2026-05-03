### wordpress-operator

A Kubernetes operator for managing WordPress deployments at scale.

A fork of the original [bitpoke/wordpress-operator](https://github.com/bitpoke/wordpress-operator), maintained independently. Releases are at [MatthewKennedy/wordpress-operator/releases](https://github.com/MatthewKennedy/wordpress-operator/releases). Minimum supported Kubernetes: 1.27.

## Goals

- Deploy scalable WordPress sites declaratively on Kubernetes.
- Best-practice rollouts: canary, slow rollout, rolling update strategy.
- Devops-friendly: Prometheus metrics, health probes, status conditions, no in-cluster shell access required.

## Architecture

Two pieces work together:

| Piece | What it is |
|---|---|
| **wordpress-operator** (this repo) | Watches `Wordpress` CRs and reconciles a Kubernetes `Deployment`, `Service`, `Secret`, and `PersistentVolumeClaim`s per site. Image: `public.ecr.aws/w0a8g6c1/wordpress-operator`. |
| **WordPress runtime** ([stack-runtimes](https://github.com/MatthewKennedy/stack-runtimes)) | The container image the operator deploys. Bundles nginx + PHP-FPM + WordPress core. Image: `ghcr.io/matthewkennedy/wordpress-runtime`. |

### How the operator and runtime interact

The operator does not configure WordPress directly. It builds a Pod spec from the `Wordpress` CR and lets the runtime image self-configure at startup using environment variables.

```
   ┌──────────────────────┐         creates           ┌────────────────────────────┐
   │ Wordpress CR         │ ─────────────────────────▶│  Deployment / Service /    │
   │  spec: routes, code, │                           │  PVCs / Secret             │
   │  media, env, ...     │                           │                            │
   └──────────────────────┘                           │  Pod = init containers     │
            │                                         │      + wordpress runtime   │
            │ env vars baked into Pod spec ───────────▶  (nginx + PHP-FPM)        │
            │                                         │                            │
            │ wp-cron-controller polls               │                            │
            │ http://<svc>/wp/wp-cron.php every 30s ▶│                            │
            │   (status: WPCronTriggering condition)  │                            │
            └─────────────────────────────────────────┘
```

**Init containers** (run once per pod start):
- `git` — clones a repo into the code volume (only if `spec.code.git` is set).
- `prepare-volumes` — chowns/permissions the code and media mount points.
- `install-wp` — runs the runtime's bootstrap if `spec.bootstrap` is set, creating the WordPress site on first launch.

**Main container** runs the runtime image. The runtime's entrypoint uses [Dockerize](https://github.com/jwilder/dockerize) templates to render nginx and PHP-FPM config at startup from environment variables. Key variables the operator sets:

| Variable | Source / purpose |
|---|---|
| `WP_HOME`, `WP_SITEURL` | Built from `spec.routes[0]` |
| `STACK_ROUTES` | Comma-separated `domain[/path]` list from `spec.routes` |
| `STACK_SITE_NAME`, `STACK_SITE_NAMESPACE` | Pod's CR name and namespace |
| `WP_CORE_DIRECTORY` | WordPress core path inside the code volume |
| `STACK_MEDIA_BUCKET` | When `spec.media.s3` or `spec.media.gcs` is set |
| `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | From `spec.env` (you wire these to your MySQL) |
| `POD_NAMESPACE`, `POD_NAME` | Downward API |

Anything in `spec.env` and `spec.envFrom` is appended verbatim — that's how you point the runtime at your DB, configure the nginx page cache (`STACK_PAGE_CACHE_*`), set the cluster-aware nginx resolver (`STACK_RESOLVER`), or pass WordPress salts.

**Volumes** mounted into the runtime container:

| Mount | Source | Default path |
|---|---|---|
| Code | `spec.code.{git,persistentVolumeClaim,hostPath,emptyDir}` | `/app/web/wp-content` |
| Media | `spec.media.{s3,gcs,persistentVolumeClaim,hostPath,emptyDir}` | `/app/web/wp-content/uploads` |
| Logs (Knative-style) | `emptyDir` | `/var/log` |

### What the operator does NOT do

- **MySQL.** Bring your own. Wire `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` via `spec.env`. The MySQL operator from Bitpoke or any external MySQL works.
- **Routing or TLS termination.** Point your Gateway API `HTTPRoute`, ingress controller, or reverse proxy at the generated `Service` (port `80` → container `8080`). The operator no longer manages `Ingress` as of v2.0.0.
- **Backup.** The runtime supports media backends (S3, GCS) for offsite storage; for the database, run your MySQL operator's backup features.

## Install

### CRD (optional — Helm installs it by default)

```shell
kubectl apply -f https://raw.githubusercontent.com/MatthewKennedy/wordpress-operator/main/config/crd/bases/wordpress.presslabs.org_wordpresses.yaml
```

### Operator chart (OCI)

```shell
helm install wordpress-operator oci://public.ecr.aws/w0a8g6c1/wordpress-operator \
  --version 3.0.0 \
  --namespace wordpress-operator \
  --create-namespace
```

## Deploying a site

```yaml
apiVersion: wordpress.presslabs.org/v1alpha1
kind: Wordpress
metadata:
  name: mysite
spec:
  replicas: 3

  routes:
    - domain: example.com
      path: /

  # image: ghcr.io/matthewkennedy/wordpress-runtime:6.9.1   # default if omitted

  # Code source — git, PVC, hostPath, or emptyDir (default).
  code:
    git:
      repository: https://github.com/example/site
      reference: main
      env:
        - name: SSH_RSA_PRIVATE_KEY
          valueFrom:
            secretKeyRef: { name: mysite, key: id_rsa }
    # persistentVolumeClaim: { ... }   # alternative
    # hostPath: { ... }                # alternative
    # emptyDir: {}                     # alternative (default)

  # Media storage — s3, gcs, PVC, hostPath, or emptyDir.
  media:
    persistentVolumeClaim:
      accessModes: [ReadWriteOnce]
      resources:
        requests:
          storage: 10Gi
    # s3: { bucket: ..., env: [...] }
    # gcs: { bucket: ..., env: [...] }

  # First-launch installation. Skipped if WordPress already exists in the code volume.
  bootstrap:
    env:
      - { name: WORDPRESS_BOOTSTRAP_USER,     valueFrom: { secretKeyRef: { name: mysite, key: USER     } } }
      - { name: WORDPRESS_BOOTSTRAP_PASSWORD, valueFrom: { secretKeyRef: { name: mysite, key: PASSWORD } } }
      - { name: WORDPRESS_BOOTSTRAP_EMAIL,    valueFrom: { secretKeyRef: { name: mysite, key: EMAIL    } } }
      - { name: WORDPRESS_BOOTSTRAP_TITLE,    valueFrom: { secretKeyRef: { name: mysite, key: TITLE    } } }

  # DB and runtime config — passed straight through to the WordPress container.
  env:
    - { name: DB_HOST,     value: mysite-mysql }
    - { name: DB_USER,     valueFrom: { secretKeyRef: { name: mysite-mysql, key: USER     } } }
    - { name: DB_PASSWORD, valueFrom: { secretKeyRef: { name: mysite-mysql, key: PASSWORD } } }
    - { name: DB_NAME,     valueFrom: { secretKeyRef: { name: mysite-mysql, key: DATABASE } } }

  envFrom:
    - prefix: "WORDPRESS_"
      secretRef: { name: mysite-salts }
```

## Routing

Point your Gateway API `HTTPRoute` (or ingress) at the operator-generated `Service`. The Service exposes:

- Port `80` → `wordpress` container port `8080` (HTTP)
- Port `9145` → Prometheus metrics endpoint

TLS termination, hostname rewriting, redirects, and rate limiting all live at the Gateway/ingress level, not in the operator.

## wp-cron

A second controller in the same operator binary polls each site's `/wp/wp-cron.php?doing_wp_cron` endpoint every 30 seconds, replacing the WordPress default of triggering cron on user requests. State is exposed on the CR as the `WPCronTriggering` status condition and printed in `kubectl get wordpress`.

## License

Apache 2.0. See [LICENSE](LICENSE).
