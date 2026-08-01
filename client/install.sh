#!/bin/bash
set -e

# ================================================
# RHEL 9 Linux Hardening Client Installation Script
# ================================================

SERVER_URL="${1:-http://10.60.254.127:8080}"
INSTALL_DIR="/opt/linux-hardening-client"

echo "=============================================="
echo "Linux Hardening Client Installer for RHEL 9"
echo "=============================================="
echo ""
echo "Server URL: $SERVER_URL"
echo "Install Directory: $INSTALL_DIR"
echo ""

# 获取设备信息
DEVICE_NAME=$(hostname)
IP_ADDRESS=$(ip a | grep global | awk '{print $2}' | head -1)

echo "Device Information:"
echo "  Hostname: $DEVICE_NAME"
echo "  IP Address: $IP_ADDRESS"
echo ""

# 检查依赖
echo "Checking dependencies..."
DEPENDENCIES=("curl" "sqlite3")
for dep in "${DEPENDENCIES[@]}"; do
    if ! command -v $dep &> /dev/null; then
        echo "❌ Required dependency '$dep' is not installed."
        echo "Please install it first: yum install $dep"
        exit 1
    fi
done
echo "✓ All dependencies satisfied"
echo ""

# Step 1: Request temporary token
echo "Step 1: Requesting temporary installation token..."
TEMP_TOKEN_RESP=$(curl -s -X POST "$SERVER_URL/api/client/request-temp-token" \
    -H "Content-Type: application/json" \
    -d "{\"device_name\":\"$DEVICE_NAME\",\"ip_address\":\"$IP_ADDRESS\"}")

if [ $? -ne 0 ]; then
    echo "❌ Failed to request temporary token from server"
    exit 1
fi

TEMP_TOKEN=$(echo $TEMP_TOKEN_RESP | jq -r '.temp_token')
EXPIRES_IN=$(echo $TEMP_TOKEN_RESP | jq -r '.expires_in')

if [ -z "$TEMP_TOKEN" ] || [ "$TEMP_TOKEN" = "null" ]; then
    echo "❌ Invalid temporary token response"
    echo "Response: $TEMP_TOKEN_RESP"
    exit 1
fi

echo "✓ Temporary token obtained (valid for ${EXPIRES_IN}s)"
echo ""

# Step 2: Create directories
echo "Step 2: Creating installation directories..."
mkdir -p "$INSTALL_DIR/{bin,scripts,data,logs}"

chmod 755 "$INSTALL_DIR"
chown root:root "$INSTALL_DIR"

echo "✓ Directories created"
echo ""

# Step 3: Copy files
echo "Step 3: Installing client binary and scripts..."

# Copy binary if exists
if [ -f "../bin/linux-hardening-client" ]; then
    cp "../bin/linux-hardening-client" "$INSTALL_DIR/bin/"
elif [ -f "../../bin/linux-hardening-client" ]; then
    cp "../../bin/linux-hardening-client" "$INSTALL_DIR/bin/"
else
    echo "❌ Client binary not found!"
    echo "Please ensure the client has been built and placed at:"
    echo "  ../bin/linux-hardening-client"
    echo "  OR ../../bin/linux-hardening-client"
    exit 1
fi

chmod +x "$INSTALL_DIR/bin/linux-hardening-client"

# Copy Shell scripts
if [ -f "../../../RHEL/System_Check-1.2.sh" ]; then
    cp "../../../RHEL/System_Check-1.2.sh" "$INSTALL_DIR/scripts/"
    chmod +x "$INSTALL_DIR/scripts/System_Check-1.2.sh"
fi

echo "✓ Files installed"
echo ""

# Step 4: Generate config file
echo "Step 4: Generating configuration file..."
cat > "$INSTALL_DIR/config.yaml" << EOF
server_url: $SERVER_URL
local_db_path: $INSTALL_DIR/data/tokens.db
device_name: $DEVICE_NAME
ip_address: $IP_ADDRESS
script_path: $INSTALL_DIR/scripts/System_Check-1.2.sh
EOF

echo "✓ Configuration file generated"
echo ""

# Step 5: Register client
echo "Step 5: Registering client with server..."
REGISTER_RESP=$(curl -s -X POST "$SERVER_URL/api/client/register" \
    -H "Content-Type: application/json" \
    -d "{
        \"temp_token\": \"$TEMP_TOKEN\",
        \"device_name\": \"$DEVICE_NAME\",
        \"ip_address\": \"$IP_ADDRESS\",
        \"os_version\": \"$(cat /etc/redhat-release)\"
    }")

if echo "$REGISTER_RESP" | jq -e '.short_token' > /dev/null 2>&1; then
    CLIENT_UUID=$(echo $REGISTER_RESP | jq -r '.client_uuid')
    SHORT_TOKEN=$(echo $REGISTER_RESP | jq -r '.short_token')
    REFRESH_TOKEN=$(echo $REGISTER_RESP | jq -r '.refresh_token')
    EXPIRES_AT=$(echo $REGISTER_RESP | jq -r '.expires_at')
    
    echo "✓ Registration successful"
    echo "  Client UUID: $CLIENT_UUID"
    echo "  Token Expires: $EXPIRES_AT"
else
    echo "❌ Registration failed"
    echo "Response: $REGISTER_RESP"
    exit 1
fi
echo ""

# Step 6: Save tokens to SQLite
echo "Step 6: Saving tokens to local database..."
sqlite3 "$INSTALL_DIR/data/tokens.db" << SQL
CREATE TABLE IF NOT EXISTS tokens (
    id INTEGER PRIMARY KEY,
    short_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    expires_at TEXT NOT NULL
);
INSERT OR REPLACE INTO tokens (id, short_token, refresh_token, expires_at)
VALUES (1, '$SHORT_TOKEN', '$REFRESH_TOKEN', '$EXPIRES_AT');
SQL

echo "✓ Tokens saved successfully"
echo ""

# Step 7: Create systemd service
echo "Step 7: Creating systemd service..."
cat > "/etc/systemd/system/linux-hardening-client.service" << 'SERVICE_EOF'
[Unit]
Description=Linux Hardening Client Service
After=network.target multi-user.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/opt/linux-hardening-client/bin/linux-hardening-client
Restart=always
RestartSec=10
StandardOutput=append:/opt/linux-hardening-client/logs/client.log
StandardError=append:/opt/linux-hardening-client/logs/client.log
User=root
Group=root
Environment="PATH=/usr/bin:/bin"

[Install]
WantedBy=multi-user.target
SERVICE_EOF

chmod 644 "/etc/systemd/system/linux-hardening-client.service"

echo "✓ Systemd service created"
echo ""

# Step 8: Enable and start service
echo "Step 8: Starting service..."
systemctl daemon-reload
systemctl enable linux-hardening-client
systemctl start linux-hardening-client

if [ $? -eq 0 ]; then
    echo "✓ Service started successfully"
else
    echo "⚠️  Failed to start service (may be due to missing binary or config)"
fi

echo ""
echo "=============================================="
echo "Installation Complete!"
echo "=============================================="
echo ""
echo "Service Status:"
systemctl status linux-hardening-client --no-pager -l | tail -n 5
echo ""
echo "Important Paths:"
echo "  Configuration: $INSTALL_DIR/config.yaml"
echo "  Log files:     $INSTALL_DIR/logs/"
echo "  Scripts:       $INSTALL_DIR/scripts/"
echo "  Data:          $INSTALL_DIR/data/"
echo ""
echo "Client Info:"
echo "  Device Name:   $DEVICE_NAME"
echo "  IP Address:    $IP_ADDRESS"
echo "  Server:        $SERVER_URL"
echo ""
echo "Quick Commands:"
echo "  View logs:      journalctl -u linux-hardening-client -f"
echo "  Check status:   systemctl status linux-hardening-client"
echo "  Restart:        systemctl restart linux-hardening-client"
echo "  Stop:           systemctl stop linux-hardening-client"
echo ""
echo "⚠️  Security Reminder:"
echo "  - This installation token is used once and will expire in ${EXPIRES_IN} seconds"
echo "  - For reinstallation, please request a new temporary token from your administrator"
echo "  - Keep your server URL secure and never expose it publicly"
echo ""
