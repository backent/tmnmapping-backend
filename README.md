# TMN Backend - Authentication Service

Complete Go backend service with JWT authentication built using clean architecture.

## Features

✅ **JWT Authentication** - Secure token-based auth with HTTP-only cookies  
✅ **Clean Architecture** - Separated layers (Controllers, Services, Repositories)  
✅ **PostgreSQL Database** - User management with migrations  
✅ **Password Hashing** - Bcrypt encryption  
✅ **Protected Routes** - Middleware-based auth  
✅ **Google Wire** - Dependency injection  
✅ **Input Validation** - Request validation with go-playground/validator  

## Project Structure

```
backend/
├── controllers/          # HTTP handlers
├── services/             # Business logic
├── repositories/         # Data access
├── middlewares/          # HTTP middlewares
├── models/               # Domain models
├── web/                  # DTOs
├── helpers/              # Utilities
├── exceptions/           # Custom errors
├── libs/                 # Core libraries
├── injector/             # Dependency injection
├── database/
│   └── migrations/       # SQL migrations
└── main.go              # Entry point
```

## Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Make (optional)

## Setup

### 1. Clone and Install Dependencies

```bash
cd backend
go mod download
```

### 2. Configure Environment

Copy the example environment file:

```bash
cp env.example .env
```

Edit `.env` with your configuration:

```bash
# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=adminlocal
POSTGRES_DATABASE=tmn_backend

# Application
APP_PORT=8088
APP_SECRET_KEY=your-secret-key-min-32-characters
APP_TOKEN_EXPIRE_IN_SEC=3600
```

### 3. Create Database

```bash
psql -U postgres -h localhost
CREATE DATABASE tmn_backend;
\q
```

### 4. Run Migrations

#### Option A: Using golang-migrate (Recommended)

Install [golang-migrate](https://github.com/golang-migrate/migrate):

```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Or install with Go
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

**Run migrations:**

```bash
# Run all pending migrations
migrate -path database/migrations -database "postgres://postgres:adminlocal@localhost:5432/tmn_backend?sslmode=disable" up

# Run specific number of migrations
migrate -path database/migrations -database "postgres://postgres:adminlocal@localhost:5432/tmn_backend?sslmode=disable" up 1

# Rollback last migration
migrate -path database/migrations -database "postgres://postgres:adminlocal@localhost:5432/tmn_backend?sslmode=disable" down 1
```

**Using Docker (no installation needed):**

```bash
# Single line
docker run -v $(pwd)/database/migrations:/migrations --network host migrate/migrate -path=/migrations -database "postgres://postgres:adminlocal@localhost:5432/tmn_backend?sslmode=disable" up

# Multi-line (better readability)
docker run -v $(pwd)/database/migrations:/migrations --network host migrate/migrate \
  -path=/migrations \
  -database "postgres://postgres:adminlocal@localhost:5432/tmn_backend?sslmode=disable" \
  up
```

#### Option B: Using PostgreSQL directly

```bash
psql -U postgres -h localhost -d tmn_backend -f database/migrations/001_create_users_table.up.sql
```

### 5. Seed Test User

```bash
# Generate password hash (you can use any bcrypt tool or this Go snippet)
# Password: "password123"
# Hash: $2a$10$rBV2NZXhzY.Db5fQHHNMF.KKlXoXqM0K7vJZCJJJJJJJJJJJJJJJJ

mysql -u root -p tmn_backend
INSERT INTO users (username, name, password, role) 
VALUES ('admin', 'Administrator', '$2a$10$rBV2NZXhzY.Db5fQHHNMF.KKlXoXqM0K7vJZCJJJJJJJJJJJJJJ', 'admin');
```

## Running the Application

### Development Mode

```bash
go run main.go
```

The server will start on `http://localhost:8088`

### Build and Run

```bash
go build -o tmn-backend
./tmn-backend
```

## API Endpoints

### Public Endpoints

#### POST /login
Login with username and password.

**Request:**
```json
{
  "username": "admin",
  "password": "password123"
}
```

**Response:**
```json
{
  "status": "OK",
  "code": 200,
  "data": {
    "user": {
      "id": 1,
      "username": "admin",
      "name": "Administrator",
      "role": "admin"
    }
  }
}
```

**Cookie:** Sets `auth_token` HTTP-only cookie

#### POST /logout
Clear authentication cookie.

**Response:**
```json
{
  "status": "OK",
  "code": 200,
  "data": "logged out successfully"
}
```

### Protected Endpoints

#### GET /current-user
Get current authenticated user info.

**Headers:** Requires `auth_token` cookie

**Response:**
```json
{
  "status": "OK",
  "code": 200,
  "data": {
    "id": 1,
    "username": "admin",
    "name": "Administrator",
    "role": "admin"
  }
}
```

### Health Check

#### GET /health
Health check endpoint.

**Response:** `OK`

## Testing

### Using cURL

**Login:**
```bash
curl -X POST http://localhost:8088/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}' \
  -c cookies.txt
```

**Get Current User:**
```bash
curl http://localhost:8088/current-user -b cookies.txt
```

**Logout:**
```bash
curl -X POST http://localhost:8088/logout -b cookies.txt
```

### Generate Password Hash

```go
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    password := "your_password"
    hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
    fmt.Println(string(hash))
}
```

## Architecture

### Request Flow

```
Request → Router → Middleware → Controller → Service → Repository → Database
                                    ↓
                                Response
```

### Clean Architecture Layers

1. **Controllers** - HTTP handlers, request/response
2. **Services** - Business logic, transactions
3. **Repositories** - Data access, SQL queries
4. **Models** - Domain entities
5. **Web** - DTOs (Data Transfer Objects)
6. **Middlewares** - Request validation, auth
7. **Helpers** - Utility functions
8. **Exceptions** - Custom errors

### Dependency Injection

Using Google Wire for automatic dependency injection:

```go
// injector/wire.go
func InitializeRouter() *httprouter.Router {
    wire.Build(
        libs.NewDB,
        libs.NewValidator,
        repositoriesAuth.NewRepositoryAuthJWTImpl,
        repositoriesUser.NewRepositoryUserImpl,
        servicesAuth.NewServiceAuthImpl,
        controllersAuth.NewControllerAuthImpl,
        middlewares.NewAuthMiddleware,
        libs.NewRouter,
    )
    return nil
}
```

## Security Features

- **Password Hashing:** Bcrypt with cost factor 10
- **JWT Tokens:** HS256 algorithm with configurable expiry
- **HTTP-only Cookies:** Prevents XSS attacks
- **Token Validation:** Middleware-based protection
- **Panic Recovery:** Centralized error handling

## Error Handling

The application uses panic/recover pattern with custom exceptions:

- `BadRequestError` - 400 (validation errors)
- `Unauthorized` - 401 (auth failures)
- `NotFoundError` - 404 (resource not found)

All panics are caught by the router's panic handler and converted to appropriate HTTP responses.

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(50) NOT NULL,
    email VARCHAR(50),
    password VARCHAR(200) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Development

### Creating New Migrations

Use golang-migrate to create migration files:

```bash
# Create a new migration
migrate create -ext sql -dir database/migrations -seq create_new_table

# This creates two files:
# - 000002_create_new_table.up.sql    (for applying the migration)
# - 000002_create_new_table.down.sql  (for rolling back)
```

**Example migration files:**

`002_create_posts_table.up.sql`:
```sql
CREATE TABLE IF NOT EXISTS posts (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_posts_user_id FOREIGN KEY (user_id) REFERENCES users(id)
);
```

`002_create_posts_table.down.sql`:
```sql
DROP TABLE IF EXISTS posts;
```

### Adding New Features

Follow the clean architecture pattern:

1. Create model in `models/`
2. Create repository interface and implementation
3. Create service interface and implementation
4. Create controller
5. Add route in `libs/router.go`
6. Update `injector/wire.go`
7. Run `wire` to regenerate `wire_gen.go`

### Code Generation with Wire

```bash
go install github.com/google/wire/cmd/wire@latest
cd injector
wire
```

## Troubleshooting

### Database Connection Error

- Check PostgreSQL is running
- Verify credentials in `.env`
- Ensure database exists

### Token Validation Fails

- Check `APP_SECRET_KEY` is set
- Verify token hasn't expired
- Check cookie is being sent

### Import Errors

```bash
go mod tidy
go mod download
```

## Production Deployment

### Build

```bash
CGO_ENABLED=0 GOOS=linux go build -o tmn-backend
```

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o tmn-backend

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/tmn-backend .
COPY --from=builder /app/.env .
EXPOSE 8088
CMD ["./tmn-backend"]
```

## Next Steps

- Add user CRUD operations
- Implement role-based access control
- Add refresh tokens
- Add API rate limiting
- Add logging with structured logs
- Add metrics and monitoring

## License

MIT

## Author

Built following TMN Backend Documentation

