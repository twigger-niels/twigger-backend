# Cloud SQL Backup and Recovery Guide

## Current Backup Configuration

**Instance:** `dev-twigger-db1`
**Project:** `twigger`

### Backup Settings
- ✅ **Automated Backups:** Enabled
- ✅ **Point-in-Time Recovery:** Enabled
- ✅ **Backup Time:** 02:00 UTC (optimized for low usage)
- ✅ **Retention:** 14 daily backups
- ✅ **Transaction Log Retention:** 7 days
- ✅ **Backup Location:** us region
- ✅ **Backup Tier:** STANDARD

## Backup Operations

### List Available Backups
```bash
gcloud sql backups list --instance=dev-twigger-db1 --project=twigger
```

### Create Manual Backup
```bash
# Create an on-demand backup
gcloud sql backups create --instance=dev-twigger-db1 --project=twigger --description="Manual backup $(date +%Y%m%d_%H%M%S)"
```

### Get Backup Details
```bash
gcloud sql backups describe BACKUP_ID --instance=dev-twigger-db1 --project=twigger
```

## Point-in-Time Recovery

### Check Recovery Window
```bash
# Show the earliest point in time you can recover to
gcloud sql instances describe dev-twigger-db1 --project=twigger --format="value(settings.backupConfiguration.transactionLogRetentionDays)"
```

### Perform Point-in-Time Recovery
```bash
# Clone to a specific point in time (creates new instance)
gcloud sql instances clone dev-twigger-db1 recovery-instance-$(date +%Y%m%d-%H%M%S) \
    --point-in-time="2025-09-30T08:00:00.000Z" \
    --project=twigger
```

## Restore from Backup

### Restore from Specific Backup
```bash
# Get backup ID from list command above
BACKUP_ID="1759168800000"

# Create new instance from backup
gcloud sql instances restore-backup dev-twigger-db1 \
    --backup-id=$BACKUP_ID \
    --restore-instance=restored-instance-$(date +%Y%m%d-%H%M%S) \
    --project=twigger
```

## Emergency Recovery Procedures

### 1. Data Corruption Detected
```bash
# Stop applications immediately
# Assess damage scope
# Determine recovery point needed

# Create recovery instance
gcloud sql instances clone dev-twigger-db1 emergency-recovery-$(date +%Y%m%d-%H%M%S) \
    --point-in-time="YYYY-MM-DDTHH:MM:SS.000Z" \
    --project=twigger

# Test recovery instance
# Update application connection strings
# Promote recovery instance if needed
```

### 2. Instance Deletion/Corruption
```bash
# Use most recent backup
LATEST_BACKUP=$(gcloud sql backups list --instance=dev-twigger-db1 --project=twigger --limit=1 --format="value(id)")

# Restore from backup
gcloud sql instances restore-backup dev-twigger-db1 \
    --backup-id=$LATEST_BACKUP \
    --restore-instance=dev-twigger-db1-restored \
    --project=twigger
```

## Monitoring and Alerts

### Check Backup Status
```bash
# Check last backup status
gcloud sql backups list --instance=dev-twigger-db1 --project=twigger --limit=1
```

### Backup Verification Script
```bash
#!/bin/bash
# Run this daily to verify backups
LATEST_BACKUP_STATUS=$(gcloud sql backups list --instance=dev-twigger-db1 --project=twigger --limit=1 --format="value(status)")

if [ "$LATEST_BACKUP_STATUS" != "SUCCESSFUL" ]; then
    echo "❌ ALERT: Latest backup failed!"
    exit 1
else
    echo "✅ Latest backup successful"
fi
```

## Best Practices

### Backup Strategy
- **Daily Backups:** Automated at 02:00 UTC
- **14-day Retention:** Covers 2 weeks of daily backups
- **7-day Transaction Logs:** Point-in-time recovery within last week
- **Manual Backups:** Before major deployments/migrations

### Recovery Testing
- Test recovery procedures monthly
- Verify backup integrity
- Document recovery times
- Practice emergency scenarios

### Security
- Backup encryption: Enabled by default
- Access control: IAM-based
- Geographic redundancy: Cross-region for production

## Cost Optimization

### Backup Storage Costs
- Standard backup storage: ~$0.08/GB/month
- Transaction logs: Included in backup storage
- Cross-region backup: Additional cost for geographic redundancy

### Recommendations
- Monitor backup storage usage
- Consider lifecycle policies for old backups
- Use point-in-time recovery for recent issues
- Use backups for older recovery points