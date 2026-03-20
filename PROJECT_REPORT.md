# DD Bank Project Report

## Executive Summary

DD Bank is a mini core banking system built as a polyglot backend project using Kotlin, Go, Rust, and PostgreSQL.

The main business flow is:

`Client -> Go transaction-service -> Rust fraud-service -> Kotlin ledger-service -> PostgreSQL`

The project goal was to build a transfer flow that is:

- atomic
- double-entry consistent
- idempotent
- fraud-aware
- observable across services

This report documents the project from Plan 1 through Plan 30, including implementation decisions, verification steps, and final outcomes.

## Business Objective

The system simulates a simplified banking transfer pipeline where:

- the Go service acts as the external API gateway and orchestrator
- the Rust service acts as a fraud-decision engine
- the Kotlin service acts as the ledger and source of truth
- PostgreSQL stores accounts, transactions, and journal entries

The most important rule in this project is:

If a transfer is not correct, the system must stop and reject it rather than create partial or inconsistent state.

## Technology Stack

- Kotlin + Spring Boot: ledger service
- Go: transaction orchestration service
- Rust + Axum: fraud service
- PostgreSQL: persistence layer
- PowerShell + Taskfile: local automation without Docker

## Architecture Overview

### Services

- `transaction-service` exposes `POST /transfer` and coordinates the full request lifecycle.
- `fraud-service` exposes `POST /fraud/check` and returns `approved` or `rejected`.
- `ledger-service` exposes `POST /ledger/transfer`, `GET /ledger/transactions`, and `GET /health`.

### Data Flow

1. The client sends a transfer request to the Go service.
2. The Go service validates and forwards the request to the Rust fraud service.
3. If the request is rejected by fraud rules, the process stops.
4. If the request is approved, the Go service forwards it to the Kotlin ledger service.
5. The ledger writes one transaction row and two journal rows inside a single database transaction.
6. A completion event is emitted after commit.
7. A standardized response is returned to the client with a correlation ID.

## Plan-by-Plan Delivery

### Plan 1 - Bootstrap

- Set up the Spring Boot Kotlin ledger project.
- Configured PostgreSQL connectivity for the ledger.
- Verified the application starts successfully.
- Added `GET /health` for service health checks.

Result:
The ledger service boots successfully and exposes a health endpoint for readiness checks.

### Plan 2 - Domain Model

- Created `Account` entity without balance storage.
- Created `LedgerTransaction` entity with a unique business reference.
- Created `JournalEntry` entity for `DEBIT` and `CREDIT`.
- Added JPA repositories for all core domain objects.

Result:
The core banking data model supports double-entry bookkeeping without storing a mutable account balance field.

### Plan 3 - Schema Validation

- Started the application and allowed tables to be created automatically.
- Verified tables in PostgreSQL.
- Confirmed important column types such as `uuid`, `numeric`, and `timestamp with time zone`.
- Confirmed transaction `reference` is unique in the database.

Result:
The database schema matches the intended domain design and prevents duplicate transaction references.

### Plan 4 - Transfer Engine Core

- Implemented ledger transfer logic in the Kotlin service.
- Created a transaction record for each successful transfer.
- Inserted one debit journal entry.
- Inserted one credit journal entry.
- Wrapped the write flow in database transactions.

Result:
A transfer is stored as one transaction plus exactly two journal entries in one atomic unit.

### Plan 5 - Transfer API

- Added `POST /ledger/transfer`.
- Created request and response DTOs.
- Connected controller, service, and persistence layers.
- Verified the endpoint with end-to-end requests.

Result:
The ledger service is callable through an explicit write API.

### Plan 6 - Validation

- Validated that `amount > 0`.
- Validated source and destination accounts exist.
- Validated `from_account != to_account`.
- Rejected invalid requests with meaningful errors.

Result:
Invalid transfers are blocked before any ledger write happens.

### Plan 7 - Consistency Guarantee

- Calculated total debit for each transfer.
- Calculated total credit for each transfer.
- Enforced `debit == credit`.
- Threw an error if a mismatch occurs.

Result:
The double-entry rule is enforced by the ledger before commit.

### Plan 8 - Idempotency Basic

- Checked transactions by reference before insert.
- Prevented duplicate writes for the same business reference.
- Returned an existing result safely on retry.
- Verified duplicate request behavior.

Result:
Retrying the same transfer is safe and does not create duplicate financial records.

### Plan 9 - Test Core Flow

- Added tests for a valid transfer.
- Added tests for duplicate transfer handling.
- Added tests for invalid amount validation.
- Verified journal entry creation in tests.

Result:
The main ledger write path is covered by automated tests.

### Plan 10 - Go Service Setup

- Created the Go transaction service structure.
- Added an HTTP server.
- Added `POST /transfer`.
- Added `GET /health`.

Result:
The orchestration layer is exposed through a dedicated public API.

### Plan 11 - Go to Ledger Integration

- Implemented the Go HTTP client for the ledger.
- Forwarded the transfer payload from Go to Kotlin.
- Returned the ledger response through the Go service.
- Preserved correlation IDs between services.

Result:
The client-facing service can delegate approved transfers to the ledger.

### Plan 12 - Rust Service Setup

- Created the Rust fraud service project.
- Added an HTTP server.
- Added `POST /fraud/check`.
- Added `GET /health`.

Result:
The fraud engine is independently callable and operational.

### Plan 13 - Fraud Logic

- Implemented an amount-threshold fraud rule.
- Returned `approved` or `rejected`.
- Handled JSON payload parsing and validation.
- Added automated tests for valid and invalid scenarios.

Result:
The fraud service blocks high-value transfers before ledger writes occur.

### Plan 14 - Full Integration

- Connected Go to Rust.
- Consumed the fraud decision in Go.
- Connected Go to Kotlin after approval.
- Verified the full request chain.

Result:
The intended full system flow now works end to end.

### Plan 15 - Flow Validation

- Verified small transfer success.
- Verified large transfer rejection.
- Verified database state after requests.
- Ensured rejected transfers do not create partial writes.

Result:
The full chain behaves correctly for both success and rejection scenarios.

### Plan 16 - Idempotency Hardening

- Enforced database uniqueness for transaction references.
- Handled duplicate insert races safely.
- Refactored the ledger posting flow to recover cleanly from unique-constraint collisions.
- Tested repeated requests and concurrent duplicate requests.

Result:
Idempotency is protected both at application level and database level.

### Plan 17 - Error Handling

- Added structured application errors.
- Handled fraud service failures.
- Handled ledger service failures.
- Returned meaningful HTTP errors without crashing the process.

Result:
The system fails safely and returns useful error responses.

### Plan 18 - Timeout and Retry

- Added bounded HTTP timeout configuration in Go.
- Added retry logic for retryable downstream failures.
- Prevented infinite retry loops.
- Added tests covering timeout and retry behavior.

Result:
The orchestration layer is resilient to temporary downstream instability.

### Plan 19 - Logging

- Added correlation ID propagation.
- Logged errors with context.
- Preserved request traceability between services.
- Included correlation IDs in responses and headers.

Result:
Cross-service debugging is much easier and request paths are observable.

### Plan 20 - Consistency Thinking

- Reviewed the end-to-end failure model.
- Ensured the ledger remains atomic.
- Added rollback verification through tests.
- Confirmed no partial journal writes remain after failure.

Result:
Consistency is treated as a system-wide requirement, not just a database detail.

### Plan 21 - Event Emit

- Added `TransactionCompletedEvent`.
- Emitted the event after a successful commit.
- Defined a clear event structure.
- Added logging-oriented side effect support.

Result:
The system can trigger post-commit side effects without mixing them into the transaction write path.

### Plan 22 - Event Handling

- Added an event handler abstraction.
- Added a side-effect interface and implementation.
- Simulated asynchronous post-success behavior through event handling.
- Verified event emission in tests.

Result:
The ledger now supports decoupled post-transfer processing.

### Plan 23 - Read Model

- Added `GET /ledger/transactions`.
- Queried persisted transactions safely.
- Returned a list response for presentation and inspection.
- Added support for `limit` to keep reads bounded.

Result:
The system includes a usable read API for transfer history.

### Plan 24 - CQRS Basic

- Kept write logic in the command service.
- Kept transaction-history retrieval in the query service.
- Reduced mixing between read and write concerns.
- Organized the structure more clearly for future growth.

Result:
The ledger codebase now separates command and query responsibilities in a simple but effective way.

### Plan 25 - Refactor

- Improved naming around transfer execution and mapping.
- Extracted focused classes such as posting executor and response mapper.
- Reduced duplication in transfer handling.
- Cleaned the service layer structure.

Result:
The codebase is easier to read, reason about, and extend.

### Plan 26 - Validation Hardening

- Hardened JSON handling and malformed request handling in Go.
- Strengthened ledger validation for invalid business input.
- Covered edge cases such as duplicate references and invalid payload structure.
- Added tests for malformed or empty downstream responses.

Result:
The services reject malformed input more predictably and safely.

### Plan 27 - Concurrency

- Added concurrent duplicate-request tests inside the Kotlin integration suite.
- Added concurrent distinct-transfer validation to confirm balanced journals under parallel load.
- Added end-to-end duplicate race validation through a PowerShell harness.
- Verified that one business reference results in only one persisted transaction.

Result:
The system holds its financial consistency guarantees during concurrent duplicate traffic.

### Plan 28 - Database Constraint

- Confirmed unique constraint behavior for transaction references.
- Confirmed foreign key failures for invalid journal entries.
- Added schema verification script assertions for required tables, indexes, and foreign keys.
- Verified integrity rules directly against PostgreSQL.

Result:
Database-level protection reinforces application-level safety.

### Plan 29 - Performance

- Added indexing on frequently accessed transaction and journal columns.
- Added a lightweight local load harness without Docker.
- Reduced unnecessary read loading through bounded queries.
- Collected latency and throughput metrics from local runs.

Result:
The project now has a repeatable local performance check and basic query optimization.

### Plan 30 - Final Polish

- Standardized API response structure across services.
- Finalized error and health response format.
- Wrote project documentation and operating instructions.
- Added Taskfile-based local automation for run, test, smoke, concurrency, load, and DB verification.

Result:
The project is presentation-ready, easier to demo, and easier to validate on a local machine without Docker.

## Key Technical Decisions

### 1. Double-Entry Instead of Stored Balance

Accounts do not store a mutable balance column. Instead, each transfer creates journal entries. This design reduces the risk of balance drift and aligns better with accounting principles.

### 2. Idempotency as a First-Class Requirement

Transaction references are unique at both application and database levels. This protects the system from client retries and duplicate delivery.

### 3. Fraud Before Ledger

Fraud checking happens before the ledger write. This ensures rejected transfers never create financial state.

### 4. Post-Commit Event Pattern

Events are emitted only after a successful commit, reducing the chance of side effects running for failed transactions.

### 5. No-Docker Local Automation

Because Docker was intentionally not used, local automation was implemented with Taskfile and PowerShell scripts to keep the workflow reproducible and presentation-friendly.

## Verification and Test Evidence

### Automated Verification Commands

The following commands were successfully executed:

- `task test`
- `task db:verify`
- `task smoke`
- `task concurrency`
- `task load`

### Kotlin Ledger Coverage

The Kotlin integration suite verifies:

- healthy startup and health endpoint
- valid transfers
- duplicate transfers
- invalid amount rejection
- bounded transaction reads
- concurrent duplicate idempotency
- concurrent distinct transfer consistency
- rollback on forced persistence failure
- unique-constraint enforcement
- foreign-key enforcement
- event emission after success

### Go Service Coverage

The Go test suite verifies:

- retry behavior on temporary downstream failures
- timeout behavior
- malformed request rejection
- health responses with correlation IDs
- downstream response validation

### Rust Service Coverage

The Rust test suite verifies:

- small amount approval
- large amount rejection
- invalid payload rejection

## Runtime Validation Results

### Smoke Test

The smoke test verifies:

- ledger, fraud, and transaction services become healthy
- a small transfer succeeds
- a large transfer is rejected by the fraud service
- a repeated request with the same reference returns `duplicate=true`
- only the expected successful transfers are persisted

### Concurrency Validation

The duplicate-request concurrency test verified:

- `10` concurrent requests with the same reference
- exactly `1` non-duplicate writer
- exactly `9` duplicate responses
- exactly `1` persisted ledger transaction
- exactly `2` persisted journal entries

### Load Validation

One verified local load run produced:

- `40` requests
- concurrency `8`
- `40/40` successful responses
- `p95 = 497 ms`
- `p99 = 564 ms`
- throughput `1.31 requests/sec`

These numbers reflect a local PowerShell-based harness, not a dedicated benchmarking tool.

## Final Success Criteria Review

### Transfer is atomic

Achieved.
The ledger writes the transaction record and both journal entries in one transactional boundary and rollback behavior is covered by tests.

### Debit equals credit always

Achieved.
The ledger enforces journal balance before commit and concurrent tests confirm balanced postings.

### Duplicate request is safe

Achieved.
Application checks, database uniqueness, and duplicate-race tests all confirm safe idempotent behavior.

### Fraud decision works

Achieved.
The Rust service rejects transfers above the configured threshold and the Go orchestrator respects that result.

### Services are integrated cleanly

Achieved.
The full chain from client to database works through the intended service boundaries.

### System is observable

Achieved.
Correlation IDs and structured response formats allow requests to be traced across all services.

## Project Outcome

DD Bank successfully meets the core goals of the assignment:

- a polyglot microservice flow
- a correct ledger implementation
- safe retry and duplicate handling
- fraud-aware orchestration
- transaction integrity under concurrency
- local automation for validation and demo

The project is now suitable for:

- backend engineering presentations
- portfolio demonstrations
- system-design walkthroughs
- discussions about consistency, idempotency, and service orchestration

## Suggested Next Evolution

If this project is extended beyond the current scope, the next strongest improvements would be:

- add versioned database migrations with Flyway
- adopt stronger tracing with OpenTelemetry
- replace the lightweight load harness with `k6` or `vegeta`
- add queue-backed async event processing
- add authentication and authorization
- introduce deployment manifests for cloud environments

## Closing Statement

From Plan 1 to Plan 30, the project evolved from a simple service bootstrap into a complete mini core banking flow with fraud checking, transactional posting, idempotency, read models, observability, automation, and verification.

The final result is not only functional, but also demonstrable, testable, and presentation-ready.
