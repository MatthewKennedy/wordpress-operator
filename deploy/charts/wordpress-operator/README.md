# wordpress-operator

Helm chart for [wordpress-operator](https://github.com/MatthewKennedy/wordpress-operator), a Kubernetes operator for managing WordPress deployments at scale.

## Install

```sh
helm install wordpress-operator oci://public.ecr.aws/w0a8g6c1/wordpress-operator \
  --version 2.0.1 \
  --namespace wordpress-operator \
  --create-namespace
```

## Configuration

| Parameter | Description | Default |
| --- | --- | --- |
| `replicaCount` | Replicas for the controller | `1` |
| `image.repository` | Controller image repository | `public.ecr.aws/w0a8g6c1/wordpress-operator` |
| `image.pullPolicy` | Controller image pull policy | `IfNotPresent` |
| `image.tag` | Controller image tag (defaults to chart `appVersion`) | `""` |
| `imagePullSecrets` | Controller image pull secrets | `[]` |
| `podAnnotations` | Extra pod annotations | `{}` |
| `podSecurityContext` | Pod security context. `65532` is the nonroot UID/GID in the distroless base image | `{runAsNonRoot: true, runAsUser: 65532, runAsGroup: 65532, fsGroup: 65532}` |
| `securityContext` | Container security context | `{}` |
| `resources` | Container resource requests/limits | `{}` |
| `nodeSelector` | Pod nodeSelector | `{}` |
| `tolerations` | Pod tolerations | `[]` |
| `affinity` | Pod node affinity | `{}` |
| `extraArgs` | Extra args passed to the controller (see CLI flags) | `[]` |
| `extraEnv` | Extra env vars passed to the controller | `{}` |
| `rbac.create` | Whether to create the RBAC ServiceAccount, ClusterRole, and ClusterRoleBinding | `true` |
| `serviceAccount.create` | Whether to create a ServiceAccount | `true` |
| `serviceAccount.annotations` | Annotations on the ServiceAccount | `{}` |
| `serviceAccount.name` | Name of the ServiceAccount (auto-generated if empty) | `""` |
