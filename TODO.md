# wordpress-operator improvements backlog

Captured 2026-05-03 after the v2.0.1 release. Priorities are guesses — confirm before working.

## Real bugs (B1–B3)

All shipped in v2.1.0 (2026-05-03). See CHANGELOG for details.

## Quick wins (Q1–Q7)

All shipped in v2.0.2 (2026-05-03). See CHANGELOG for details.

## Medium (M1–M2)

All shipped in v3.0.0 (2026-05-03). See CHANGELOG for details.

## Bigger upgrades

These are **coupled** — U1 and U3 share the same dep tree. Plan together.

- **U1+U3. Bump the Kubernetes-API ecosystem.** Two natural stopping points:
  - **Modest (the U3 floor).** Bump to `presslabs/controller-util v0.18.0`. Forces:
    - Go 1.22 → **1.24**
    - `k8s.io/*` v0.30.6 → **v0.33.2**
    - `sigs.k8s.io/controller-runtime` v0.18.5 → **v0.21.0**
    - controller-util v0.15.0 → **v0.18.0**
    - Likely smaller API churn than the v2.0.0 jump (which was cr v0.9 → v0.18). Half-day to a day. Ship as v3.1.0.
  - **Latest (full U1).** Bump to controller-runtime v0.24. Adds:
    - Go 1.24 → **1.26** (bleeding edge — Go 1.26 is recent)
    - `k8s.io/*` v0.33.2 → **v0.36.0**
    - controller-runtime v0.21.0 → **v0.24.0**
    - Another half-day of API churn on top of the modest path. Ship as v4.0.0 (or v3.2.0 if no breaking ops surface).

  B1 (FieldRef thrash) is already fixed in v2.1.0; we shouldn't expect controller-util v0.18 to bring its own fixes worth the cost.

- **U2. golangci-lint v1.56.2 → v2.x.** Breaking config schema change (`run.skip-dirs` removed, `output.format` reshaped, several linters renamed). Worth it for newer rules. Independent of U1/U3 — can ship as a small standalone change. Half day.

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
