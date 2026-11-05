# Go backend — Modulith

- Run Go commands inside `BE/`. Module: `sannha/be`, Go 1.25+, PostgreSQL 17+ via pgx; no CGO dependency.
- One executable, one PostgreSQL database. `internal/app` is the composition root. `DATABASE_URL` is required; never print it or commit credentials.
- `identity` owns users/sessions, `venue` owns venues/courts/favorites, `booking` owns reservations/slot locks.
- Booking also owns deposit payments and VietQR callback tokens/events. Keep event persistence, deposit_paid and booking confirmation atomic in PostgreSQL. Never treat generated QR, browser polling or a client assertion as evidence of payment. Use provider transaction IDs for idempotency; cancelled slots must never be revived by late callbacks.
- Booking owns vouchers and redemptions too. Quote never reserves quota. Reserve quota and insert booking in one transaction; cancellation must lock voucher before booking and only release unpaid uses. Never reprice existing bookings or QR orders when vouchers change.
- Never query another module's tables. The booking module uses its `Catalog` interface and exported venue contracts.
- Cross-module foreign keys are intentionally absent. Validate references through module APIs. Keep intra-module foreign keys.
- Add forward-only, named schema migrations; never edit an already-applied migration for an upgrade. Active migrations are `schema.postgres.sql` (`*_pg_v1`); `schema.sqlite.sql` is archived, never executed. Do not delete old SQLite data.
- No repository/service interface scaffolding unless a real caller needs it. Keep SQL inside its owning module.
- Amounts are integer VND. Prices and deposits are captured at reservation creation.
- Authentication/authorization must be server-side. No public admin registration, hard-coded credentials, plaintext passwords/tokens, or PII in logs.
- OAuth Google/GitHub lives in identity. Keep state single-use, browser-bound and expiring, use PKCE, never auto-link accounts by email, and never expose provider tokens/secrets to FE. New OAuth users have role user only.
- Run `gofmt`, `go vet ./...`, `go test -race ./...`, `go build ./...`.
- Integration tests require `TEST_DATABASE_URL` for a dedicated test database. Each test creates/drops only its random schema. Use `make test` to fail when the URL is missing; without it, integration tests skip explicitly.
