# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
### Changed
### Removed
### Fixed

## [2.1.0] - 2026-05-03
### Fixed
 * **B1: Reconcile thrash on `FieldRef.APIVersion`.** The init-container env vars (`POD_NAMESPACE`, `POD_NAME`) constructed `corev1.ObjectFieldSelector` with no `APIVersion`, leaving Go's zero-value `""`. The apiserver defaults it to `"v1"` server-side, so the syncer detected a phantom diff every reconcile and tried (and failed) to patch it back. Set `APIVersion: "v1"` explicitly. (`pkg/internal/wordpress/pod_template.go`)
 * **B2: `KubeAPIWarningLogger: unknown field "spec.code.metadata.creationTimestamp"`.** `CodeVolumeSpec` and `MediaVolumeSpec` embedded `metav1.ObjectMeta`, which exposes 14+ fields none of which the operator reads except `Labels` and `Annotations`. Replaced the embed with a small `VolumeMetadata` struct containing only those two fields. JSON shape (`spec.code.metadata.labels`, `spec.code.metadata.annotations`) is unchanged — anyone using only those keys is unaffected.
 * **B3: wp-cron-controller log spam.** Every wp-cron poll against an unreachable WordPress pod logged a full-stack-trace error. The `WPCronTriggering` status condition already records this state persistently. Dropped the per-poll error log; suppressed status-update logs for expected optimistic-concurrency conflicts. (`pkg/controller/wp-cron/wpcron_controller.go`)
### Notes
 * The B2 fix is a narrow schema change. The CRD no longer accepts `spec.code.metadata.{name,namespace,uid,creationTimestamp,...}` — but those fields were never functional. Functional fields (`labels`, `annotations`) are unaffected.

## [2.0.2] - 2026-05-03
### Changed
 * Helm chart `appVersion` is now auto-set by the Makefile prepare target (`yq '.appVersion = "v$(IMAGE_TAG)"'`). No more manual bumps in source `Chart.yaml` per release.
 * GitHub Actions release workflow now triggers on `v[0-9]+.[0-9]+.[0-9]+` tags (was bare-numeric). Releases need only the v-prefixed tag — no more dual-tag-push dance. The TAG-extraction step strips the `v` prefix so downstream image/chart versioning is unchanged.
 * Bumped `actions/setup-go@v4` → `@v5` (Node.js 20 deprecation lands 2026-06-02).
 * Bumped operator base image from `gcr.io/distroless/static-debian10` (Buster, EOL 2024-06) to `gcr.io/distroless/static-debian12`. New digest pinned.
 * Bumped default `WordpressRuntimeImage` from `docker.io/bitpoke/wordpress-runtime:5.8.2` to `ghcr.io/matthewkennedy/wordpress-runtime:6.9.1`.
 * Bumped default `GitCloneImage` from `docker.io/library/buildpack-deps:stretch-scm` (Stretch, EOL 2022-06) to `bookworm-scm`.
 * Stripped descriptions from generated CRDs (`controller-gen ... crd:maxDescLen=0`). The CRD shrinks from 478 KB to 151 KB and now fits under the 256 KB `kubectl apply` annotation limit. Users can drop the `--server-side` workaround from the install instructions.
### Removed
 * Deleted dead Drone CI files (`.drone.yml`, `build/.drone.yml`); CI lives in GitHub Actions.
 * Deleted `pkg/webhook/` (single empty file kept for kubebuilder-v1 folder structure compliance; obsolete under kubebuilder v3).

## [2.0.1] - 2026-05-03

## [2.0.1] - 2026-05-03
### Changed
 * Helm chart defaults updated to point at `public.ecr.aws/w0a8g6c1/wordpress-operator`. Bare `helm install oci://public.ecr.aws/w0a8g6c1/wordpress-operator --version 2.0.1` now works without `--set` overrides.
 * Chart `appVersion` set to `v2.0.1` so the deployment template's `image.tag` fallback resolves to a published image tag.
 * Chart `version` bumped from `0.0.0` to `2.0.1` (chart and app versions now track together).
 * Chart metadata (description, home, sources, keywords) rebranded for the detached fork.
 * Makefile `DOCKER_REGISTRY` default changed to `public.ecr.aws/w0a8g6c1`.
 * Bumped documented minimum Kubernetes version to 1.27 (matches controller-runtime v0.18 official support).
### Removed
 * Dropped `bitpoke` references from the chart `Chart.yaml`, `values.yaml`, and chart `README.md`.

## [2.0.0] - 2026-05-03
### Changed
 * Bump Kubernetes client libraries to v0.30.6 (`k8s.io/api`, `k8s.io/apimachinery`, `k8s.io/client-go`)
 * Bump `sigs.k8s.io/controller-runtime` to v0.18.5; rewrite controller setup to use the `ctrl.NewControllerManagedBy(...).For(...).Owns(...)` builder API
 * Bump `presslabs/controller-util` to v0.15.0 (imports moved under `pkg/`)
 * Bump Ginkgo to v2 (`github.com/onsi/ginkgo/v2`); table tests no longer require the `extensions/table` package
 * Switch logger from `klogr` to controller-runtime's `log/zap`
 * Bump kubebuilder envtest assets to 1.30.6
 * Bump golangci-lint timeout to 5m (large dep upgrade pushed cold-cache lint past the 1m default)
 * `HomeURL` always returns `https://` — TLS/scheme is now terminated upstream by the Gateway/HTTPRoute, not the operator
### Removed
 * **BREAKING:** Operator no longer manages Ingress. Routing is expected to be configured externally (Gateway API HTTPRoute, ingress controllers, etc.) targeting the WordPress `Service`. The following CR fields are removed: `spec.tlsSecretRef`, `spec.ingressAnnotations`, `spec.skipIngress`. The `--ingress-class` CLI flag is removed. The `networking.k8s.io/ingresses` RBAC rule is removed.
 * **BREAKING:** Drop the `WATCH_NAMESPACE` env var (previously parsed but not wired into the manager since the controller-runtime upgrade). The operator now always runs cluster-wide. Was also documented as `SCOPED_NAMESPACE` in the 0.12.2 changelog — that name was never wired up either.
 * Drop GCP auth plugin import (`k8s.io/client-go/plugin/pkg/client/auth/gcp`); no longer needed in newer client-go
 * Drop unused `runtime.Scheme` field from reconcilers (defaulting now handled by `wp.SetDefaults()`)
 * Drop unused `SecretRef` type from the API package

## [0.12.2] - 2023-05-23
### Changed
 * Minimum required Kubernetes version is 1.21
 * Bump https://github.com/bitpoke/build to 0.8.3
 * Allow the operator to be namespace scoped (`SCOPED_NAMESPACE` environment variable)
 * Set resources for `init-wp` initContainer

## [0.12.1] - 2021-12-22
### Changed
 * Bump https://github.com/bitpoke/build to 0.7.1
### Fixed
 * Fix the app version in the published Helm charts

## [0.12.0] - 2021-12-22
### Added
### Changed
 * Minimum required Kubernetes version is 1.19
 * Use `networking.k8s.io/v1` for `Ingress` resources
 * Run WordPress Operator as non-root user
 * Bump https://github.com/bitpoke/build to 0.7.0
### Removed
### Fixed

## [0.11.1] - 2021-11-22
### Changed
 * Change the default image to `docker.io/bitpoke/wordpress-runtime:5.8.2`

## [0.11.0] - 2021-11-15
### Changed
 * Use [Bitpoke Build](https://github.com/bitpoke/build) for building the
   project
### Removed
 * Drop support for Helm v2
### Fixed
