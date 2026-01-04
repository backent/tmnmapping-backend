# Quick Start Guide

## 5-Minute Setup

### 1. Create Database
```bash
psql -U postgres -h localhost
CREATE DATABASE tmn_backend;
\q
```

### 2. Run Migration

#### Option A: Using golang-migrate (Recommended)

```bash
# Install migrate (if not installed)
# macOS: brew install golang-migrate
# Linux: see README.md for installation

# Run migration
migrate -path database/migrations \
  -database "postgres://postgres:adminlocal@localhost:5432/tmn_backend?sslmode=disable" \
  up
```

**Or with Docker (no installation):**

```bash
docker run -v $(pwd)/database/migrations:/migrations --network host migrate/migrate \
  -path=/migrations \
  -database "postgres://postgres:adminlocal@localhost:5432/tmn_backend?sslmode=disable" \
  up
```

#### Option B: Using PostgreSQL directly

```bash
psql -U postgres -h localhost -d tmn_backend -f database/migrations/001_create_users_table.up.sql
```

### 3. Create Test User
```bash
# Hash for password "password123"
psql -U postgres -h localhost -d tmn_backend
INSERT INTO users (username, name, password, role) 
VALUES ('admin', 'Administrator', '$2a$10$N9qo8uLOickgx2ZMRZoMye5xvJ5UYEPlL7O0IYdH7g3LVXh9dkQvq', 'admin');
\q
```

### 4. Configure Environment
```bash
cp env.example .env
# Edit .env with your database credentials
```

Example `.env`:
```bash
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=adminlocal
POSTGRES_DATABASE=tmn_backend

APP_PORT=8088
APP_SECRET_KEY=my-super-secret-key-min-32-chars-long
APP_TOKEN_EXPIRE_IN_SEC=3600
```

### 5. Download Dependencies
```bash
go mod download
```

### 6. Run Application
```bash
go run main.go
```

Server starts on: `http://localhost:8088`

## Test the API

### Login
```bash
curl -X POST http://localhost:8088/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}' \
  -c cookies.txt -v
```

### Get Current User
```bash
curl http://localhost:8088/current-user -b cookies.txt
```

### Logout
```bash
curl -X POST http://localhost:8088/logout -b cookies.txt
```

## Generate Your Own Password Hash

Create a file `hash_password.go`:

```go
package main

import (
    "fmt"
    "os"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run hash_password.go <password>")
        return
    }
    
    password := os.Args[1]
    hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Password:", password)
    fmt.Println("Hash:", string(hash))
}
```

Run it:
```bash
go run hash_password.go mypassword
```

## Troubleshooting

**Can't connect to database?**
```bash
# Check MySQL is running
mysql -u root -p -e "SELECT 1;"

# Check database exists
mysql -u root -p -e "SHOW DATABASES;"
```

**Import errors?**
```bash
go mod tidy
go mod download
```

**Port already in use?**
```bash
# Change APP_PORT in .env
APP_PORT=8089
```

## Next Steps

1. ✅ Basic auth is working
2. Read the full [README.md](README.md)
3. Check the [backend documentation](../docs/backend/)
4. Connect frontend to this backend

## Common Commands

```bash
# Build binary
go build -o tmn-backend

# Run binary
./tmn-backend

# Run with specific port
APP_PORT=9000 go run main.go

# Check all endpoints
curl http://localhost:8088/health
```

## Success! 🎉

You now have a working authentication backend with:
- JWT authentication
- HTTP-only cookies
- Password hashing
- Protected routes
- Clean architecture

Ready to connect to your Vue.js frontend!

