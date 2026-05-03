# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
### Changed
### Removed
### Fixed

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
