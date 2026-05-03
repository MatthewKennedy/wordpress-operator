# wordpress-operator improvements backlog

Captured 2026-05-03 after the v2.0.1 release. Priorities are guesses — confirm before working.

## Real bugs (B1–B3)

All shipped in v2.1.0 (2026-05-03). See CHANGELOG for details.

## Quick wins (Q1–Q7)

All shipped in v2.0.2 (2026-05-03). See CHANGELOG for details.

## Medium (M1–M2)

All shipped in v3.0.0 (2026-05-03). See CHANGELOG for details.

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
