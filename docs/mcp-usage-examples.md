# MCP Server Usage Examples

This document provides comprehensive usage examples for the MCP Server for Product Requirements Management.

## Table of Contents

1. [Basic Setup Examples](#basic-setup-examples)
2. [Configuration Scenarios](#configuration-scenarios)
3. [Integration Examples](#integration-examples)
4. [Workflow Examples](#workflow-examples)
5. [Advanced Use Cases](#advanced-use-cases)
6. [Troubleshooting Scenarios](#troubleshooting-scenarios)

## Basic Setup Examples

### First Time Setup

#### Interactive Setup (Recommended)

```bash
# Step 1: Run interactive setup
./bin/mcp-server -i

# Example interaction:
```
```
🚀 Welcome to MCP Server Setup!

This wizard will help you configure the MCP server for your
Product Requirements Management system.

🌐 Please enter the Backend API URL (e.g., https://api.example.com): 
https://requirements.mycompany.com

🔗 Testing server connectivity...
✅ Server is reachable and ready

🔑 Please enter your username: john.doe
🔒 Please enter your password: [hidden input]

🎟️ Generating Personal Access Token...
✅ Token name: MCP Server - johns-laptop - 2024-01-15
✅ Expires: 2025-01-15

📝 Writing configuration file...
✅ Configuration saved to: /home/john/.requirements-mcp/config.json

🔍 Validating PAT token...
✅ Configuration is valid and ready to use

🎉 Setup completed successfully!

Next steps:
1. Add MCP server to your AI client configuration
2. Start using the server with your AI assistant

For Claude Desktop, add this to ~/.claude/claude_desktop_config.json:
{
  "mcpServers": {
    "requirements-mcp": {
      "command": "/home/john/bin/mcp-server"
    }
  }
}
```

#### Manual Setup

```bash
# Step 1: Create configuration directory
mkdir -p ~/.requirements-mcp

# Step 2: Create configuration file
cat > ~/.requirements-mcp/config.json << EOF
{
  "backend_api_url": "https://requirements.mycompany.com",
  "pat_token": "your_personal_access_token_here",
  "request_timeout": "30s",
  "log_level": "info"
}
EOF

# Step 3: Set secure permissions
chmod 700 ~/.requirements-mcp
chmod 600 ~/.requirements-mcp/config.json

# Step 4: Test configuration
./bin/mcp-server 2>&1 | head -5
```

### Verification Steps

```bash
# Verify configuration loads correctly
./bin/mcp-server 2>&1 | grep "configuration loaded"

# Expected output:
# {"backend_url":"https://requirements.mycompany.com","level":"info","msg":"MCP Server configuration loaded","time":"2024-01-15T10:30:00Z","timeout":"30s"}

# Test basic connectivity
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | ./bin/mcp-server

# Expected response should include:
# {"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-06-18","capabilities":{"resources":{},"tools":{},"prompts":{}},"serverInfo":{"name":"requirements-mcp-server","version":"0.1.0"}}}
```

## Configuration Scenarios

### Development Environment

```bash
# Setup for local development
./bin/mcp-server -i -config ./configs/dev-config.json

# Example configuration for development:
```
```json
{
  "backend_api_url": "http://localhost:8080",
  "pat_token": "dev_token_12345",
  "request_timeout": "10s",
  "log_level": "debug"
}
```

### Testing Environment

```bash
# Setup for testing environment
./bin/mcp-server -i -config ./configs/test-config.json

# Example configuration for testing:
```
```json
{
  "backend_api_url": "https://test-api.mycompany.com",
  "pat_token": "test_token_67890",
  "request_timeout": "30s",
  "log_level": "info"
}
```

### Production Environment

```bash
# Setup for production (with elevated privileges if needed)
sudo ./bin/mcp-server -i -config /etc/requirements-mcp/config.json

# Example configuration for production:
```
```json
{
  "backend_api_url": "https://api.mycompany.com",
  "pat_token": "prod_token_abcdef",
  "request_timeout": "60s",
  "log_level": "warn"
}
```

### Multi-Project Setup

```bash
# Project A configuration
./bin/mcp-server -i -config ~/.requirements-mcp/project-a.json

# Project B configuration  
./bin/mcp-server -i -config ~/.requirements-mcp/project-b.json

# Usage with specific project
./bin/mcp-server -config ~/.requirements-mcp/project-a.json
```

## Integration Examples

### Claude Desktop Integration

#### Basic Integration

```json
# ~/.claude/claude_desktop_config.json
{
  "mcpServers": {
    "requirements-mcp": {
      "command": "/home/user/bin/mcp-server"
    }
  }
}
```

#### Custom Configuration Path

```json
# ~/.claude/claude_desktop_config.json
{
  "mcpServers": {
    "requirements-mcp": {
      "command": "/home/user/bin/mcp-server",
      "args": ["-config", "/home/user/.requirements-mcp/custom-config.json"]
    }
  }
}
```

#### Multiple Projects

```json
# ~/.claude/claude_desktop_config.json
{
  "mcpServers": {
    "requirements-project-a": {
      "command": "/home/user/bin/mcp-server",
      "args": ["-config", "/home/user/.requirements-mcp/project-a.json"]
    },
    "requirements-project-b": {
      "command": "/home/user/bin/mcp-server",
      "args": ["-config", "/home/user/.requirements-mcp/project-b.json"]
    }
  }
}
```

#### Development vs Production

```json
# ~/.claude/claude_desktop_config.json
{
  "mcpServers": {
    "requirements-dev": {
      "command": "/home/user/bin/mcp-server",
      "args": ["-config", "/home/user/.requirements-mcp/dev-config.json"]
    },
    "requirements-prod": {
      "command": "/home/user/bin/mcp-server", 
      "args": ["-config", "/home/user/.requirements-mcp/prod-config.json"]
    }
  }
}
```

### Kiro IDE Integration

```json
# ~/.kiro/settings/mcp.json
{
  "mcpServers": {
    "requirements-mcp": {
      "command": "/home/user/bin/mcp-server",
      "args": ["-config", "/home/user/.requirements-mcp/config.json"],
      "disabled": false,
      "autoApprove": [
        "search_global",
        "search_requirements"
      ]
    }
  }
}
```

### Custom MCP Client Integration

```python
# Example Python MCP client integration
import subprocess
import json

class RequirementsMCPClient:
    def __init__(self, server_path, config_path=None):
        args = [server_path]
        if config_path:
            args.extend(["-config", config_path])
        
        self.process = subprocess.Popen(
            args,
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
    
    def send_request(self, method, params=None):
        request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": method,
            "params": params or {}
        }
        
        self.process.stdin.write(json.dumps(request) + "\n")
        self.process.stdin.flush()
        
        response = self.process.stdout.readline()
        return json.loads(response)

# Usage
client = RequirementsMCPClient("/path/to/bin/mcp-server")
response = client.send_request("initialize", {
    "protocolVersion": "2025-06-18",
    "capabilities": {},
    "clientInfo": {"name": "custom-client", "version": "1.0.0"}
})
```

## Workflow Examples

### Configuration Management Workflow

#### Initial Setup

```bash
# 1. Install MCP server
make build

# 2. Run interactive setup
./bin/mcp-server -i

# 3. Verify setup
./bin/mcp-server 2>&1 | head -5

# 4. Configure AI client (Claude Desktop)
cat >> ~/.claude/claude_desktop_config.json << EOF
{
  "mcpServers": {
    "requirements-mcp": {
      "command": "$(pwd)/bin/mcp-server"
    }
  }
}
EOF
```

#### Configuration Updates

```bash
# 1. Backup existing configuration
cp ~/.requirements-mcp/config.json ~/.requirements-mcp/config.json.backup

# 2. Update configuration
./bin/mcp-server -i

# 3. Test new configuration
./bin/mcp-server 2>&1 | grep "configuration loaded"

# 4. Rollback if needed
mv ~/.requirements-mcp/config.json.backup ~/.requirements-mcp/config.json
```

#### Token Rotation

```bash
# 1. Generate new token through web interface or API
# 2. Update configuration with new token
./bin/mcp-server -i

# 3. Verify new token works
./bin/mcp-server 2>&1 | grep -E "(authentication|token)"

# 4. Remove old token from web interface
```

### Development Workflow

#### Local Development Setup

```bash
# 1. Start local backend API
make dev

# 2. Configure MCP server for local development
./bin/mcp-server -i -config ./dev-config.json
# Enter: http://localhost:8080

# 3. Test integration
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./bin/mcp-server -config ./dev-config.json

# 4. Configure Claude Desktop for development
jq '.mcpServers["requirements-dev"] = {"command": "'$(pwd)'/bin/mcp-server", "args": ["-config", "'$(pwd)'/dev-config.json"]}' ~/.claude/claude_desktop_config.json > /tmp/config.json && mv /tmp/config.json ~/.claude/claude_desktop_config.json
```

#### Testing Workflow

```bash
# 1. Run tests
make test-fast

# 2. Build server
make build

# 3. Test with test environment
./bin/mcp-server -config ./test-config.json 2>&1 | head -10

# 4. Run integration tests
# (assuming test backend is running)
```

### Deployment Workflow

#### Production Deployment

```bash
# 1. Build production binary
make build

# 2. Copy to production server
scp bin/mcp-server user@prod-server:/usr/local/bin/

# 3. Setup configuration on production server
ssh user@prod-server
sudo /usr/local/bin/mcp-server -i -config /etc/requirements-mcp/config.json

# 4. Set up systemd service (optional)
sudo tee /etc/systemd/system/mcp-server.service << EOF
[Unit]
Description=MCP Server for Requirements Management
After=network.target

[Service]
Type=simple
User=mcp-server
ExecStart=/usr/local/bin/mcp-server -config /etc/requirements-mcp/config.json
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable mcp-server
sudo systemctl start mcp-server
```

## Advanced Use Cases

### Load Balancing Setup

```bash
# Setup multiple MCP server instances
./bin/mcp-server -i -config ~/.requirements-mcp/instance-1.json
./bin/mcp-server -i -config ~/.requirements-mcp/instance-2.json

# Configure Claude Desktop with multiple servers
```
```json
{
  "mcpServers": {
    "requirements-primary": {
      "command": "/path/to/bin/mcp-server",
      "args": ["-config", "/home/user/.requirements-mcp/instance-1.json"]
    },
    "requirements-secondary": {
      "command": "/path/to/bin/mcp-server",
      "args": ["-config", "/home/user/.requirements-mcp/instance-2.json"]
    }
  }
}
```

### High Availability Setup

```bash
# Primary server configuration
./bin/mcp-server -i -config ~/.requirements-mcp/primary.json
# URL: https://api-primary.mycompany.com

# Backup server configuration
./bin/mcp-server -i -config ~/.requirements-mcp/backup.json  
# URL: https://api-backup.mycompany.com

# Health check script
cat > check-mcp-health.sh << 'EOF'
#!/bin/bash
PRIMARY_CONFIG="$HOME/.requirements-mcp/primary.json"
BACKUP_CONFIG="$HOME/.requirements-mcp/backup.json"

# Test primary server
if timeout 10 ./bin/mcp-server -config "$PRIMARY_CONFIG" 2>&1 | grep -q "configuration loaded"; then
    echo "Primary server is healthy"
    ln -sf "$PRIMARY_CONFIG" ~/.requirements-mcp/active-config.json
else
    echo "Primary server failed, switching to backup"
    ln -sf "$BACKUP_CONFIG" ~/.requirements-mcp/active-config.json
fi
EOF

chmod +x check-mcp-health.sh
```

### Monitoring and Logging

```bash
# Setup structured logging
./bin/mcp-server 2>&1 | jq -r '.time + " " + .level + " " + .msg' | tee /var/log/mcp-server.log

# Monitor performance
./bin/mcp-server 2>&1 | grep -E "(duration|timeout|error)" | tee /var/log/mcp-performance.log

# Setup log rotation
cat > /etc/logrotate.d/mcp-server << EOF
/var/log/mcp-server.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 mcp-server mcp-server
}
EOF
```

### Security Hardening

```bash
# 1. Secure file permissions
chmod 700 ~/.requirements-mcp
chmod 600 ~/.requirements-mcp/*.json

# 2. Create dedicated user (production)
sudo useradd -r -s /bin/false mcp-server
sudo mkdir -p /etc/requirements-mcp
sudo chown mcp-server:mcp-server /etc/requirements-mcp
sudo chmod 700 /etc/requirements-mcp

# 3. Setup configuration with restricted permissions
sudo -u mcp-server /usr/local/bin/mcp-server -i -config /etc/requirements-mcp/config.json
sudo chmod 600 /etc/requirements-mcp/config.json

# 4. Setup firewall rules (if needed)
sudo ufw allow out 443/tcp comment "HTTPS for MCP server"
sudo ufw allow out 80/tcp comment "HTTP for MCP server"
```

## Troubleshooting Scenarios

### Scenario 1: Fresh Installation Issues

```bash
# Problem: Server won't start after fresh installation
# Solution steps:

# 1. Verify binary
ls -la ./bin/mcp-server
file ./bin/mcp-server

# 2. Check dependencies
ldd ./bin/mcp-server

# 3. Test basic execution
./bin/mcp-server --help

# 4. Create configuration
./bin/mcp-server -i

# 5. Test configuration
./bin/mcp-server 2>&1 | head -5
```

### Scenario 2: Configuration Migration

```bash
# Problem: Migrating from old configuration format
# Solution steps:

# 1. Backup old configuration
cp ~/.requirements-mcp/config.json ~/.requirements-mcp/config.json.old

# 2. Create new configuration
./bin/mcp-server -i

# 3. Compare configurations
diff ~/.requirements-mcp/config.json.old ~/.requirements-mcp/config.json

# 4. Test new configuration
./bin/mcp-server 2>&1 | grep "configuration loaded"
```

### Scenario 3: Network Connectivity Issues

```bash
# Problem: Cannot connect to backend API
# Diagnosis steps:

# 1. Test basic connectivity
ping api.mycompany.com

# 2. Test HTTP connectivity
curl -I https://api.mycompany.com/ready

# 3. Test with authentication
curl -H "Authorization: Bearer YOUR_TOKEN" https://api.mycompany.com/auth/profile

# 4. Check MCP server logs
./bin/mcp-server 2>&1 | grep -E "(error|timeout|connection)"

# 5. Test with increased timeout
jq '.request_timeout = "120s"' ~/.requirements-mcp/config.json > /tmp/config.json && mv /tmp/config.json ~/.requirements-mcp/config.json
```

### Scenario 4: Claude Desktop Integration Issues

```bash
# Problem: MCP server not appearing in Claude Desktop
# Solution steps:

# 1. Verify Claude Desktop configuration
cat ~/.claude/claude_desktop_config.json | jq .

# 2. Test server path
ls -la /path/to/bin/mcp-server

# 3. Test server manually
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"claude","version":"1.0.0"}}}' | /path/to/bin/mcp-server

# 4. Check Claude Desktop logs
tail -f ~/.claude/logs/claude_desktop.log

# 5. Use absolute paths
which mcp-server
realpath ./bin/mcp-server
```

### Scenario 5: Performance Issues

```bash
# Problem: Slow response times
# Diagnosis and solution:

# 1. Monitor response times
time ./bin/mcp-server -config <(echo '{"backend_api_url":"https://api.example.com","pat_token":"test","log_level":"debug"}') < /dev/null

# 2. Test API directly
time curl -H "Authorization: Bearer TOKEN" https://api.example.com/api/v1/epics

# 3. Increase timeouts
jq '.request_timeout = "180s"' ~/.requirements-mcp/config.json > /tmp/config.json && mv /tmp/config.json ~/.requirements-mcp/config.json

# 4. Monitor system resources
top -p $(pgrep mcp-server)
```

These examples cover the most common usage scenarios and should help users get started with the MCP server quickly and troubleshoot common issues effectively.