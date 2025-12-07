---
title: Bulk Import Hardening Plan
date: 2025-11-12
author: Platform Team
---

# Bulk Import Hardening Plan

This document captures the outstanding work required to make the bulk-import pipeline resilient in production. Tasks are grouped by theme and include concrete sub-tasks.

## 1. Error Handling & Recovery

- Add retry/backoff around transient dependencies:
  - Wrap MinIO downloads and CSV parsing in exponential backoff with jitter.
  - Retry DB writes (category/supplier/location creation) on serialization or connection failures.
- Introduce a dead-letter queue (DLQ):
  - Configure a `bulk.import.dlq` and write failed messages there once retries are exhausted.
  - Provide tooling to inspect/replay DLQ items.
- Expand job failure logging:
  - Persist structured error context (bucket/object names, row counts, offending row) in `jobs.last_error`.
  - Emit log entries with correlation/job IDs for every failure path.

## 2. Per-Row Validation & Partial Success

- Capture row-level validation results:
  - Extend `BulkImportResult` to include row number, original data, and validation errors.
  - Store this structure in a dedicated table or JSON column for later download.
- Allow partial imports:
  - Separate validation from insertion by persisting valid rows and letting the user confirm per-row or per-batch.
  - Optionally continue inserting valid records even if some rows fail, surfacing the failures back to the user.
- Expose results via API/UI:
  - Add endpoints to fetch row-level errors and warnings.
  - Update the frontend to show detailed validation outcomes and allow CSV export of failed rows.

## 3. Concurrency & Throughput

- Chunk processing to avoid huge transactions:
  - Split validated products into smaller batches (e.g., 500 rows) and commit incrementally.
  - Ensure entity caches (categories, suppliers, locations) remain consistent across chunks.
- Configurable worker pools:
  - Allow separate worker counts for validation vs. insertion (e.g., `BULK_IMPORT_VALIDATE_WORKERS`, `BULK_IMPORT_INSERT_WORKERS`).
  - Support horizontal scaling by shard key (job ID or user ID).
- Back-pressure and queue visibility:
  - Monitor RabbitMQ queue length and pause uploads if consumers fall behind.
  - Surface estimated completion times to the UI.

## 4. Monitoring & Observability

- Metrics:
  - Publish Prometheus counters/gauges for job counts, durations, failures, DLQ depth.
  - Track per-step timings (validation, finalize, confirmation) for SLO dashboards.
- Structured logging & tracing:
  - Add log fields (`job_id`, `user_id`, `bucket`, `object`) to every log statement in the consumer.
  - Instrument the pipeline with OpenTelemetry spans so traces connect upload → validation → insert.
- Alerting:
  - Create alerts for DLQ growth, repeated failures per job, or long-running jobs.

## 5. Hardening & Administrative Tools

- Retry policy:
  - Implement per-job retry counters with escalating delays (e.g., 3 retries with exponential backoff).
  - Auto-fail jobs after max attempts and notify users/admins.
- Admin recovery:
  - Provide CLI/HTTP endpoints to requeue failed jobs, purge DLQ messages, or mark jobs as resolved.
  - Build a simple admin UI/table showing job history, error reasons, and recovery actions.
- Data consistency checks:
  - Add periodic audits that verify products created via bulk import have valid foreign keys and expected counts.

## 6. Documentation & Runbooks

- Update developer docs with the new architecture (workers, DLQ, metrics).
- Write an SRE runbook that covers:
  - How to replay DLQ messages.
  - How to recover from partial imports.
  - How to interpret job metrics and logs.

---

**Timeline & Ownership**

- Phase 1 (Error handling, DLQ, job logging): Platform team, 1 sprint.
- Phase 2 (Per-row validation + partial success): App team, 1–2 sprints.
- Phase 3 (Metrics, tracing, admin tooling): Shared effort, 1 sprint.

Progress should be tracked in the engineering board with separate epics per section.
