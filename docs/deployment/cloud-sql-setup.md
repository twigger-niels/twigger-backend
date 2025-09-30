# Cloud SQL Setup Guide

## Complete Setup Process for PostgreSQL 17 with PostGIS

This guide covers the complete setup process for Google Cloud SQL PostgreSQL 17 instance with PostGIS extensions for the plant database.

## Prerequisites

- Google Cloud Project with billing enabled
- `gcloud` CLI installed and configured
- Appropriate IAM permissions for Cloud SQL

### Required IAM Roles
```bash
# Minimum required roles
roles/cloudsql.admin
roles/cloudsql.client
roles/compute.networkAdmin  # For VPC setup
roles/serviceusage.serviceUsageAdmin
```

## 1. Initial Cloud SQL Instance Creation

### Create PostgreSQL 17 Instance

```bash
# Set environment variables
export PROJECT_ID="twigger"
export INSTANCE_ID="dev-twigger-db1"
export REGION="us-central1"
export DB_VERSION="POSTGRES_17"

# Create the instance
gcloud sql instances create $INSTANCE_ID \
    --database-version=$DB_VERSION \
    --tier=db-custom-1-3840 \
    --region=$REGION \
    --storage-type=SSD \
    --storage-size=20GB \
    --storage-auto-increase \
    --backup-start-time=02:00 \
    --backup-location=us \
    --maintenance-window-day=SUN \
    --maintenance-window-hour=06 \
    --maintenance-release-channel=production \
    --availability-type=zonal \
    --enable-bin-log \
    --deletion-protection \
    --project=$PROJECT_ID
```

### Instance Configuration Details

**Performance Tier:** `db-custom-1-3840`
- 1 vCPU
- 3.75 GB RAM
- Suitable for development/testing

**Storage Configuration:**
- SSD storage (better performance than HDD)
- 20GB initial size with auto-increase
- Automatic storage increase prevents outages

**Backup Configuration:**
- Daily backups at 02:00 UTC
- 7-day backup retention (default)
- Point-in-time recovery enabled
- Cross-region backup location for disaster recovery

## 2. Network Security Configuration

### Configure Authorized Networks

```bash
# Add your development IP
export DEV_IP="82.217.141.244"

gcloud sql instances patch $INSTANCE_ID \
    --authorized-networks=$DEV_IP/32 \
    --project=$PROJECT_ID

# Add network range for flexibility
gcloud sql instances patch $INSTANCE_ID \
    --authorized-networks=$DEV_IP/32,$DEV_IP.0/24 \
    --project=$PROJECT_ID
```

### Private IP Configuration (Production)

```bash
# Create VPC network (if not exists)
gcloud compute networks create twigger-vpc \
    --subnet-mode=custom \
    --project=$PROJECT_ID

# Create subnet
gcloud compute networks subnets create twigger-subnet \
    --network=twigger-vpc \
    --range=10.0.0.0/24 \
    --region=$REGION \
    --project=$PROJECT_ID

# Allocate IP range for Google services
gcloud compute addresses create google-managed-services-twigger-vpc \
    --global \
    --purpose=VPC_PEERING \
    --prefix-length=16 \
    --network=twigger-vpc \
    --project=$PROJECT_ID

# Create private connection
gcloud services vpc-peerings connect \
    --service=servicenetworking.googleapis.com \
    --ranges=google-managed-services-twigger-vpc \
    --network=twigger-vpc \
    --project=$PROJECT_ID

# Configure instance for private IP
gcloud sql instances patch $INSTANCE_ID \
    --network=projects/$PROJECT_ID/global/networks/twigger-vpc \
    --no-assign-ip \
    --project=$PROJECT_ID
```

## 3. Database and User Setup

### Set Root Password

```bash
# Set a strong password for postgres user
gcloud sql users set-password postgres \
    --instance=$INSTANCE_ID \
    --password='your-secure-password-here' \
    --project=$PROJECT_ID
```

### Create Application Database

```bash
# Create the main database
gcloud sql databases create twigger_db \
    --instance=$INSTANCE_ID \
    --project=$PROJECT_ID

# Create application user
gcloud sql users create twigger_app \
    --instance=$INSTANCE_ID \
    --password='app-user-password' \
    --project=$PROJECT_ID
```

### Create IAM Database User (Recommended)

```bash
# Create service account
gcloud iam service-accounts create twigger-db-user \
    --display-name="Twigger Database User" \
    --project=$PROJECT_ID

# Create IAM database user
gcloud sql users create twigger-db-user@$PROJECT_ID.iam.gserviceaccount.com \
    --instance=$INSTANCE_ID \
    --type=cloud_iam_service_account \
    --project=$PROJECT_ID
```

## 4. PostGIS Extensions Setup

### Enable Required Extensions

```sql
-- Connect to the database and run:
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE EXTENSION IF NOT EXISTS ltree;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Verify PostGIS installation
SELECT PostGIS_Version();
SELECT PostGIS_Full_Version();
```

### Extension Verification Script

```bash
# Create verification script
cat > verify_extensions.sql << 'EOF'
\echo 'Checking PostGIS extensions...'
SELECT name, default_version, installed_version
FROM pg_available_extensions
WHERE name LIKE 'postgis%';

\echo 'Checking PostGIS functions...'
SELECT PostGIS_Version();
SELECT PostGIS_GEOS_Version();
SELECT PostGIS_PROJ_Version();

\echo 'Testing spatial functionality...'
SELECT ST_AsText(ST_Point(1, 2));
SELECT ST_Distance(
    ST_Point(-71.06, 42.36)::geography,
    ST_Point(-71.07, 42.37)::geography
) as distance_meters;
EOF

# Run verification
psql -h $INSTANCE_IP -U postgres -d twigger_db -f verify_extensions.sql
```

## 5. Cloud SQL Proxy Setup

### Install Cloud SQL Proxy

**Linux/macOS:**
```bash
curl -o cloud-sql-proxy https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.18.2/cloud-sql-proxy.linux.amd64
chmod +x cloud-sql-proxy
sudo mv cloud-sql-proxy /usr/local/bin/
```

**Windows:**
```bash
curl -o cloud-sql-proxy.exe https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.18.2/cloud-sql-proxy.windows.amd64.exe
# Move to a directory in your PATH
```

### Configure and Start Proxy

```bash
# Authenticate with Google Cloud
gcloud auth application-default login

# Start Cloud SQL Proxy
cloud-sql-proxy $PROJECT_ID:$REGION:$INSTANCE_ID \
    --port 5432 \
    --private-ip  # Use this for private IP instances
```

### Proxy as a Service (Linux)

```bash
# Create systemd service file
sudo tee /etc/systemd/system/cloud-sql-proxy.service > /dev/null << EOF
[Unit]
Description=Google Cloud SQL Proxy
After=network.target

[Service]
Type=simple
User=cloudsql
Group=cloudsql
ExecStart=/usr/local/bin/cloud-sql-proxy $PROJECT_ID:$REGION:$INSTANCE_ID --port=5432
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Create user for the service
sudo useradd -r -s /bin/false cloudsql

# Enable and start service
sudo systemctl enable cloud-sql-proxy
sudo systemctl start cloud-sql-proxy
```

## 6. Backup and Recovery Configuration

### Automated Backup Settings

```bash
# Configure backup retention and scheduling
gcloud sql instances patch $INSTANCE_ID \
    --backup-start-time=02:00 \
    --retained-backups-count=14 \
    --retained-transaction-log-days=7 \
    --backup-location=us \
    --project=$PROJECT_ID
```

### Manual Backup Creation

```bash
# Create on-demand backup
gcloud sql backups create \
    --instance=$INSTANCE_ID \
    --description="Manual backup $(date +%Y%m%d_%H%M%S)" \
    --project=$PROJECT_ID

# List available backups
gcloud sql backups list \
    --instance=$INSTANCE_ID \
    --project=$PROJECT_ID
```

### Point-in-Time Recovery

```bash
# Clone instance to specific point in time
gcloud sql instances clone $INSTANCE_ID \
    recovery-instance-$(date +%Y%m%d-%H%M%S) \
    --point-in-time="2025-09-30T08:00:00.000Z" \
    --project=$PROJECT_ID

# Restore from specific backup
BACKUP_ID="1759168800000"  # Get from backup list
gcloud sql instances restore-backup $INSTANCE_ID \
    --backup-id=$BACKUP_ID \
    --restore-instance=restored-instance-$(date +%Y%m%d-%H%M%S) \
    --project=$PROJECT_ID
```

## 7. Performance Optimization

### Connection Pooling Configuration

```bash
# Configure connection limits
gcloud sql instances patch $INSTANCE_ID \
    --database-flags=max_connections=100 \
    --database-flags=shared_preload_libraries="pg_stat_statements,auto_explain" \
    --project=$PROJECT_ID
```

### Enable Query Performance Insights

```bash
# Enable query insights
gcloud sql instances patch $INSTANCE_ID \
    --insights-config-query-insights-enabled \
    --insights-config-record-application-tags \
    --insights-config-record-client-address \
    --project=$PROJECT_ID
```

### PostGIS Performance Tuning

```sql
-- Recommended PostgreSQL settings for PostGIS
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '256MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET effective_io_concurrency = 200;

-- Restart required for some settings
SELECT pg_reload_conf();
```

## 8. Monitoring and Alerting

### Enable Monitoring

```bash
# Configure monitoring
gcloud sql instances patch $INSTANCE_ID \
    --enable-bin-log \
    --project=$PROJECT_ID
```

### Create Alerting Policies

```bash
# CPU utilization alert
gcloud alpha monitoring policies create \
    --policy-from-file=cpu-alert-policy.yaml \
    --project=$PROJECT_ID

# Connection count alert
gcloud alpha monitoring policies create \
    --policy-from-file=connection-alert-policy.yaml \
    --project=$PROJECT_ID
```

**cpu-alert-policy.yaml:**
```yaml
displayName: "Cloud SQL CPU Usage"
conditions:
  - displayName: "CPU usage above 80%"
    conditionThreshold:
      filter: 'resource.type="cloudsql_database" resource.label.database_id="twigger:dev-twigger-db1"'
      comparison: COMPARISON_GREATER_THAN
      thresholdValue: 0.8
      duration: "300s"
      aggregations:
        - alignmentPeriod: "60s"
          perSeriesAligner: ALIGN_MEAN
          crossSeriesReducer: REDUCE_MEAN
          groupByFields: ["resource.label.database_id"]
```

## 9. Security Hardening

### SSL/TLS Configuration

```bash
# Require SSL connections
gcloud sql instances patch $INSTANCE_ID \
    --require-ssl \
    --project=$PROJECT_ID

# Download server certificate
gcloud sql ssl-certs describe server-cert \
    --instance=$INSTANCE_ID \
    --format="value(cert)" > server-ca.pem
```

### Audit Logging

```bash
# Enable audit logging
gcloud sql instances patch $INSTANCE_ID \
    --database-flags=cloudsql.enable_pgaudit=on \
    --database-flags=pgaudit.log=all \
    --project=$PROJECT_ID
```

### IAM Integration

```sql
-- Grant database permissions to IAM user
GRANT CONNECT ON DATABASE twigger_db TO "twigger-db-user@twigger.iam.gserviceaccount.com";
GRANT USAGE ON SCHEMA public TO "twigger-db-user@twigger.iam.gserviceaccount.com";
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO "twigger-db-user@twigger.iam.gserviceaccount.com";
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO "twigger-db-user@twigger.iam.gserviceaccount.com";
```

## 10. Testing and Validation

### Connection Testing

```bash
# Test direct connection
psql -h $INSTANCE_IP -U postgres -d twigger_db -c "SELECT version();"

# Test through Cloud SQL Proxy
psql -h 127.0.0.1 -p 5432 -U postgres -d twigger_db -c "SELECT PostGIS_Version();"

# Test application connection
go run cmd/test-connection/main.go
```

### Schema Deployment

```bash
# Deploy comprehensive schema
go run cmd/apply-schema/main.go

# Run schema tests
go run cmd/test-schema/main.go

# Verify backup functionality
./scripts/verify-backups.sh
```

## 11. Troubleshooting

### Common Issues

**Connection Refused:**
```bash
# Check instance status
gcloud sql instances describe $INSTANCE_ID --project=$PROJECT_ID

# Check authorized networks
gcloud sql instances describe $INSTANCE_ID \
    --format="value(settings.ipConfiguration.authorizedNetworks[].value)" \
    --project=$PROJECT_ID

# Verify IP address
curl -s https://ipinfo.io/ip
```

**PostGIS Extension Issues:**
```sql
-- Check available extensions
SELECT * FROM pg_available_extensions WHERE name LIKE '%postgis%';

-- Check installed extensions
SELECT * FROM pg_extension WHERE extname LIKE '%postgis%';

-- Reinstall if necessary
DROP EXTENSION IF EXISTS postgis CASCADE;
CREATE EXTENSION postgis;
```

**Performance Issues:**
```sql
-- Check query performance
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY total_time DESC
LIMIT 10;

-- Check spatial index usage
SELECT schemaname, tablename, indexname, idx_scan
FROM pg_stat_user_indexes
WHERE indexname LIKE '%gist%'
ORDER BY idx_scan DESC;
```

### Maintenance Tasks

**Weekly:**
- Review backup status
- Check performance metrics
- Validate SSL certificates
- Review audit logs

**Monthly:**
- Update instance to latest patch version
- Review and optimize query performance
- Validate disaster recovery procedures
- Review security settings

## Current Instance Configuration

**Instance Details:**
- **Instance ID:** `dev-twigger-db1`
- **Project:** `twigger`
- **Region:** `us-central1-c`
- **Version:** PostgreSQL 17
- **Public IP:** `162.222.181.26`
- **Private IP:** `10.68.240.3`

**Authorized Networks:**
- `82.217.141.244/32` (Development machine)
- `82.217.141.0/24` (Network range)

**Backup Configuration:**
- **Schedule:** Daily at 02:00 UTC
- **Retention:** 14 days
- **Transaction logs:** 7 days
- **Location:** us region

This setup provides a production-ready Cloud SQL instance with comprehensive PostGIS support for the plant database system.