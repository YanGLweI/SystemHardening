#!/bin/bash
# ==========================================
# Create Complete Client Installation Package
# With systemd service file and interactive installer
# Includes all necessary files for real-world deployment
# ==========================================

set -e

OUTPUT_DIR="dist"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
PACKAGE_NAME="linux-hardening-client_${TIMESTAMP}.zip"

echo "=== Creating Client Installation Package ==="
echo ""

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# 1. Binary file
echo "1/7 - Including binary..."
cp /Users/yeung/Projects/system_hardening/bin/linux-hardening-client "${OUTPUT_DIR}/"

# 2. Shell scripts
echo "2/7 - Including shell scripts..."
# Check if source already exists in output directory to avoid overwriting
if [ ! -f "${OUTPUT_DIR}/System_Check-1.2.sh" ]; then
    cp /Users/yeung/Projects/system_hardening/dist/System_Check-1.2.sh "${OUTPUT_DIR}/"
fi

# 3. Systemd service file
echo "3/7 - Including systemd service file..."
cat > "${OUTPUT_DIR}/linux-hardening-client.service" << 'SERVICE_EOF'
[Unit]
Description=Linux Hardening Client
After=network.target
Wants=network.target

[Service]
Type=simple
ExecStart=/opt/linux-hardening-client/bin/linux-hardening-client
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal
User=root
Group=root

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=/opt/linux-hardening-client/logs
ReadWritePaths=/opt/linux-hardening-client/data

[Install]
WantedBy=multi-user.target
SERVICE_EOF

# 4. README and documentation
echo "4/7 - Including README..."
cat > "${OUTPUT_DIR}/README.md" << 'EOF'
# Linux Hardening Client Installation Guide

## Prerequisites
- RHEL 9 or compatible distribution
- curl, bash, sqlite3 (for token management)

## Installation Steps

### Quick Start (Recommended)

1. Copy package to server:
   ```bash
   scp linux-hardening-client_XYZ.zip root@server:/tmp/
   cd /tmp && unzip linux-hardening-client_XYZ.zip
   ```

2. Run interactive installer (auto-detects system info, prompts for server IP):
   ```bash
   bash install_client_interactive.sh
   # Or provide server URL directly:
   bash install_client_interactive.sh http://YOUR_SERVER_IP:8080
   ```

3. Register client with backend and save tokens to SQLite (see below)

4. Verify installation:
   ```bash
   systemctl status linux-hardening-client
   journalctl -u linux-hardening-client -f
   ```

### Full Manual Installation

1. Create directories:
   ```bash
   mkdir -p /opt/linux-hardening-client/{bin,scripts,data,logs}
   ```

2. Extract files:
   ```bash
   mv linux-hardening-client /opt/linux-hardening-client/bin/
   chmod +x /opt/linux-hardening-client/bin/linux-hardening-client
   
   mv System_Check-1.2.sh /opt/linux-hardening-client/scripts/
   chmod +x /opt/linux-hardening-client/scripts/System_Check-1.2.sh
   
   mv linux-hardening-client.service /etc/systemd/system/
   ```

3. Configure application (update values according to your environment):
   ```bash
   cat > /opt/linux-hardening-client/config.yaml << 'CONFIG_EOF'
server_url: http://YOUR_SERVER_IP:8080
local_db_path: /opt/linux-hardening-client/data/tokens.json
device_name: YOUR_HOSTNAME
ip_address: YOUR_IP_ADDRESS
script_path: /opt/linux-hardening-client/scripts/System_Check-1.2.sh
CONFIG_EOF
   ```
   
   **Note:** The interactive installer (`install_client_interactive.sh`) does this automatically!

4. Enable service:
   ```bash
   systemctl daemon-reload
   systemctl enable linux-hardening-client
   systemctl start linux-hardening-client
   ```

5. Register client with backend:

   a. Request temporary token:
      ```bash
      TEMP_TOKEN=$(curl -s POST http://YOUR_SERVER_IP:8080/api/client/request-temp-token \
        -H "Content-Type: application/json" \
        -d "{\"device_name\":\"$(hostname)\",\"ip_address\":\"$(hostname -I | awk '{print $1}')\"}" | jq -r '.temp_token')
      echo "Temp token: $TEMP_TOKEN"
      ```

   b. Register client and obtain short/refresh tokens:
      ```bash
      REGISTER_RESP=$(curl -s POST http://YOUR_SERVER_IP:8080/api/client/register \
        -H "Content-Type: application/json" \
        -d "{\"temp_token\":\"$TEMP_TOKEN\",\"device_name\":\"$(hostname)\",\"ip_address\":\"$(hostname -I | awk '{print $1}')\",os_version:\"Red Hat Enterprise Linux release 9.7 (Plow)\"}")
      
      SHORT_TOKEN=$(echo $REGISTER_RESP | jq -r '.short_token')
      REFRESH_TOKEN=$(echo $REGISTER_RESP | jq -r '.refresh_token')
      echo "Short Token: $SHORT_TOKEN"
      echo "Refresh Token: $REFRESH_TOKEN"
      ```

   c. Save tokens to SQLite database:
      ```bash
      TOKENS_DB=/opt/linux-hardening-client/data/tokens.db
      
      # Create table if it doesn't exist
      sqlite3 $TOKENS_DB "CREATE TABLE IF NOT EXISTS tokens(id INTEGER PRIMARY KEY,short_token TEXT NOT NULL,refresh_token TEXT NOT NULL,expires_at TEXT NOT NULL);"
      
      # Insert tokens (expires at 14 days from now)
      EXPIRES_DATE=$(date -u -d "+14 days" +"%Y-%m-%dT%H:%M:%SZ")
      sqlite3 $TOKENS_DB "INSERT OR REPLACE INTO tokens VALUES(1,'${SHORT_TOKEN}','${REFRESH_TOKEN}','${EXPIRES_DATE}');"
      
      echo "Tokens saved to $TOKENS_DB"
      ```

## Package Contents

- `linux-hardening-client` - Main client binary (Go executable)
- `System_Check-1.2.sh` - Security check script for RedHat systems
- `linux-hardening-client.service` - Systemd service definition
- `README.md` - This documentation
- `config.example.yaml` - Example configuration file
- `install_client_interactive.sh` - Interactive installer script

## Configuration File Format

```yaml
server_url: http://BACKEND_SERVER_IP:PORT
local_db_path: /opt/linux-hardening-client/data/tokens.db
device_name: YOUR_SYSTEM_HOSTNAME
ip_address: YOUR_SYSTEM_IP_ADDRESS
script_path: /opt/linux-hardening-client/scripts/System_Check-1.2.sh
```

**Important:** Use the interactive installer to automatically detect and configure these settings!

## Interactive Installer

The `install_client_interactive.sh` script provides an easy way to install and configure the client:

**Features:**
- Automatically detects hostname and IP address
- Prompts for backend server URL (required for production)
- Creates all necessary directories
- Copies files to correct locations
- Installs and enables systemd service
- Generates configuration file
- Provides next-step instructions

**Usage:**
```bash
# Interactive mode (prompts for server URL)
sudo bash install_client_interactive.sh

# Non-interactive mode (provide server URL as argument)
sudo bash install_client_interactive.sh http://10.60.254.191:8080
```

## Troubleshooting

1. **Token refresh failed**: 
   - Check if backend is running at `server_url`
   - Verify network connectivity: `curl -v http://SERVER_IP:PORT/api/client/request-temp-token`

2. **Script execution failed**: 
   - Ensure `System_Check-1.2.sh` is executable: `chmod +x /opt/linux-hardening-client/scripts/System_Check-1.2.sh`

3. **Cannot connect to server**: 
   - Verify firewall rules allow port access
   - Check DNS resolution if using domain names

4. **Service won't start**: 
   - Check logs: `journalctl -u linux-hardening-client -f`
   - Verify config file: `cat /opt/linux-hardening-client/config.yaml`
   - Test binary manually: `/opt/linux-hardening-client/bin/linux-hardening-client --debug`

5. **Database errors**: 
   - Check SQLite permissions: `ls -la /opt/linux-hardening-client/data/`
   - Verify tokens exist in database: `sqlite3 /opt/linux-hardening-client/data/tokens.db "SELECT * FROM tokens;"`

EOF

# 5. Example config file
echo "5/7 - Including example config..."
cat > "${OUTPUT_DIR}/config.example.yaml" << 'EXAMPLE_EOF'
# Example configuration file
# This file shows the required format and sample values
# IMPORTANT: Use install_client_interactive.sh for automatic configuration!
server_url: http://localhost:8080
local_db_path: /opt/linux-hardening-client/data/tokens.db
device_name: localhost
ip_address: 127.0.0.1
script_path: /opt/linux-hardening-client/scripts/System_Check-1.2.sh
EXAMPLE_EOF

# 6. Interactive installer
echo "6/7 - Including interactive installer..."
cp /Users/yeung/Projects/system_hardening/install_client_interactive.sh "${OUTPUT_DIR}/"

# Create package
echo ""
echo "7/7 - Creating zip package..."
cd "${OUTPUT_DIR}"
zip -r "${PACKAGE_NAME}" linux-hardening-client System_Check-1.2.sh linux-hardening-client.service README.md config.example.yaml install_client_interactive.sh

if [ -f "${OUTPUT_DIR}/${PACKAGE_NAME}" ]; then
    echo ""
    echo "✅ Package created successfully!"
    echo "Location: ${OUTPUT_DIR}/${PACKAGE_NAME}"
    echo "Size: $(ls -lh "${OUTPUT_DIR}/${PACKAGE_NAME}" | awk '{print $5}')"
    echo ""
    echo "Package contents:"
    unzip -l "${OUTPUT_DIR}/${PACKAGE_NAME}" | grep -E "linux-hardening|systemd|install|README|config" || true
    
    echo ""
    echo "Next steps:"
    echo "1. Upload ${PACKAGE_NAME} to RHEL 9 server via SCP"
    echo "2. Unzip on server: unzip ${PACKAGE_NAME}"
    echo "3. Install: sudo bash install_client_interactive.sh"
    echo "4. Follow on-screen instructions (you'll be prompted for server URL)"
    echo "5. Register client with backend and save tokens to SQLite"
else
    echo "❌ Failed to create package"
    exit 1
fi
