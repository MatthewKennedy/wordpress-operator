# wordpress-operator improvements backlog

Captured 2026-05-03 after the v2.0.1 release. Priorities are guesses — confirm before working.

## Real bugs

- **B1. Reconcile churn from `FieldRef.APIVersion` syncer thrash.** Every reconcile logs an error: `Spec.Template.Spec.InitContainers.slice[0].Env.slice[0].ValueFrom.FieldRef.APIVersion: v1 != ""`. The controller-util syncer detects a phantom diff (apiserver defaults `APIVersion` to `v1` server-side, but the desired object has `""`), tries to patch it, and fails with optimistic-concurrency every time. Pre-existing, not from v2.0 work. Likely fixed in `presslabs/controller-util` v0.18 — verify by reading release notes before bumping.
- **B2. `KubeAPIWarningLogger` log spam.** PVC syncers (`pkg/controller/wordpress/internal/sync/code_pvc.go`, `media_pvc.go`) write back `metadata.creationTimestamp: null` into `spec.code.metadata` / `spec.media.metadata`, which the CRD doesn't define. Cosmetic, but constant.
- **B3. wp-cron-controller error spam.** `pkg/controller/wp-cron/wpcron_controller.go:115` logs at `error` level with a full stack trace every time the WordPress pod is unreachable. With many sites this is a log-volume problem. Should be `warn`, or quieted on transient connection errors.

## Quick wins (Q1–Q7)

All shipped in v2.0.2 (2026-05-03). See CHANGELOG for details.

## Medium

- **M1. Remove `cleanupCronJob`** (`pkg/controller/wordpress/wordpress_controller.go:131`/`176`). Migration code from when wp-cron was a Kubernetes `CronJob`. With v2.0.x being a clean break, this is likely dead. Could be wrong for users hopping multiple major versions. Verify before deleting.
- **M2. Remove deprecated `spec.domains`** (`pkg/apis/wordpress/v1alpha1/wordpress_types.go:81`). Already migrated to `spec.routes` via `maybeMigrate` in the controller. Worth removing for v3.

## Bigger upgrades (warrant their own release each)

- **U1. controller-runtime v0.18.5 → v0.24.0** (six minor versions). **Requires Go 1.26** for the latest. Bundles with `k8s.io/*` v0.30.6 → v0.36.0. Real API churn — similar in shape to the v0.9 → v0.18 jump in v2.0.0. Likely a couple of days.
- **U2. golangci-lint v1.56.2 → v2.x.** Breaking config schema change. Worth it for newer rules. Half day.
- **U3. presslabs/controller-util v0.15.0 → v0.18.0.** May fix B1. Read release notes first.

## Reference

| Component | Current | Latest |
|---|---|---|
| controller-runtime | v0.18.5 | v0.24.0 |
| k8s.io/{api,apimachinery,client-go} | v0.30.6 | v0.36.0 |
| presslabs/controller-util | v0.15.0 | v0.18.0 |
| ginkgo/v2 | v2.22.2 | v2.28.3 |
| gomega | v1.36.2 | v1.40.0 |
| Go | 1.22 | 1.26 (required by cr v0.24) |
| golangci-lint | v1.56.2 | v2.x |
| controller-gen | v0.15.0 | v0.16+ |
