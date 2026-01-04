# Migration Commands Reference

This document contains useful migration commands for database management.

## Prerequisites

Install [golang-migrate](https://github.com/golang-migrate/migrate):

```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Or with Go
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Verify installation
migrate -version
```

## Create Migration File

```bash
migrate create -ext sql -dir database/migrations -seq create_users_table
```

This command will create two files:
- `{timestamp}_create_users_table.up.sql` - For applying the migration
- `{timestamp}_create_users_table.down.sql` - For rolling back

**Note:** The `-seq` flag creates sequential numbering (001, 002, etc.) instead of timestamps.

## Run Migration

### Format

```bash
migrate -path [path] -database [connection_string] [action] [N]
```

### Examples

**Run all pending migrations:**

```bash
migrate -path database/migrations \
  -database "mysql://root:password@tcp(localhost:3306)/tmn_backend" \
  up
```

**Run specific number of migrations:**

```bash
migrate -path database/migrations \
  -database "mysql://root:password@tcp(localhost:3306)/tmn_backend" \
  up 1
```

**Rollback last migration:**

```bash
migrate -path database/migrations \
  -database "mysql://root:password@tcp(localhost:3306)/tmn_backend" \
  down 1
```

**Rollback all migrations:**

```bash
migrate -path database/migrations \
  -database "mysql://root:password@tcp(localhost:3306)/tmn_backend" \
  down
```

**Force to a specific version (use with caution):**

```bash
migrate -path database/migrations \
  -database "mysql://root:password@tcp(localhost:3306)/tmn_backend" \
  force 1
```

**Check current version:**

```bash
migrate -path database/migrations \
  -database "mysql://root:password@tcp(localhost:3306)/tmn_backend" \
  version
```

## Run Migration with Docker

If you prefer to use Docker instead of installing migrate on your host:

### Single line command:

```bash
docker run -v $(pwd)/database/migrations:/migrations --network host migrate/migrate -path=/migrations -database "mysql://root:password@tcp(127.0.0.1:3306)/tmn_backend" up
```

### Multi-line command (better readability):

```bash
docker run -v $(pwd)/database/migrations:/migrations --network host migrate/migrate \
  -path=/migrations \
  -database "mysql://root:password@tcp(127.0.0.1:3306)/tmn_backend" \
  up
```

**⚠️ Note:** When splitting commands across multiple lines in zsh/bash, you must use a backslash (`\`) at the end of each line to indicate line continuation. Without it, each line will be treated as a separate command.

### Docker commands for different actions:

**Rollback with Docker:**

```bash
docker run -v $(pwd)/database/migrations:/migrations --network host migrate/migrate \
  -path=/migrations \
  -database "mysql://root:password@tcp(127.0.0.1:3306)/tmn_backend" \
  down 1
```

**Check version with Docker:**

```bash
docker run -v $(pwd)/database/migrations:/migrations --network host migrate/migrate \
  -path=/migrations \
  -database "mysql://root:password@tcp(127.0.0.1:3306)/tmn_backend" \
  version
```

## Connection String Format

```
mysql://username:password@tcp(host:port)/database
```

**Examples:**

```bash
# Local development
mysql://root:password@tcp(localhost:3306)/tmn_backend

# Remote server
mysql://user:pass@tcp(192.168.1.100:3306)/tmn_backend

# With special characters in password (URL encode)
mysql://root:p%40ssw0rd@tcp(localhost:3306)/tmn_backend
```

## Common Migration Patterns

### Creating a table:

**up.sql:**
```sql
CREATE TABLE IF NOT EXISTS posts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    user_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

**down.sql:**
```sql
DROP TABLE IF EXISTS posts;
```

### Adding a column:

**up.sql:**
```sql
ALTER TABLE users ADD COLUMN phone VARCHAR(20) AFTER email;
```

**down.sql:**
```sql
ALTER TABLE users DROP COLUMN phone;
```

### Creating an index:

**up.sql:**
```sql
CREATE INDEX idx_users_email ON users(email);
```

**down.sql:**
```sql
DROP INDEX idx_users_email ON users;
```

## Troubleshooting

### Error: "Dirty database version"

This happens when a migration fails halfway:

```bash
# Check current version and dirty state
migrate -path database/migrations \
  -database "mysql://root:password@tcp(localhost:3306)/tmn_backend" \
  version

# Force to a specific clean version
migrate -path database/migrations \
  -database "mysql://root:password@tcp(localhost:3306)/tmn_backend" \
  force 1
```

### Error: "no change"

This means all migrations are already applied. To force re-run:

```bash
# Rollback first
migrate -path database/migrations -database "..." down 1

# Then run again
migrate -path database/migrations -database "..." up 1
```

### Error: "database connection refused"

- Check MySQL is running: `mysql -u root -p`
- Verify credentials in connection string
- Ensure database exists: `CREATE DATABASE tmn_backend;`

## Best Practices

1. **Always write down migrations** - Every up should have a corresponding down
2. **Test migrations** - Test both up and down on a copy of production data
3. **Keep migrations small** - One logical change per migration
4. **Never modify applied migrations** - Create new migrations for changes
5. **Use transactions** - MySQL InnoDB supports transactional DDL
6. **Version control** - Commit migration files to git
7. **Document complex migrations** - Add comments explaining why

## Build Commands

Build the application for production:

```bash
# Local build
go build -o tmn-backend

# Linux build (from macOS/Windows)
env GOOS=linux GOARCH=amd64 go build -o tmn-backend

# With optimization
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o tmn-backend
```

## Useful MySQL Commands

```bash
# Show all tables
mysql -u root -p tmn_backend -e "SHOW TABLES;"

# Show table structure
mysql -u root -p tmn_backend -e "DESCRIBE users;"

# Show all migrations applied
mysql -u root -p tmn_backend -e "SELECT * FROM schema_migrations;"

# Drop database (careful!)
mysql -u root -p -e "DROP DATABASE tmn_backend;"

# Recreate database
mysql -u root -p -e "CREATE DATABASE tmn_backend;"
```

## Quick Reference

| Command | Description |
|---------|-------------|
| `migrate create -ext sql -dir path -seq name` | Create new migration |
| `migrate -path path -database "..." up` | Run all pending |
| `migrate -path path -database "..." up 1` | Run one migration |
| `migrate -path path -database "..." down 1` | Rollback one |
| `migrate -path path -database "..." version` | Check version |
| `migrate -path path -database "..." force N` | Force version |

## Resources

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [MySQL Migration Best Practices](https://github.com/golang-migrate/migrate/blob/master/database/mysql/README.md)

