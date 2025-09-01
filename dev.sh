#!/bin/bash
set -e

# Configuration
CERT_DIR="./certs"
COMPOSE_PROJECT="pgweb"
TEST_POSTGRES_HOST="localhost"
TEST_POSTGRES_PORT="5433"
TEST_POSTGRES_USER="postgres"
TEST_POSTGRES_PASSWORD="postgres"
TEST_DATABASE="booktown"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

echo_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

echo_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

echo_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Help message
show_help() {
    cat << EOF
🐘 pgweb Development Helper

Usage: ./dev.sh [OPTIONS] [COMMAND] [DOCKER_COMPOSE_ARGS...]

Commands:
  start, up        Start development environment
  stop, down       Stop development environment
  restart          Restart and rebuild
  logs             Show logs (optional: service name)
  clean            Clean up containers and volumes
  test-setup       Set up PostgreSQL test environment
  test-smart       Run tests with smart environment detection
  test-run         Run Go tests with full setup
  test-quick       Run tests without setup
  test             Open browser test with parameters
  validate         Validate environment setup

Options:
  --https          Enable HTTPS with SSL certificates
  --cert-gen       Generate SSL certificates only
  --clean-deps     Clean Docker volumes
  -h, --help       Show this help

Examples:
  ./dev.sh start                        # Start with local PostgreSQL
  ./dev.sh --https start                # Start with HTTPS
  ./dev.sh test-smart                   # Smart test (auto-detect setup)
  ./dev.sh test-setup                   # Set up test environment only
  ./dev.sh test-quick                   # Run tests (skip setup)
  ./dev.sh logs pgweb                   # View pgweb logs
  ./dev.sh --cert-gen                   # Generate certificates only

Production Database:
  Set environment variables:
    DB_USER          - Database username
    SQL_API_PASSWORD - Database password (auto URL-encoded)
    SQL_API_HOST     - Database host
    DB_NAME          - Database name

  Example: DB_USER=user SQL_API_PASSWORD="pa\$\$@word" ./dev.sh start
EOF
}

# Function to URL encode password
url_encode_password() {
    local password="$1"
    # URL encode special characters
    encoded=$(printf "%s" "$password" | sed \
        -e "s/%/%25/g" \
        -e "s/\*/%2A/g" \
        -e "s/@/%40/g" \
        -e "s/:/%3A/g" \
        -e "s/?/%3F/g" \
        -e "s/#/%23/g" \
        -e "s/\[/%5B/g" \
        -e "s/\]/%5D/g" \
        -e "s/ /%20/g" \
        -e "s/&/%26/g" \
        -e "s/=/%3D/g" \
        -e "s/+/%2B/g")
    echo "$encoded"
}

# Detect PostgreSQL client tools path dynamically
detect_postgresql_path() {
    local pg_path=""
    
    # First try to find tools in PATH
    if command -v createdb >/dev/null 2>&1; then
        pg_path=$(dirname "$(command -v createdb)")
        echo "$pg_path"
        return 0
    fi
    
    # Common installation locations by OS
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS - check multiple locations
        local paths=(
            "/opt/homebrew/opt/postgresql@15/bin"     # Apple Silicon Homebrew
            "/opt/homebrew/opt/postgresql/bin"        # Apple Silicon Homebrew (latest)
            "/usr/local/opt/postgresql@15/bin"        # Intel Homebrew
            "/usr/local/opt/postgresql/bin"           # Intel Homebrew (latest)
            "/Applications/Postgres.app/Contents/Versions/latest/bin"  # Postgres.app
            "/usr/local/pgsql/bin"                    # Source install
        )
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux - check package manager locations
        local paths=(
            "/usr/bin"                                # Standard package install
            "/usr/local/bin"                          # Local install
            "/usr/local/pgsql/bin"                    # Source install
            "/opt/postgresql/bin"                     # Custom install
        )
    elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
        # Windows
        local paths=(
            "/c/Program Files/PostgreSQL/15/bin"
            "/c/Program Files/PostgreSQL/14/bin"
            "/c/Program Files/PostgreSQL/13/bin"
        )
    fi
    
    # Check each path
    for path in "${paths[@]}"; do
        if [[ -f "$path/createdb" ]]; then
            echo "$path"
            return 0
        fi
    done
    
    # Not found
    return 1
}

# Check if test environment is ready
check_test_environment_ready() {
    local pg_path
    pg_path=$(detect_postgresql_path)
    
    # Check PostgreSQL tools
    if [[ -z "$pg_path" ]]; then
        return 1
    fi
    
    # Check if container is running
    if ! docker ps --format "table {{.Names}}" | grep -q "pgweb-postgres"; then
        return 1
    fi
    
    # Check if booktown database exists
    local pg_env="PATH=$pg_path:\$PATH PGHOST=$TEST_POSTGRES_HOST PGPORT=$TEST_POSTGRES_PORT PGUSER=$TEST_POSTGRES_USER PGPASSWORD=$TEST_POSTGRES_PASSWORD"
    if ! eval "$pg_env psql -lqt" 2>/dev/null | cut -d \| -f 1 | grep -qw "$TEST_DATABASE"; then
        return 1
    fi
    
    return 0
}

# Cross-platform PostgreSQL client installation
install_postgresql_client() {
    echo_info "Installing PostgreSQL client tools..."
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            brew install postgresql@15
            echo_success "PostgreSQL client tools installed via Homebrew"
        else
            echo_error "Homebrew not found. Please install: https://brew.sh"
            exit 1
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        if command -v apt &> /dev/null; then
            sudo apt update && sudo apt install -y postgresql-client
            echo_success "PostgreSQL client tools installed via apt"
        elif command -v yum &> /dev/null; then
            sudo yum install -y postgresql
            echo_success "PostgreSQL client tools installed via yum"
        elif command -v dnf &> /dev/null; then
            sudo dnf install -y postgresql
            echo_success "PostgreSQL client tools installed via dnf"
        else
            echo_error "No supported package manager found"
            exit 1
        fi
    else
        echo_error "Unsupported OS: $OSTYPE"
        exit 1
    fi
}

# Auto certificate generation
generate_certificates() {
    echo_info "Generating SSL certificates..."
    mkdir -p $CERT_DIR
    
    if ! command -v mkcert &> /dev/null; then
        echo_info "Installing mkcert..."
        install_mkcert
        if ! command -v mkcert &> /dev/null; then
            echo_warning "mkcert installation failed. Using manual certificate generation..."
            generate_manual_certificates
            return
        fi
        mkcert -install
    fi
    
    mkcert -key-file $CERT_DIR/key.pem -cert-file $CERT_DIR/cert.pem localhost 127.0.0.1 ::1
    echo_success "Certificates generated in $CERT_DIR"
}

# Cross-platform mkcert installation
install_mkcert() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            brew install mkcert
        else
            echo_error "Homebrew not found. Please install: https://brew.sh"
            return 1
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        if command -v apt &> /dev/null; then
            sudo apt update && sudo apt install -y mkcert
        elif command -v snap &> /dev/null; then
            sudo snap install mkcert
        else
            echo_warning "No supported package manager found for mkcert"
            return 1
        fi
    else
        echo_warning "Unsupported OS for mkcert: $OSTYPE"
        return 1
    fi
}

# Manual certificate generation fallback
generate_manual_certificates() {
    echo_info "Generating self-signed certificates manually..."
    
    if ! command -v openssl &> /dev/null; then
        echo_error "OpenSSL not found. Cannot generate certificates."
        exit 1
    fi
    
    # Generate private key
    openssl genrsa -out "$CERT_DIR/key.pem" 2048
    
    # Generate certificate signing request config
    cat > "$CERT_DIR/cert.conf" << EOF
[req]
default_bits = 2048
prompt = no
default_md = sha256
distinguished_name = dn
req_extensions = v3_req

[dn]
CN = localhost

[v3_req]
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = *.localhost
IP.1 = 127.0.0.1
IP.2 = ::1
EOF
    
    # Generate certificate
    openssl req -new -x509 -key "$CERT_DIR/key.pem" -out "$CERT_DIR/cert.pem" \
        -days 365 -config "$CERT_DIR/cert.conf" -extensions v3_req
    
    # Cleanup config file
    rm "$CERT_DIR/cert.conf"
    
    echo_success "Self-signed certificates generated in $CERT_DIR"
    echo_warning "Note: Self-signed certificates will show security warnings"
}

# PostgreSQL test environment setup
setup_test_environment() {
    echo_info "Setting up PostgreSQL test environment..."
    
    # Install PostgreSQL client tools if needed
    local pg_path
    pg_path=$(detect_postgresql_path)
    
    if [[ -z "$pg_path" ]]; then
        echo_info "PostgreSQL client tools not found"
        install_postgresql_client
        
        # Re-detect path after installation
        pg_path=$(detect_postgresql_path)
        if [[ -z "$pg_path" ]]; then
            echo_error "Failed to detect PostgreSQL tools after installation"
            exit 1
        fi
    fi
    
    # Set PATH
    export PATH="$pg_path:$PATH"
    echo_info "PostgreSQL tools found: $pg_path"
    
    # Check if postgres container is running
    if ! docker ps --format "table {{.Names}}" | grep -q "pgweb-postgres"; then
        echo_info "Starting PostgreSQL container..."
        docker compose -f docker-compose.dev.yml up -d postgres
        
        # Wait for PostgreSQL to be ready
        echo_info "Waiting for PostgreSQL to be ready..."
        local count=0
        while ! docker exec pgweb-postgres pg_isready -U $TEST_POSTGRES_USER -h localhost >/dev/null 2>&1; do
            sleep 1
            count=$((count + 1))
            if [[ $count -gt 60 ]]; then
                echo_error "PostgreSQL failed to start within 60 seconds"
                exit 1
            fi
        done
        echo_success "PostgreSQL is ready"
    else
        echo_success "PostgreSQL container already running"
    fi
    
    # Setup test database
    echo_info "Setting up $TEST_DATABASE test database..."
    local pg_env="PGHOST=$TEST_POSTGRES_HOST PGPORT=$TEST_POSTGRES_PORT PGUSER=$TEST_POSTGRES_USER PGPASSWORD=$TEST_POSTGRES_PASSWORD"
    
    if ! eval "$pg_env psql -lqt" | cut -d \| -f 1 | grep -qw "$TEST_DATABASE"; then
        echo_info "Creating $TEST_DATABASE database..."
        eval "$pg_env createdb $TEST_DATABASE"
        
        echo_info "Loading test data..."
        eval "$pg_env psql -f data/booktown.sql $TEST_DATABASE"
        echo_success "Test database setup complete"
    else
        echo_success "$TEST_DATABASE database already exists"
    fi
    
    echo_success "Test environment ready!"
}

# Run Go tests with smart environment detection
test_smart() {
    if check_test_environment_ready; then
        echo_success "Test environment already ready, running tests directly..."
        test_quick_internal
    else
        echo_info "Setting up test environment first..."
        setup_test_environment
        test_quick_internal
    fi
}

# Run Go tests with full setup
run_tests() {
    echo_info "Running Go tests with full setup..."
    setup_test_environment
    test_quick_internal
}

# Run tests without setup (internal function)
test_quick_internal() {
    local pg_path
    pg_path=$(detect_postgresql_path)
    
    if [[ -z "$pg_path" ]]; then
        echo_error "PostgreSQL tools not found. Run './dev.sh test-setup' first"
        exit 1
    fi
    
    # Set PATH and run tests
    local test_env="PATH=$pg_path:\$PATH PGHOST=$TEST_POSTGRES_HOST PGPORT=$TEST_POSTGRES_PORT PGUSER=$TEST_POSTGRES_USER PGPASSWORD=$TEST_POSTGRES_PASSWORD PGDATABASE=$TEST_DATABASE"
    
    echo_info "Running tests..."
    # Only test packages that have test files to avoid 0% coverage noise
    local test_packages="./pkg/api ./pkg/bookmarks ./pkg/cache ./pkg/client ./pkg/command ./pkg/connection ./pkg/queries"
    eval "$test_env go test -v -race -cover $test_packages"
}

# Quick test without setup (user-facing command)
test_quick() {
    echo_info "Running tests (skipping environment setup)..."
    test_quick_internal
}

# Environment validation
validate_environment() {
    echo_info "Validating environment..."
    
    # Check Docker
    if ! docker compose version &> /dev/null; then
        echo_error "Docker Compose not available"
        exit 1
    fi
    
    # Check if docker-compose.dev.yml exists
    if [[ ! -f "docker-compose.dev.yml" ]]; then
        echo_error "docker-compose.dev.yml not found"
        exit 1
    fi
    
    # Check if test data exists
    if [[ ! -f "data/booktown.sql" ]]; then
        echo_error "Test data file data/booktown.sql not found"
        exit 1
    fi
    
    echo_success "Environment validation passed"
}

# Start development environment
start_development() {
    echo_info "Starting pgweb development environment..."
    
    # Check if production database URL is provided
    if [[ -n "$DB_USER" && -n "$SQL_API_PASSWORD" && -n "$SQL_API_HOST" && -n "$DB_NAME" ]]; then
        echo_info "Using production database configuration"
        
        # URL encode the password
        local encoded_password
        encoded_password=$(url_encode_password "$SQL_API_PASSWORD")
        
        # Build database URL
        local ssl_mode="require"
        if [[ "${SSL_DISABLE:-}" == "true" ]]; then
            ssl_mode="disable"
        fi
        DATABASE_URL="postgres://$DB_USER:$encoded_password@$SQL_API_HOST:5432/$DB_NAME?sslmode=$ssl_mode"
        
        # Mask passwords for logging
        local password_masked encoded_masked
        password_masked=$(echo "$SQL_API_PASSWORD" | sed 's/./*/g')
        encoded_masked=$(echo "$encoded_password" | sed 's/./*/g')
        
        echo_info "Database user: $DB_USER"
        echo_info "Database host: $SQL_API_HOST"
        echo_info "Database name: $DB_NAME"
        echo_info "Original password: $password_masked"
        echo_info "Encoded password: $encoded_masked"
        echo_info "SSL mode: $ssl_mode"
        
        # Export for docker compose
        export DATABASE_URL
    else
        echo_info "Using local PostgreSQL container"
    fi
    
    # Generate SSL certificates if HTTPS mode and certificates don't exist
    if [[ "${HTTPS_ENABLED:-}" == "true" && ! -f "$CERT_DIR/cert.pem" ]]; then
        echo_warning "HTTPS enabled but certificates missing"
        generate_certificates
    fi
    
    # Start services
    docker compose -f docker-compose.dev.yml up -d --build
    
    echo_success "Services started!"
    echo_info "pgweb: http://localhost:8081"
    if [[ "${HTTPS_ENABLED:-}" == "true" ]]; then
        echo_info "pgweb (HTTPS): https://localhost:8081"
    fi
    if [[ -z "$DATABASE_URL" ]]; then
        echo_info "postgres: localhost:5433"
    fi
    
    # Test parameter substitution URL
    echo_info "Test URL with parameters:"
    echo_info "http://localhost:8081/?Client=client&Instance=instance&ClientName=client-name&InstanceName=instance-name&AccountId=account-id"
}

# Stop development environment
stop_development() {
    echo_info "Stopping pgweb development environment..."
    docker compose -f docker-compose.dev.yml down
    echo_success "Services stopped!"
}

# Restart development environment
restart_development() {
    echo_info "Restarting pgweb development environment..."
    docker compose -f docker-compose.dev.yml down
    docker compose -f docker-compose.dev.yml up -d --build
    echo_success "Services restarted!"
}

# Show logs
show_logs() {
    local service="${1:-}"
    echo_info "Showing logs..."
    docker compose -f docker-compose.dev.yml logs -f $service
}

# Clean up environment
clean_environment() {
    echo_warning "This will remove all containers and volumes!"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker compose -f docker-compose.dev.yml down -v
        docker system prune -f
        echo_success "Cleanup complete!"
    else
        echo_info "Cleanup cancelled"
    fi
}

# Open browser test
test_browser() {
    echo_info "Testing parameter substitution..."
    sleep 2
    echo_info "Opening pgweb with test parameters..."
    local test_url="http://localhost:8081/?Client=client&Instance=instance&ClientName=client-name&InstanceName=instance-name&AccountId=account-id&AccountPerspective=account-perspective&AccountDbUser=account-db-user&AccountName=account-name&AccountEmail=account-email&FolderName=folder-name&InvalidParameter=shouldnotshow"
    open "$test_url" 2>/dev/null || echo_warning "Could not open browser automatically"
    echo_info "Test URL: $test_url"
}

# Main execution
main() {
    local command="$1"
    shift 2>/dev/null || true
    
    case "$command" in
        "start"|"up")
            validate_environment
            start_development
            ;;
        "stop"|"down")
            stop_development
            ;;
        "restart")
            validate_environment
            restart_development
            ;;
        "logs")
            show_logs "$1"
            ;;
        "clean")
            clean_environment
            ;;
        "test-setup")
            validate_environment
            setup_test_environment
            ;;
        "test-smart")
            validate_environment
            test_smart
            ;;
        "test-run")
            validate_environment
            run_tests
            ;;
        "test-quick")
            test_quick
            ;;
        "test")
            test_browser
            ;;
        "validate")
            validate_environment
            ;;
        "")
            show_help
            ;;
        *)
            # Pass through to docker compose
            validate_environment
            docker compose -f docker-compose.dev.yml "$command" "$@"
            ;;
    esac
}

# Parse arguments
HTTPS_ENABLED=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --https)
            HTTPS_ENABLED=true
            export HTTPS_ENABLED
            shift
            ;;
        --cert-gen)
            generate_certificates
            exit 0
            ;;
        --clean-deps)
            echo_info "Cleaning Docker volumes..."
            docker compose -f docker-compose.dev.yml down -v
            echo_success "Volumes cleaned!"
            exit 0
            ;;
        --help|-h)
            show_help
            exit 0
            ;;
        *)
            break
            ;;
    esac
done

main "$@"