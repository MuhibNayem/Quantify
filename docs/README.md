# Quantify Documentation

This directory contains specification, operations, and reference documentation
for the Quantify backend. Start with the core docs below; the API contract is
in [`backend/api/openapi.yaml`](../backend/api/openapi.yaml).

## Core specification

| Document | What it covers |
| :--- | :--- |
| [architecture.md](./architecture.md) | Layered design (Handler → Service → Repository), the DI container, ADRs (router split, golang-migrate, koanf config), deployment topology, and the request lifecycle. |
| [business_logic.md](./business_logic.md) | RBAC model and permission catalog, checkout flow (FEFO, locking, promotions, tax, loyalty), return flow, bulk import lifecycle, and settings-driven loyalty tiers. |
| [security.md](./security.md) | STRIDE-lite threat model, controls per threat, secret handling, and the key rotation runbook. |
| [runbook.md](./runbook.md) | Health checks, common incidents (Postgres/Redis/RabbitMQ/MinIO down, job backlog), deploy and rollback procedures. |

## Reference & guides

| Document | What it covers |
| :--- | :--- |
| [access_control.md](./access_control.md) | Mapping of permissions to sidebar items and UI routes, plus recommended role configurations. |
| [reports_user_guide.md](./reports_user_guide.md) | End-user guide to the reporting suite. |
| [report-walkthrough.md](./report-walkthrough.md) | Technical walkthrough of report implementation. |
| [shop_owner_manual.md](./shop_owner_manual.md) | Operator-facing manual for running a store on Quantify. |
| [bulk_import_hardening_plan.md](./bulk_import_hardening_plan.md) | Design notes behind the confirmation-gated bulk import. |
| [backend_analysis.md](./backend_analysis.md) | Historical backend analysis notes. |
| [log-shipping.md](./log-shipping.md) | Log shipping / aggregation notes. |

## Where else to look

- **API contract:** `backend/api/openapi.yaml` (Swagger UI at `/swagger/` when running).
- **Change history:** `backend/CHANGELOG.md` — the authoritative record of behavior changes per phase.
- **Config reference:** the environment variable table in the root [`README.md`](../README.md), sourced from `backend/internal/config/config.go`.
