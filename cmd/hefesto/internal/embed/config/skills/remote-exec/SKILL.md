---
name: remote-exec
description: Execute commands on remote servers via SSH — DevOps/infrastructure operations
trigger: When SSH, SCP, rsync, VPS, or remote server operations are needed
version: 1.0.0
---

## When to Use

- Run commands on remote servers/VPS
- Deploy or configure remote infrastructure
- Debug issues on production or staging
- Transfer files between local and remote

## Safety Rules (MANDATORY)

| Rule | Action |
|------|--------|
| **Destructive commands** | ALWAYS confirm before: `rm -rf`, `DROP`, `TRUNCATE`, `mkfs`, `dd` |
| **Production changes** | NEVER modify without explicit user approval |
| **Authentication** | Use SSH keys, NEVER passwords in commands |
| **Preview** | Show command BEFORE executing |
| **Idempotency** | Prefer operations safe to run multiple times |

### Destructive Commands Requiring Confirmation

```text
rm -rf, rmdir          # File deletes
DROP, TRUNCATE, DELETE # Database operations
mkfs, fdisk, dd        # Disk operations
iptables -F, ufw disable # Firewall changes
systemctl stop/disable  # Critical services
--force, -f (on prod)   # Force flags
```

## Connection Patterns

### SSH Config (Recommended)

```
# ~/.ssh/config
Host production
    HostName 192.168.1.100
    User admin
    Port 2222
    IdentityFile ~/.ssh/id_rsa
    ServerAliveInterval 60
```

Then: `ssh production`

### Basic Commands

```bash
# Single command
ssh user@host "command"

# Specific key/port
ssh -i ~/.ssh/key -p 2222 user@host "command"

# Multi-step (stop on error)
ssh user@host "cd /app && git pull && npm install && pm2 restart"

# Here-doc for complex scripts
ssh user@host << 'EOF'
  cd /var/www
  git pull origin main
  composer install --no-dev
  php artisan migrate --force
EOF
```

### Connection Testing

```bash
# Test connectivity
ssh -o ConnectTimeout=5 user@host "echo connected"

# Check key auth works
ssh -o BatchMode=yes -o ConnectTimeout=5 user@host "hostname"
```

## File Transfer

### scp (Simple)

```bash
# Local → Remote
scp file.txt user@host:/path/

# Remote → Local
scp user@host:/path/file.txt ./

# Directory recursive
scp -r ./dir user@host:/remote/
```

### rsync (Recommended)

```bash
# Dry-run first!
rsync -avz --dry-run ./src/ user@host:/dest/

# Sync
rsync -avz ./src/ user@host:/dest/

# With excludes
rsync -avz --exclude 'node_modules' --exclude '.git' ./src/ user@host:/dest/

# Delete (CAUTION - confirm first!)
rsync -avz --delete ./src/ user@host:/dest/

# Resume large file
rsync -avz --partial --progress ./large-file user@host:/dest/
```

## Common Operations

### System Health

```bash
ssh user@host "df -h"              # Disk space
ssh user@host "free -h"            # Memory
ssh user@host "uptime && w"        # Load and users
ssh user@host "docker ps"          # Container status
ssh user@host "ss -tlnp"           # Open ports
```

### Service Management

```bash
ssh user@host "systemctl status nginx"
ssh user@host "sudo systemctl restart nginx"
ssh user@host "journalctl -u nginx -n 100 --no-pager"
```

### Log Analysis

```bash
ssh user@host "tail -f /var/log/app.log"
ssh user@host "grep 'ERROR' /var/log/app.log | tail -50"
ssh user@host "journalctl -u myservice --since '1 hour ago'"
```

### Database Backup

```bash
# Create backup
ssh user@host "mysqldump -u user -p database > /tmp/backup_\$(date +%Y%m%d).sql"

# Download
scp user@host:/tmp/backup_*.sql ./backups/
```

## Output Handling

```bash
# Capture to variable
result=$(ssh user@host "command")

# Capture stdout + stderr
ssh user@host "command" 2>&1 | tee output.log

# Parse specific field
ssh user@host "df -h /" | awk 'NR==2 {print $5}'  # Disk usage %

# Parse JSON
ssh user@host "cat /etc/config.json" | jq '.database.host'
```

## Error Handling

```bash
# Check exit code
ssh user@host "command" && echo "Success" || echo "Failed: $?"

# Capture and check
output=$(ssh user@host "command" 2>&1)
if [ $? -ne 0 ]; then
  echo "Error: $output"
fi
```

## Multi-Server Pattern

```bash
# Sequential loop
for server in prod1 prod2 prod3; do
  echo "=== $server ==="
  ssh $server "hostname && uptime"
done

# Parallel (background)
for server in prod1 prod2 prod3; do
  ssh $server "uptime" &
done
wait

# With pssh (if installed)
pssh -h servers.txt "uptime"
```

## Pre-Execution Checklist

- [ ] SSH key configured and tested
- [ ] Command tested on staging first (if production)
- [ ] Destructive commands confirmed by user
- [ ] Output captured for review
- [ ] Rollback plan exists for critical changes
