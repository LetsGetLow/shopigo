# Shopigo

**Work in progress.**

Shopigo is a Go-based e-commerce/catalog service currently focused on the catalog domain: categories, products, variants, and audit-aware persistence.

## What exists today

- Domain models for catalog entities
- Shared audit metadata for created/updated/deleted state
- PostgreSQL-backed category persistence
- Integration tests for category CRUD and soft delete behavior
- Initial catalog use-case notes

## Current focus

- Category hierarchy and lifecycle
- Product and variant domain modeling
- Persistence and repository behavior
- Catalog audit fields and soft deletion

## Tech stack

- Go 1.26
- PostgreSQL
- `sqlx`
- `lib/pq`
- `google/uuid`
- `go-money`

## Local development

### Start PostgreSQL

```bash
docker compose up -d
```

### Configure environment

Set the usual Postgres variables in `.env` or your shell:

```bash
POSTGRES_USER=shopigo
POSTGRES_PASSWORD=password
POSTGRES_DB=shopigo
POSTGRES_PORT=15432
```

### Run tests

```bash
go test ./...
```

## Catalog notes

The catalog is being built around a few core workflows:

- Create and update categories
- Nest categories under parent categories
- Create products and assign them to categories
- Create variants for products
- Track audit fields across writes and soft deletes

## Status

This repository is actively under construction. Expect APIs, data models, and persistence details to evolve.