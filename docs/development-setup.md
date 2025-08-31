# pgweb Development Setup

Simple one-command development environment for pgweb with PostgreSQL and parameter substitution.

## Quick Start

```bash
# Start development environment (builds automatically)
./dev.sh start

# Access pgweb: http://localhost:8081
# Access postgres: localhost:5433
```

## Commands

```bash
./dev.sh start    # Start development environment
./dev.sh stop     # Stop development environment
./dev.sh restart  # Restart and rebuild
./dev.sh logs     # Show logs (add service name: logs pgweb)
./dev.sh clean    # Clean up containers and volumes
./dev.sh test     # Test parameter substitution
./dev.sh help     # Show help
```

## Parameter Substitution

The system automatically replaces `@parameter` placeholders in SQL queries with URL parameters.

**Example:**

- URL: `http://localhost:8081/?Client=client&Instance=instance&ClientName=client-name&InstanceName=instance-name&AccountId=account-id&AccountPerspective=account-perspective&AccountDbUser=account-db-user&AccountName=account-name&AccountEmail=account-email&FolderName=folder-name&InvalidParameter=shouldnotshow`
- Query: `SELECT * FROM table WHERE client = @Client AND instance = @Instance`
- Executed: `SELECT * FROM table WHERE client = 'test-client' AND instance = 'test-instance'`

**Custom Parameters:**

Parameters are configurable via the `PGWEB_CUSTOM_PARAMS` environment variable. Default parameters include:

- `Client`, `Instance`, `ClientName`, `InstanceName`
- `AccountId`, `AccountPerspective`, `AccountDbUser`
- `AccountName`, `AccountEmail`, `FolderName`

You can customize these by setting `PGWEB_CUSTOM_PARAMS` in your `.env` file with a comma-separated list of parameter names.

## Production Database

For production database with special characters in password:

```bash
# Set environment variables
export DB_USER="your_username"
export SQL_API_PASSWORD="your*pa$$@word"  # Special chars auto-encoded
export SQL_API_HOST="your.database.host"
export DB_NAME="your_database"

# Start with production database
./dev.sh start
```

## Docker Compose Only

If you prefer direct docker-compose:

```bash
# Local development
docker-compose up -d

# Production database
DATABASE_URL="postgres://user:encoded_password@host:5432/db" docker-compose up -d
```

## How It Works

1. **No Separate Build Step**: Docker builds the binary automatically during `docker-compose up`
2. **Automatic Password Encoding**: Special characters in passwords are URL-encoded automatically
3. **Frontend Parameter Substitution**: JavaScript replaces `@parameters` before sending queries to backend
4. **One File**: Everything configured in single `docker-compose.yml`

## Files

- `docker-compose.yml` - Main configuration (safe to commit)
- `dev.sh` - Helper script with password encoding
- `.env` - Secrets (never committed)
- `.env.example` - Template for environment variables
