---
name: remote-exec
description: >
  Execute commands on remote servers via SSH. 
  Trigger: When user needs to run commands on VPS, servers, or remote machines.
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.0"
---

## When to Use

- User asks to run commands on a remote server/VPS
- User mentions SSH, scp, rsync, or remote file operations
- User needs to debug issues on production or staging servers
- User wants to deploy or configure remote infrastructure

## Safety Rules (MANDATORY)

| Rule | Reason |
|------|--------|
| **ALWAYS confirm destructive commands** | `rm -rf`, `DROP DATABASE`, `truncate`, disk operations |
| **NEVER modify production without explicit approval** | Ask user first, show exact command |
| **Use SSH keys, NOT passwords** | Security best practice |
| **Show command before execution** | User must see what will run |
| **Prefer idempotent operations** | Safe to run multiple times |

### Destructive Commands Require Confirmation

These commands MUST be confirmed by user before execution:
- `rm -rf`, `rmdir` (recursive deletes)
- `DROP`, `TRUNCATE`, `DELETE` (database operations)
- `mkfs`, `fdisk`, `dd` (disk operations)
- `iptables -F`, `ufw disable` (firewall changes)
- `systemctl stop/disable` on critical services
- Any command with `--force` or `-f` on production

## SSH Connection Patterns

### Basic SSH

```bash
# Connect to server
ssh user@hostname

# Connect with specific key
ssh -i ~/.ssh/id_rsa user@hostname

# Connect with specific port
ssh -p 2222 user@hostname

# Run single command
ssh user@hostname "command"

# Run multiple commands
ssh user@hostname "cmd1 && cmd2 && cmd3"
```

### SSH Config (Recommended)

Use `~/.ssh/config` for easier connections:

```
Host myserver
    HostName 192.168.1.100
    User admin
    Port 2222
    IdentityFile ~/.ssh/id_rsa
    ServerAliveInterval 60
```

Then simply: `ssh myserver`

## Remote Command Execution

### Execute and Capture Output

```bash
# Capture stdout
ssh user@host "command" > output.txt

# Capture both stdout and stderr
ssh user@host "command" 2>&1 | tee output.log

# Store in variable
result=$(ssh user@host "command")
```

### Multi-Step Sessions

```bash
# Chain commands (stop on error)
ssh user@host "cd /app && git pull && npm install && pm2 restart"

# Chain commands (continue on error)
ssh user@host "cmd1; cmd2; cmd3"

# Use here-doc for complex scripts
ssh user@host << 'EOF'
  cd /var/www
  git pull origin main
  composer install --no-dev
  php artisan migrate --force
  php artisan config:cache
EOF
```

### Parallel Execution on Multiple Servers

```bash
# Using GNU parallel
parallel ssh {} "uptime" ::: server1 server2 server3

# Using shell loop (sequential)
for server in server1 server2 server3; do
  ssh $server "hostname && uptime" &
done
wait

# Using pssh (parallel-ssh)
pssh -h servers.txt "uptime"
```

## File Transfer

### scp (Simple Copy)

```bash
# Local to remote
scp file.txt user@host:/path/to/destination/

# Remote to local
scp user@host:/path/to/file.txt ./local/

# Directory (recursive)
scp -r ./local-dir user@host:/remote/path/

# Using SSH config alias
scp file.txt myserver:/tmp/
```

### rsync (Recommended for Syncing)

```bash
# Sync local to remote (dry-run first!)
rsync -avz --dry-run ./src/ user@host:/dest/

# Actually sync
rsync -avz ./src/ user@host:/dest/

# Sync with delete (CAUTION - confirm first!)
rsync -avz --delete ./src/ user@host:/dest/

# Sync with exclude
rsync -avz --exclude 'node_modules' --exclude '.git' ./src/ user@host:/dest/

# Resume interrupted transfer
rsync -avz --partial --progress ./large-file user@host:/dest/
```

## Reading Remote Files

```bash
# Cat a remote file
ssh user@host "cat /path/to/file"

# Check file exists
ssh user@host "test -f /path/to/file && echo 'exists'"

# Get file size
ssh user@host "stat -c%s /path/to/file"

# Tail logs in real-time
ssh user@host "tail -f /var/log/app.log"

# Grep remote file
ssh user@host "grep 'error' /var/log/app.log | tail -20"
```

## Debugging Remote Issues

### System Health

```bash
# Check disk space
ssh user@host "df -h"

# Check memory
ssh user@host "free -h"

# Check running processes
ssh user@host "ps aux | grep node"

# Check open ports
ssh user@host "ss -tlnp"

# Check system load
ssh user@host "uptime && w"
```

### Log Analysis

```bash
# Recent errors
ssh user@host "journalctl -u myservice -n 100 --no-pager"

# Tail multiple logs
ssh user@host "tail -f /var/log/app.log /var/log/nginx/error.log"

# Search logs for pattern
ssh user@host "grep -r 'ERROR' /var/log/ 2>/dev/null | tail -50"
```

### Network Debugging

```bash
# Test connectivity
ssh user@host "curl -I https://example.com"

# Check DNS
ssh user@host "nslookup example.com"

# Test port connectivity
ssh user@host "nc -zv internal-host 3306"

# Check firewall rules
ssh user@host "iptables -L -n"
```

## Common Operations

### Deployment Pattern

```bash
# Safe deployment sequence
ssh user@host << 'EOF'
  cd /var/www/app
  git fetch origin
  git diff --stat HEAD origin/main  # Show changes
EOF

# Ask user confirmation before continuing
# Then:
ssh user@host << 'EOF'
  cd /var/www/app
  git pull origin main
  npm ci
  npm run build
  pm2 reload app
  pm2 logs app --lines 20
EOF
```

### Database Backup

```bash
# Create backup
ssh user@host "mysqldump -u user -p database > /tmp/backup_$(date +%Y%m%d).sql"

# Download backup
scp user@host:/tmp/backup_*.sql ./backups/
```

### Service Management

```bash
# Check service status
ssh user@host "systemctl status nginx"

# Restart service
ssh user@host "sudo systemctl restart nginx"

# View service logs
ssh user@host "journalctl -u nginx -f"
```

## Output Parsing Patterns

```bash
# Extract specific field
ssh user@host "df -h /" | awk 'NR==2 {print $5}'  # Get use percentage

# Parse JSON
ssh user@host "cat /etc/config.json" | jq '.database.host'

# Extract IPs from log
ssh user@host "grep 'login' /var/log/auth.log" | grep -oE '[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' | sort | uniq -c
```

## Checklist Before Remote Execution

- [ ] SSH key is configured and working
- [ ] Command tested in dry-run or on staging first
- [ ] Destructive commands have user confirmation
- [ ] Production changes have explicit user approval
- [ ] Output is captured for review
- [ ] Rollback plan exists for critical changes
