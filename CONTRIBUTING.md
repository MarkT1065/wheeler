# Contributing to Wheeler

Wheeler is a web-based portfolio tracker for options trading (particularly the wheel strategy). It's built with Go, SQLite, and vanilla web tech.

## Before You Start

**Tests first.** PRs without tests won't be merged.

Wheeler uses Test Driven Development. Write the test, watch it fail, make it pass. Every PR should be self-testing and `make test` should always pass.

## Quick Commands

```bash
make test      # Run all tests (this should always work)
make build     # Build the application
make run       # Build and run
make clean     # Clean build artifacts
make help      # Show all commands
```

## Database Migrations

Wheeler uses file-based SQL migrations tracked in `internal/database/migrations/`.

**Read the migration docs first**: See `internal/database/migrations/README.md`

Key rules:
- Migrations are append-only (never edit existing migrations)
- Use idempotent SQL (`CREATE TABLE IF NOT EXISTS`, `INSERT OR IGNORE`)
- Timestamp-based naming: `YYYYMMDDHHMMSS_description.sql`
- Test migrations by running `make test` against a fresh database

## Pull Request Checklist

Before opening a PR:

- [ ] `make test` passes completely
- [ ] New code has test coverage
- [ ] Database changes include a migration file
- [ ] No hardcoded secrets or API keys
- [ ] Follows existing code patterns (check similar files first)

## Development Workflow

1. **Read `CLAUDE.md`** - Contains project architecture and conventions
2. **Run the app**: `go run main.go` → http://localhost:8080
3. **Load test data**: Help → Tutorial → "Generate Test Data"
4. **Make your changes**
5. **Write tests first**
6. **Run `make test`** until it passes
7. **Open PR**

## Code Style

- No comments unless you're asked to add them
- Follow existing patterns (check similar files)
- Use semantic naming (readability over brevity)
- Keep functions small and testable

## Questions?

Check these files first:
- `CLAUDE.md` - Architecture and development guidance
- `model.md` - Database schema and data model
- `README.md` - User-facing documentation
- `internal/database/migrations/README.md` - Migration system

---

Wheeler is a part-time project. Keep PRs focused, tested, and easy to review.
