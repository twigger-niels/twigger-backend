#!/bin/bash

# Test Cloud SQL connection script
# Usage: ./scripts/test-connection.sh [public-ip]

if [ -z "$1" ]; then
    echo "Usage: $0 <cloud-sql-public-ip>"
    echo "Example: $0 34.123.45.67"
    exit 1
fi

PUBLIC_IP=$1
echo "Testing connection to Cloud SQL instance at $PUBLIC_IP..."

# Test basic connectivity
echo "1. Testing port connectivity..."
nc -zv $PUBLIC_IP 5432

# Test PostgreSQL connection (requires psql)
echo "2. Testing PostgreSQL connection..."
echo "Enter the postgres user password when prompted"
psql -h $PUBLIC_IP -U postgres -d postgres -c "SELECT version();"

# Test with application connection string
echo "3. Testing with Go application..."
export DATABASE_URL="postgres://postgres:YOUR_PASSWORD@$PUBLIC_IP:5432/postgres?sslmode=require"
go run cmd/migrate/main.go up