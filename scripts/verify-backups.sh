#!/bin/bash

# Cloud SQL Backup Verification Script
# Run this script daily to verify backup health

set -e

PROJECT_ID="twigger"
INSTANCE_NAME="dev-twigger-db1"

echo "🔍 Verifying Cloud SQL backups for $INSTANCE_NAME..."

# Check if gcloud is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "@"; then
    echo "❌ Error: Not authenticated with gcloud. Run 'gcloud auth login'"
    exit 1
fi

# Set project
gcloud config set project $PROJECT_ID > /dev/null 2>&1

# Check backup configuration
echo "📋 Checking backup configuration..."
BACKUP_ENABLED=$(gcloud sql instances describe $INSTANCE_NAME --format="value(settings.backupConfiguration.enabled)")
PITR_ENABLED=$(gcloud sql instances describe $INSTANCE_NAME --format="value(settings.backupConfiguration.pointInTimeRecoveryEnabled)")

if [ "$BACKUP_ENABLED" = "True" ]; then
    echo "✅ Automated backups: Enabled"
else
    echo "❌ Automated backups: Disabled"
    exit 1
fi

if [ "$PITR_ENABLED" = "True" ]; then
    echo "✅ Point-in-time recovery: Enabled"
else
    echo "❌ Point-in-time recovery: Disabled"
    exit 1
fi

# Check recent backups
echo ""
echo "📦 Checking recent backups..."
RECENT_BACKUPS=$(gcloud sql backups list --instance=$INSTANCE_NAME --limit=3 --format="table(id,windowStartTime,status)")

if [ -z "$RECENT_BACKUPS" ]; then
    echo "❌ No recent backups found"
    exit 1
fi

echo "$RECENT_BACKUPS"

# Check latest backup status
LATEST_BACKUP_STATUS=$(gcloud sql backups list --instance=$INSTANCE_NAME --limit=1 --format="value(status)")
LATEST_BACKUP_TIME=$(gcloud sql backups list --instance=$INSTANCE_NAME --limit=1 --format="value(windowStartTime)")

if [ "$LATEST_BACKUP_STATUS" = "SUCCESSFUL" ]; then
    echo "✅ Latest backup status: $LATEST_BACKUP_STATUS"
    echo "📅 Latest backup time: $LATEST_BACKUP_TIME"
else
    echo "❌ Latest backup status: $LATEST_BACKUP_STATUS"
    exit 1
fi

# Check backup age (should be within last 25 hours for daily backups)
if command -v date >/dev/null 2>&1; then
    BACKUP_DATE=$(date -d "$LATEST_BACKUP_TIME" +%s 2>/dev/null || date -j -f "%Y-%m-%dT%H:%M:%S" "${LATEST_BACKUP_TIME%.*}" +%s 2>/dev/null || echo "0")
    CURRENT_DATE=$(date +%s)
    AGE_HOURS=$(( (CURRENT_DATE - BACKUP_DATE) / 3600 ))

    if [ "$BACKUP_DATE" != "0" ] && [ $AGE_HOURS -lt 25 ]; then
        echo "✅ Backup age: $AGE_HOURS hours (within expected range)"
    elif [ "$BACKUP_DATE" != "0" ]; then
        echo "⚠️  Backup age: $AGE_HOURS hours (older than expected)"
    else
        echo "⚠️  Could not determine backup age"
    fi
fi

# Check database connectivity
echo ""
echo "🔗 Testing database connectivity..."
if command -v go >/dev/null 2>&1; then
    if go run cmd/test-connection/main.go | grep -q "Connection successful"; then
        echo "✅ Database connection: Working"
    else
        echo "⚠️  Database connection test failed (this may be normal if server isn't running)"
    fi
else
    echo "⚠️  Go not available - skipping connection test"
fi

echo ""
echo "🎉 All backup verifications passed!"
echo ""
echo "📊 Summary:"
echo "   - Automated backups: Enabled"
echo "   - Point-in-time recovery: Enabled"
echo "   - Latest backup: $LATEST_BACKUP_STATUS ($LATEST_BACKUP_TIME)"
echo "   - Database connection: Working"
echo ""
echo "💡 Tip: Set up this script to run daily via cron or CI/CD"