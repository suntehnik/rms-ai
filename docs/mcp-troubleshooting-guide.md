# MCP Server Troubleshooting Guide

This guide provides comprehensive troubleshooting information for the MCP Server for Product Requirements Management.

## Table of Contents

1. [Quick Diagnostics](#quick-diagnostics)
2. [Configuration Issues](#configuration-issues)
3. [Connection Problems](#connection-problems)
4. [Authentication Errors](#authentication-errors)
5. [Initialization Failures](#initialization-failures)
6. [Runtime Errors](#runtime-errors)
7. [Performance Issues](#performance-issues)
8. [Integration Problems](#integration-problems)
9. [Advanced Debugging](#advanced-debugging)

## Quick Diagnostics

### Health Check Commands

```bash
# Check if MCP server binary exists and is executable
ls -la ./bin/mcp-server
file ./bin/mcp-server

# Test configuration loading
./bin/mcp-server 2>&1 | head -5

# Verify configuration file
cat ~/.requirements-mcp/config.json | jq .

# Test backend API connectivity
curl -I https://your-api-server.com/ready
```

### Common Quick Fixes

```bash
# Recreate configuration from scratch
rm -f ~/.requirements-mcp/config.json
./bin/mcp-server -i

# Reset permissions
chmod 700 ~/.requirements-mcp/
chmod 600 ~/.requirements-mcp/config.json

# Clear any cached data
rm -rf ~/.requirements-mcp/.cache/
```

## Configuration Issues

### Problem: Configuration File Not Found

**Error Messages:**
```
Failed to load configuration from ~/.requirements-mcp/config.json: failed to read config file
```

**Diagnosis:**
```bash
# Check if config file exists
ls -la ~/.requirements-mcp/config.json

# Check directory permissions
ls -ld ~/.requirements-mcp/
```

**Solutions:**

1. **Create configuration using interactive mode:**
   ```bash
   ./bin/mcp-server -i
   ```

2. **Manual configuration creation:**
   ```bash
   mkdir -p ~/.requirements-mcp
   cp config.example.json ~/.requirements-mcp/config.json
   nano ~/.requirements-mcp/config.json
   ```

3. **Use custom config path:**
   ```bash
   ./bin/mcp-server -config /path/to/your/config.json
   ```

### Problem: Invalid JSON Format

**Error Messages:**
```
Failed to load configuration: failed to parse config file
invalid character '}' looking for beginning of object key string
```

**Diagnosis:**
```bash
# Validate JSON syntax
jq . ~/.requirements-mcp/config.json

# Check for common issues
cat ~/.requirements-mcp/config.json | grep -E "(,$|^[[:space:]]*$)"
```

**Solutions:**

1. **Use JSON validator:**
   ```bash
   python -m json.tool ~/.requirements-mcp/config.json
   ```

2. **Recreate configuration:**
   ```bash
   mv ~/.requirements-mcp/config.json ~/.requirements-mcp/config.json.backup
   ./bin/mcp-server -i
   ```

3. **Fix common JSON errors:**
   - Remove trailing commas
   - Ensure all strings are quoted
   - Check bracket matching

### Problem: Missing Required Fields

**Error Messages:**
```
Failed to load configuration: invalid configuration: backend_api_url is required
Failed to load configuration: invalid configuration: pat_token is required
```

**Solutions:**

1. **Check required fields:**
   ```bash
   jq 'keys' ~/.requirements-mcp/config.json
   # Should include: backend_api_url, pat_token
   ```

2. **Add missing fields:**
   ```json
   {
     "backend_api_url": "https://your-api-server.com",
     "pat_token": "your_personal_access_token",
     "request_timeout": "30s",
     "log_level": "info"
   }
   ```

## Connection Problems

### Problem: Backend API Unreachable

**Error Messages:**
```
Server error: failed to connect to backend API
dial tcp: lookup api.example.com: no such host
connection refused
```

**Diagnosis:**
```bash
# Test basic connectivity
ping api.example.com

# Test HTTP connectivity
curl -I https://api.example.com/ready

# Check DNS resolution
nslookup api.example.com

# Test with different protocols
curl -I http://api.example.com/ready
curl -I https://api.example.com/ready
```

**Solutions:**

1. **Verify server URL:**
   - Ensure URL includes protocol (http:// or https://)
   - Check for typos in domain name
   - Verify port number if non-standard

2. **Network troubleshooting:**
   ```bash
   # Check network connectivity
   ping 8.8.8.8
   
   # Check firewall rules
   sudo iptables -L
   
   # Test from different network
   curl -I https://api.example.com/ready --interface eth1
   ```

3. **Proxy configuration:**
   ```bash
   # Set proxy if needed
   export HTTP_PROXY=http://proxy.company.com:8080
   export HTTPS_PROXY=http://proxy.company.com:8080
   ./bin/mcp-server
   ```

### Problem: SSL/TLS Certificate Issues

**Error Messages:**
```
x509: certificate signed by unknown authority
x509: certificate has expired
```

**Solutions:**

1. **Update CA certificates:**
   ```bash
   # Ubuntu/Debian
   sudo apt-get update && sudo apt-get install ca-certificates
   
   # CentOS/RHEL
   sudo yum update ca-certificates
   ```

2. **Test certificate:**
   ```bash
   openssl s_client -connect api.example.com:443 -servername api.example.com
   ```

3. **Temporary workaround (not recommended for production):**
   ```bash
   # Only for testing with self-signed certificates
   export GODEBUG=x509ignoreCN=0
   ./bin/mcp-server
   ```

## Authentication Errors

### Problem: Invalid PAT Token

**Error Messages:**
```
Server error: authentication failed: invalid PAT token
HTTP 401 Unauthorized
```

**Diagnosis:**
```bash
# Test token manually
curl -H "Authorization: Bearer YOUR_TOKEN" \
     https://api.example.com/auth/profile

# Check token in config
jq -r '.pat_token' ~/.requirements-mcp/config.json
```

**Solutions:**

1. **Generate new PAT token:**
   ```bash
   ./bin/mcp-server -i
   # This will create a new token automatically
   ```

2. **Manual token creation:**
   - Log into web interface
   - Go to Profile Settings → Personal Access Tokens
   - Create new token with name "MCP Server"
   - Copy token to configuration

3. **Verify token permissions:**
   - Ensure token has required scopes
   - Check token expiration date
   - Verify user account is active

### Problem: Token Expired

**Error Messages:**
```
Server error: authentication failed: token expired
HTTP 401 Unauthorized: Token has expired
```

**Solutions:**

1. **Automatic token renewal:**
   ```bash
   ./bin/mcp-server -i
   # Will create new token with 1-year expiration
   ```

2. **Check token expiration:**
   ```bash
   # Decode JWT token (if applicable)
   echo "YOUR_TOKEN" | base64 -d | jq .exp
   ```

## Initialization Failures

### Problem: Interactive Setup Fails

**Error Messages:**
```
Initialization failed: Network Error: Failed to connect to server
Initialization failed: Authentication Error: Invalid credentials
```

**Diagnosis Steps:**

1. **Test server connectivity:**
   ```bash
   curl -I https://your-server.com/ready
   ```

2. **Verify credentials:**
   - Try logging in through web interface
   - Check username/password for special characters
   - Verify account is not locked

3. **Check server endpoints:**
   ```bash
   curl -X POST https://your-server.com/auth/login \
        -H "Content-Type: application/json" \
        -d '{"username":"test","password":"test"}'
   ```

**Solutions:**

1. **Manual configuration:**
   ```bash
   # Skip interactive setup
   cp config.example.json ~/.requirements-mcp/config.json
   # Edit manually with known good values
   ```

2. **Debug mode:**
   ```bash
   # Run with debug logging
   LOG_LEVEL=debug ./bin/mcp-server -i
   ```

### Problem: File Permission Errors

**Error Messages:**
```
Initialization failed: File System Error: permission denied
mkdir ~/.requirements-mcp: permission denied
```

**Solutions:**

1. **Fix permissions:**
   ```bash
   # Ensure home directory is writable
   ls -ld ~/
   
   # Create directory manually
   mkdir -p ~/.requirements-mcp
   chmod 700 ~/.requirements-mcp
   ```

2. **Use alternative location:**
   ```bash
   ./bin/mcp-server -i -config /tmp/mcp-config.json
   ```

3. **Run with appropriate permissions:**
   ```bash
   # If necessary (not recommended)
   sudo ./bin/mcp-server -i -config /etc/mcp-server/config.json
   ```

## Runtime Errors

### Problem: JSON-RPC Protocol Errors

**Error Messages:**
```
Invalid JSON-RPC request
Method not found
Parse error
```

**Diagnosis:**
```bash
# Check MCP client configuration
cat ~/.claude/claude_desktop_config.json | jq .

# Test with minimal client
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | ./bin/mcp-server
```

**Solutions:**

1. **Verify MCP client configuration:**
   ```json
   {
     "mcpServers": {
       "requirements-mcp": {
         "command": "/correct/path/to/bin/mcp-server",
         "args": ["-config", "/correct/path/to/config.json"]
       }
     }
   }
   ```

2. **Check protocol version compatibility:**
   - Ensure client supports MCP version 2025-06-18
   - Update client if necessary

### Problem: Backend API Errors

**Error Messages:**
```
HTTP 500 Internal Server Error
HTTP 404 Not Found
Request timeout
```

**Solutions:**

1. **Check backend API status:**
   ```bash
   curl https://api.example.com/health
   curl https://api.example.com/api/v1/mcp
   ```

2. **Increase timeout:**
   ```json
   {
     "request_timeout": "60s"
   }
   ```

3. **Check API version compatibility:**
   - Verify MCP endpoint exists: `/api/v1/mcp`
   - Check API documentation for changes

## Performance Issues

### Problem: Slow Response Times

**Diagnosis:**
```bash
# Monitor response times
./bin/mcp-server 2>&1 | grep -E "(duration|timeout)"

# Test API directly
time curl -H "Authorization: Bearer TOKEN" \
     https://api.example.com/api/v1/epics
```

**Solutions:**

1. **Increase timeouts:**
   ```json
   {
     "request_timeout": "120s"
   }
   ```

2. **Check network latency:**
   ```bash
   ping api.example.com
   traceroute api.example.com
   ```

3. **Monitor backend performance:**
   - Check backend server resources
   - Review database performance
   - Check for API rate limiting

### Problem: Memory Usage

**Diagnosis:**
```bash
# Monitor memory usage
ps aux | grep mcp-server
top -p $(pgrep mcp-server)
```

**Solutions:**

1. **Restart server periodically:**
   ```bash
   # Add to cron for daily restart
   0 2 * * * pkill mcp-server && sleep 5 && /path/to/bin/mcp-server
   ```

2. **Check for memory leaks:**
   - Update to latest version
   - Report issue with memory usage patterns

## Integration Problems

### Problem: Claude Desktop Integration

**Common Issues:**

1. **Server not appearing in Claude:**
   ```bash
   # Check Claude Desktop config
   cat ~/.claude/claude_desktop_config.json | jq .
   
   # Verify server path
   ls -la /path/to/bin/mcp-server
   
   # Test server manually
   echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"claude","version":"1.0.0"}}}' | /path/to/bin/mcp-server
   ```

2. **Server crashes on startup:**
   ```bash
   # Check Claude Desktop logs
   tail -f ~/.claude/logs/claude_desktop.log
   
   # Test server independently
   ./bin/mcp-server 2>&1 | head -20
   ```

**Solutions:**

1. **Verify configuration format:**
   ```json
   {
     "mcpServers": {
       "requirements-mcp": {
         "command": "/absolute/path/to/bin/mcp-server",
         "args": ["-config", "/absolute/path/to/config.json"]
       }
     }
   }
   ```

2. **Use absolute paths:**
   ```bash
   # Find absolute path
   which mcp-server
   realpath ./bin/mcp-server
   ```

3. **Check permissions:**
   ```bash
   chmod +x /path/to/bin/mcp-server
   ```

## Advanced Debugging

### Enable Debug Logging

```bash
# Method 1: Environment variable
LOG_LEVEL=debug ./bin/mcp-server

# Method 2: Configuration file
jq '.log_level = "debug"' ~/.requirements-mcp/config.json > /tmp/config.json && mv /tmp/config.json ~/.requirements-mcp/config.json

# Method 3: Temporary config
./bin/mcp-server -config <(echo '{"backend_api_url":"https://api.example.com","pat_token":"YOUR_TOKEN","log_level":"debug"}')
```

### Network Debugging

```bash
# Capture network traffic
sudo tcpdump -i any -w mcp-traffic.pcap host api.example.com

# Monitor HTTP requests
./bin/mcp-server 2>&1 | grep -E "(POST|GET|PUT|DELETE)"

# Test with curl
curl -v -H "Authorization: Bearer TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"method":"tools/list"}' \
     https://api.example.com/api/v1/mcp
```

### Process Debugging

```bash
# Monitor system calls
strace -e trace=network,file ./bin/mcp-server

# Monitor file access
lsof -p $(pgrep mcp-server)

# Check process status
ps aux | grep mcp-server
pstree -p $(pgrep mcp-server)
```

### Log Analysis

```bash
# Parse JSON logs
./bin/mcp-server 2>&1 | jq -r '.time + " " + .level + " " + .msg'

# Filter by log level
./bin/mcp-server 2>&1 | jq 'select(.level == "error")'

# Monitor specific operations
./bin/mcp-server 2>&1 | jq 'select(.msg | contains("authentication"))'

# Count error types
./bin/mcp-server 2>&1 | jq -r 'select(.level == "error") | .msg' | sort | uniq -c
```

## Getting Help

### Information to Collect

When reporting issues, please include:

1. **Version information:**
   ```bash
   ./bin/mcp-server --version
   go version
   uname -a
   ```

2. **Configuration (sanitized):**
   ```bash
   jq 'del(.pat_token)' ~/.requirements-mcp/config.json
   ```

3. **Error logs:**
   ```bash
   ./bin/mcp-server 2>&1 | head -50
   ```

4. **Network test results:**
   ```bash
   curl -I https://your-api-server.com/ready
   ```

### Support Channels

- **Documentation**: Check README-mcp-server.md
- **Issue Tracker**: Create detailed bug reports
- **Debug Mode**: Always include debug logs with issues

### Self-Help Checklist

Before seeking help, try:

- [ ] Recreate configuration with `./bin/mcp-server -i`
- [ ] Test with debug logging enabled
- [ ] Verify backend API is accessible
- [ ] Check file permissions
- [ ] Try with minimal configuration
- [ ] Test network connectivity
- [ ] Review recent changes to environment