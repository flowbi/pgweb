# pgweb Development Setup

Simple one-command development environment for pgweb with PostgreSQL, testing, and parameter substitution.

## Quick Start

```bash
# Start development environment (builds automatically)
./dev.sh start

# Access pgweb: http://localhost:8081
# Access postgres: localhost:5433
```

## Commands

### Development Commands

```bash
./dev.sh start          # Start development environment
./dev.sh stop           # Stop development environment
./dev.sh restart        # Restart and rebuild
./dev.sh logs           # Show logs (optional: service name)
./dev.sh clean          # Clean up containers and volumes
./dev.sh --clean-deps   # Clean Docker volumes only
```

### Testing Commands

```bash
./dev.sh test-smart     # Smart test (auto-detect if setup needed)
./dev.sh test-setup     # Set up PostgreSQL test environment
./dev.sh test-run       # Run Go tests with full setup
./dev.sh test-quick     # Run tests (skip setup)
./dev.sh test           # Open browser test with parameters
make test               # Smart test via Makefile
make test-quick         # Quick test via Makefile
make test-setup         # Setup test environment only
```

### HTTPS Support

```bash
./dev.sh --https start  # Start with HTTPS (auto-generates certificates)
./dev.sh --cert-gen     # Generate SSL certificates only
```

### Utility Commands

```bash
./dev.sh validate       # Validate environment setup
./dev.sh help           # Show help
```

## Testing

The enhanced testing system features **smart environment detection** and **cross-platform support**:

### Smart Detection:

1. **Auto-detects** if PostgreSQL tools, container, and database are ready
2. **Skips setup** if environment is already configured
3. **Runs full setup** only when needed
4. **Cross-platform** path detection for PostgreSQL tools

### One-Command Testing:

```bash
make test               # Smart detection - fast if already set up
```

### Manual Testing Options:

```bash
./dev.sh test-smart     # Smart detection (recommended)
./dev.sh test-setup     # Setup environment only
./dev.sh test-quick     # Run tests (no setup)
./dev.sh test-run       # Full setup + run tests
```

### Cross-Platform Support:

- **macOS**: Homebrew (Intel/ARM), Postgres.app, source installs
- **Linux**: apt, yum, dnf package managers, source installs
- **Windows**: Git Bash, Cygwin support
- **Dynamic Detection**: Uses `which` command + common paths

### Test Environment Details:

- **PostgreSQL**: localhost:5433 (postgres/postgres)
- **Test Database**: booktown (loaded from data/booktown.sql)
- **Auto-Discovery**: Finds PostgreSQL tools in standard locations

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
export SSL_DISABLE="true"  # Optional: disable SSL

# Start with production database
./dev.sh start
```

## HTTPS Development

For HTTPS development with trusted certificates:

```bash
# Generate certificates (installs mkcert if needed)
./dev.sh --cert-gen

# Start with HTTPS
./dev.sh --https start

# Access: https://localhost:8081
```

**Certificate Features:**

- **Auto-installation**: Installs `mkcert` via package manager
- **Cross-platform**: Supports macOS (Homebrew), Linux (apt/yum/dnf/snap)
- **Fallback**: Uses OpenSSL for manual certificate generation
- **Trusted**: mkcert certificates are trusted by browsers

## Docker Compose Only

If you prefer direct docker-compose:

```bash
# Local development
docker-compose -f docker-compose.dev.yml up -d

# Production database
DATABASE_URL="postgres://user:encoded_password@host:5432/db" docker-compose -f docker-compose.dev.yml up -d
```

## How It Works

1. **No Separate Build Step**: Docker builds the binary automatically during `docker-compose up`
2. **Automatic Password Encoding**: Special characters in passwords are URL-encoded automatically
3. **Frontend Parameter Substitution**: JavaScript replaces `@parameters` before sending queries to backend
4. **Auto Test Setup**: PostgreSQL client tools and test database setup automatically
5. **Cross-platform Support**: Auto-detects paths and package managers

## Files

- `docker-compose.dev.yml` - Development configuration with PostgreSQL
- `dev.sh` - Enhanced helper script with test setup, HTTPS, and password encoding
- `Makefile` - Build and test commands
- `.env` - Secrets (never committed)
- `.env.example` - Template for environment variables
- `data/booktown.sql` - Test database schema and data

## Troubleshooting

### PostgreSQL Client Tools Not Found

```bash
# Auto-install via dev.sh (cross-platform)
./dev.sh test-setup

# Or install manually by platform:
# macOS: brew install postgresql@15
# Ubuntu/Debian: sudo apt install postgresql-client
# RHEL/CentOS: sudo yum install postgresql
# Fedora: sudo dnf install postgresql
# Arch Linux: sudo pacman -S postgresql
```

### Docker Issues

```bash
# Check Docker Compose
docker compose version

# Clean and restart
./dev.sh clean
./dev.sh start
```

### Test Database Issues

```bash
# Reset test environment
./dev.sh clean
./dev.sh test-setup
```
