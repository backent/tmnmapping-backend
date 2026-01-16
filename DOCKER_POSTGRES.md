# PostgreSQL with PostGIS Docker Setup

This Dockerfile creates a PostgreSQL 17 image with PostGIS extension installed, compatible with both ARM64 and x86_64 architectures.

## Building the Image

```bash
cd backend
docker build -f Dockerfile.postgres -t postgres-postgis:17 .
```

## Running the Container

### Basic Run
```bash
docker run -d \
  --name postgres-postgis \
  -e POSTGRES_PASSWORD=adminlocal \
  -e POSTGRES_DB=tmn_backend \
  -p 5432:5432 \
  postgres-postgis:17
```

### With Volume Persistence
```bash
docker run -d \
  --name postgres-postgis \
  -e POSTGRES_PASSWORD=adminlocal \
  -e POSTGRES_DB=tmn_backend \
  -p 5432:5432 \
  -v postgres_data:/var/lib/postgresql/data \
  postgres-postgis:17
```

## Verifying PostGIS Installation

After the container starts, verify PostGIS is available:

```bash
# Connect to the database
docker exec -it postgres-postgis psql -U postgres -d tmn_backend

# Check available extensions
SELECT * FROM pg_available_extensions WHERE name = 'postgis';

# Create the extension (this will be done automatically by your migration)
CREATE EXTENSION IF NOT EXISTS postgis;
```

## Migration from Existing Container

If you have an existing PostgreSQL container:

1. **Backup your data** (if needed):
   ```bash
   docker exec <old_container> pg_dump -U postgres tmn_backend > backup.sql
   ```

2. **Stop and remove old container**:
   ```bash
   docker stop <old_container>
   docker rm <old_container>
   ```

3. **Start new PostGIS-enabled container**:
   ```bash
   docker run -d \
     --name postgres-postgis \
     -e POSTGRES_PASSWORD=adminlocal \
     -e POSTGRES_DB=tmn_backend \
     -p 5432:5432 \
     -v postgres_data:/var/lib/postgresql/data \
     postgres-postgis:17
   ```

4. **Restore data** (if you backed up):
   ```bash
   docker exec -i postgres-postgis psql -U postgres -d tmn_backend < backup.sql
   ```

5. **Run migrations**:
   ```bash
   docker run -v $(pwd)/database/migrations:/migrations --network host migrate/migrate \
     -path=/migrations/ \
     -database "postgres://postgres:adminlocal@127.0.0.1:5432/tmn_backend?sslmode=disable" \
     up
   ```

## Architecture Support

This Dockerfile uses the official `postgres:17` image which supports:
- ✅ x86_64 (Intel/AMD)
- ✅ ARM64 (Apple Silicon, ARM servers)

The PostGIS packages from Debian repositories work on both architectures.
