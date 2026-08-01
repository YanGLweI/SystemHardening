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

