# Database Migrations

This directory contains SQL migration files for the hr-api database.

## Migration Tool

We use [golang-migrate/migrate](https://github.com/golang-migrate/migrate) for database migrations.

## File Naming Convention

Migrations use sequential numbering: `NNNNNN_name_of_change.{up|down}.sql`

- **up**: Applies the migration (CREATE TABLE, ALTER TABLE, etc.)
- **down**: Rolls back the migration (DROP TABLE, ALTER TABLE rollback, etc.)

Example: `000002_add_index_to_users_table.up.sql`

## Running Migrations

### Apply all pending migrations:
```bash
make migrate-up
```

### Rollback last migration:
```bash
make migrate-down
```

### Create new migration:
```bash
make migrate-create add_email_verification_table
```

### Check current migration version:
```bash
make migrate-version
```

### Force specific version (use with caution):
```bash
make migrate-force 000003
```

### Drop all tables (DESTRUCTIVE):
```bash
make migrate-drop
```

## Environment Variables

- `DATABASE_URL`: PostgreSQL connection string (required)
  Format: `postgres://user:password@host:port/database?sslmode=disable`

## Migrations

| ID | Name | Description |
|----|-------|-------------|
| 000001 | init_schema | Creates initial users and employees tables |

## Development Notes

- Always create both `.up.sql` and `.down.sql` files
- Test `down` migrations to ensure rollback works
- Keep migrations reversible when possible
- Use transactions in complex migrations (multiple statements)

## Current Version

Check with `make migrate-version`
