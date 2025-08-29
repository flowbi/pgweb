#!/bin/bash

# Simple script to start pgweb development environment
# Handles password encoding automatically

set -e

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

# Default action
ACTION="${1:-start}"

case "$ACTION" in
    "start"|"up")
        echo_info "Starting pgweb development environment..."
        
        # Check if production database URL is provided
        if [[ -n "$DB_USER" && -n "$SQL_API_PASSWORD" && -n "$SQL_API_HOST" && -n "$DB_NAME" ]]; then
            echo_info "Using production database configuration"
            
            # URL encode the password
            ENCODED_PASSWORD=$(url_encode_password "$SQL_API_PASSWORD")
            
            # Build database URL
            DATABASE_URL="postgres://$DB_USER:$ENCODED_PASSWORD@$SQL_API_HOST:5432/$DB_NAME?sslmode=require"
            
            # Mask passwords for logging
            PASSWORD_MASKED=$(echo "$SQL_API_PASSWORD" | sed 's/./*/g')
            ENCODED_MASKED=$(echo "$ENCODED_PASSWORD" | sed 's/./*/g')
            
            echo_info "Database user: $DB_USER"
            echo_info "Database host: $SQL_API_HOST"
            echo_info "Database name: $DB_NAME"
            echo_info "Original password: $PASSWORD_MASKED"
            echo_info "Encoded password: $ENCODED_MASKED"
            echo_info "Attempting connection..."
            
            # Export for docker-compose
            export DATABASE_URL
        else
            echo_info "Using local PostgreSQL container (no production DB vars found)"
        fi
        
        # Start services
        docker-compose -f docker-compose.dev.yml up -d --build
        
        echo_success "Services started!"
        echo_info "pgweb: http://localhost:8081"
        if [[ -z "$DATABASE_URL" ]]; then
            echo_info "postgres: localhost:5433"
        fi
        
        # Test parameter substitution
        echo_info "Test URL with parameters:"
        echo_info "http://localhost:8081/?Client=client&Instance=instance&ClientName=clientname&InstanceName=instance-name&AccountId=account-id&AccountPerspective=account-perspective&AccountDbUser=account-db-user&AccountName=account-name&AccountEmail=account-email&FolderName=folder-name&InvalidParameter=shouldnotshow"
        ;;
        
    "stop"|"down")
        echo_info "Stopping pgweb development environment..."
        docker-compose -f docker-compose.dev.yml down
        echo_success "Services stopped!"
        ;;
        
    "restart")
        echo_info "Restarting pgweb development environment..."
        docker-compose -f docker-compose.dev.yml down
        docker-compose -f docker-compose.dev.yml up -d --build
        echo_success "Services restarted!"
        ;;
        
    "logs")
        echo_info "Showing logs..."
        docker-compose -f docker-compose.dev.yml logs -f ${2:-}
        ;;
        
    "clean")
        echo_warning "This will remove all containers and volumes!"
        read -p "Are you sure? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker-compose -f docker-compose.dev.yml down -v
            docker system prune -f
            echo_success "Cleanup complete!"
        else
            echo_info "Cleanup cancelled"
        fi
        ;;
        
    "test")
        echo_info "Testing parameter substitution..."
        sleep 2
        echo_info "Opening pgweb with test parameters..."
        open "http://localhost:8081/?Client=client&Instance=instance&ClientName=clientname&InstanceName=instance-name&AccountId=account-id&AccountPerspective=account-perspective&AccountDbUser=account-db-user&AccountName=account-name&AccountEmail=account-email&FolderName=folder-name&InvalidParameter=shouldnotshow" 2>/dev/null || echo_warning "Could not open browser automatically"
        ;;
        
    "help"|*)
        echo_info "pgweb Development Helper"
        echo ""
        echo "Usage: ./dev.sh <command>"
        echo ""
        echo "Commands:"
        echo "  start    - Start development environment"
        echo "  stop     - Stop development environment"  
        echo "  restart  - Restart and rebuild"
        echo "  logs     - Show logs (optional: service name)"
        echo "  clean    - Clean up containers and volumes"
        echo "  test     - Test parameter substitution"
        echo "  help     - Show this help"
        echo ""
        echo "For production database, set these environment variables:"
        echo "  DB_USER          - Database username"
        echo "  SQL_API_PASSWORD - Database password (special chars auto-encoded)"
        echo "  SQL_API_HOST     - Database host"
        echo "  DB_NAME          - Database name"
        echo ""
        echo "Examples:"
        echo "  ./dev.sh start"
        echo "  ./dev.sh logs pgweb"
        echo "  ./dev.sh test"
        echo ""
        echo "Production example:"
        echo '  DB_USER=myuser SQL_API_PASSWORD="my*pa$$@word" SQL_API_HOST=db.example.com DB_NAME=mydb ./dev.sh start'
        ;;
esac