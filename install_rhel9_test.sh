#!/bin/bash
# ==========================================
# Manual Installation & Test Script for RHEL 9
# Run this script on your local Mac terminal
# ==========================================

echo "=========================================="
echo "RHEL 9 Linux Hardening Client - Manual Install"
echo "=========================================="
echo ""
echo "This script will:"
echo "1. Copy files to RHEL 9 server via SSH"
echo "2. Install the client remotely"
echo "3. Request temporary token from server"
echo "4. Register the client"
echo "5. Test data upload"
echo "6. Verify in database"
echo ""

# Configuration
SERVER_IP="10.60.254.127"
SSH_USER="root"
SSH_PASSWORD="!Qw2!Qw2"
INSTALL_DIR="/opt/linux-hardening-client"
LOCAL_BINARY="/Users/yeung/Projects/system_hardening/bin/linux-hardening-client"
LOCAL_SHELL_SCRIPT="/Users/yeung/Projects/未命名文件夹/RHEL/System_Check-1.2.sh"

# Check prerequisites
echo "[Step 0] Checking prerequisites..."
if ! command -v sshpass &> /dev/null; then
    echo "⚠️  sshpass not found. Install it first:"
    echo "   macOS: brew install sshpass"
    exit 1
fi

if [ ! -f "${LOCAL_BINARY}" ]; then
    echo "❌ Binary file not found at ${LOCAL_BINARY}"
    echo "   Please build the client first:"
    echo "   cd /Users/yeung/Projects/system_hardening/client"
    echo "   GOOS=linux GOARCH=amd64 go build -o ../bin/linux-hardening-client ."
    exit 1
fi

echo "✓ Prerequisites met"
echo ""

# Step 1: Copy files
echo "[Step 1] Copying files to RHEL 9 server..."
echo "   Target: ${SSH_USER}@${SERVER_IP}"
echo ""

sshpass -p "${SSH_PASSWORD}" scp -o StrictHostKeyChecking=no \
    "${LOCAL_BINARY}" \
    "${SSH_USER}@${SERVER_IP}:/tmp/linux-hardening-client"

if [ $? -ne 0 ]; then
    echo "❌ Failed to copy binary file"
    echo ""
    echo "Try running manually:"
    echo "   scp -o StrictHostKeyChecking=no ${LOCAL_BINARY} ${SSH_USER}@${SERVER_IP}:/tmp/"
    exit 1
fi

echo "✓ Binary copied successfully"

sshpass -p "${SSH_PASSWORD}" scp -o StrictHostKeyChecking=no \
    "${LOCAL_SHELL_SCRIPT}" \
    "${SSH_USER}@${SERVER_IP}:/tmp/System_Check-1.2.sh"

if [ $? -ne 0 ]; then
    echo "❌ Failed to copy shell script"
    exit 1
fi

echo "✓ Shell script copied successfully"
echo ""

# Step 2: Install remotely
echo "[Step 2] Installing on remote server..."
echo ""

sshpass -p "${SSH_PASSWORD}" ssh -o StrictHostKeyChecking=no "${SSH_USER}@${SERVER_IP}" << 'REMOTEOF'
#!/bin/bash
set -e

INSTALL_DIR="/opt/linux-hardening-client"

echo "=== Starting Remote Installation ==="
echo "Hostname: $(hostname)"
echo "Date: $(date)"
echo ""

# Create directories
mkdir -p $INSTALL_DIR/{bin,scripts,data,logs}
echo "✓ Created installation directories"

# Move binary
if [ -f /tmp/linux-hardening-client ]; then
    mv /tmp/linux-hardening-client $INSTALL_DIR/bin/
    chmod +x $INSTALL_DIR/bin/linux-hardening-client
    echo "✓ Installed binary"
else
    echo "❌ Binary not found in /tmp/"
    exit 1
fi

# Move shell script
if [ -f /tmp/System_Check-1.2.sh ]; then
    mv /tmp/System_Check-1.2.sh $INSTALL_DIR/scripts/
    chmod +x $INSTALL_DIR/scripts/System_Check-1.2.sh
    echo "✓ Installed shell script"
else
    echo "❌ Shell script not found in /tmp/"
    exit 1
fi

# Get system info
DEVICE_NAME=$(hostname)
IP_ADDRESS=$(ip addr show | grep global | awk '{print $2}' | head -1 || echo "unknown")

# Generate config
cat > $INSTALL_DIR/config.yaml << EOF
server_url: http://localhost:8080
local_db_path: $INSTALL_DIR/data/tokens.db
device_name: $DEVICE_NAME
ip_address: $IP_ADDRESS
script_path: $INSTALL_DIR/scripts/System_Check-1.2.sh
EOF

echo "✓ Generated config"
echo ""

echo "=== Installation Complete ==="
echo ""
echo "Directory structure:"
ls -la $INSTALL_DIR/
echo ""
echo "Configuration:"
cat $INSTALL_DIR/config.yaml
echo ""
echo "Next steps:"
echo "1. Update server_url in config.yaml (change localhost to your Mac's IP)"
echo "2. Request temp token from server API"
echo "3. Register client"
echo "4. Start the client"
REMOTEOF

if [ $? -eq 0 ]; then
    echo "✓ Remote installation completed successfully"
else
    echo "❌ Remote installation failed"
    exit 1
fi

echo ""
echo "[Step 3] Updating configuration..."
echo ""

# Get Mac's actual IP address
MAC_IP=$(ipconfig getifaddr en0)
if [ -z "$MAC_IP" ]; then
    MAC_IP=$(ifconfig en0 | grep "inet " | awk '{print $2}')
fi

echo "Detected Mac IP: ${MAC_IP}"
if [ -z "$MAC_IP" ]; then
    echo "⚠️  Could not detect Mac IP. Please enter manually:"
    read -p "Enter your Mac's IP address: " MAC_IP
fi

# Update config with Mac IP
sshpass -p "${SSH_PASSWORD}" ssh -o StrictHostKeyChecking=no "${SSH_USER}@${SERVER_IP}" << UPDATECONFIG_EOF
sed -i.bak "s|server_url:.*|server_url: http://${MAC_IP}:8080|" $INSTALL_DIR/config.yaml
echo "Updated server_url to: http://${MAC_IP}:8080"
cat $INSTALL_DIR/config.yaml
UPDATECONFIG_EOF

echo "✓ Config updated"
echo ""

# Step 4: Request temp token
echo "[Step 4] Requesting temporary token from backend server..."
echo "   Server URL: http://localhost:8080"
echo ""

TEMP_TOKEN_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/client/request-temp-token" \
    -H "Content-Type: application/json" \
    -d "{
        \"device_name\": \"RHEL9-Server-${SERVER_IP}\",
        \"ip_address\": \"${SERVER_IP}\"
    }")

echo "Response:"
echo "${TEMP_TOKEN_RESPONSE}" | jq .

TEMP_TOKEN=$(echo "${TEMP_TOKEN_RESPONSE}" | jq -r '.temp_token')

if [ "${TEMP_TOKEN}" = "null" ] || [ -z "${TEMP_TOKEN}" ]; then
    echo ""
    echo "❌ Failed to get temporary token"
    echo ""
    echo "Please ensure backend is running:"
    echo "   curl http://localhost:8080/api/health"
    exit 1
fi

echo ""
echo "✓ Temporary token received successfully"
echo ""

# Step 5: Register client
echo "[Step 5] Registering client with backend server..."
echo ""

REGISTER_RESP=$(curl -s -X POST "http://localhost:8080/api/client/register" \
    -H "Content-Type: application/json" \
    -d "{
        \"temp_token\": \"${TEMP_TOKEN}\",
        \"device_name\": \"RHEL9-Server-${SERVER_IP}\",
        \"ip_address\": \"${SERVER_IP}\",
        \"os_version\": \"$(sshpass -p '${SSH_PASSWORD}' ssh -o StrictHostKeyChecking=no ${SSH_USER}@${SERVER_IP} 'cat /etc/redhat-release 2>/dev/null || echo RHEL 9.4')"
    }")

echo "Response:"
echo "${REGISTER_RESP}" | jq .

CLIENT_UUID=$(echo "${REGISTER_RESP}" | jq -r '.client_uuid')
SHORT_TOKEN=$(echo "${REGISTER_RESP}" | jq -r '.short_token')
REFRESH_TOKEN=$(echo "${REGISTER_RESP}" | jq -r '.refresh_token')
EXPIRES_AT=$(echo "${REGISTER_RESP}" | jq -r '.expires_at')

echo ""
echo "✅ Registration successful!"
echo "  Client UUID: ${CLIENT_UUID}"
echo "  Short Token: ${SHORT_TOKEN}"
echo "  Expires At: ${EXPIRES_AT}"
echo ""

# Step 6: Save tokens to SQLite remotely
echo "[Step 6] Saving tokens to local SQLite..."

sshpass -p "${SSH_PASSWORD}" ssh -o StrictHostKeyChecking=no "${SSH_USER}@${SERVER_IP}" << SAVE_TOKENS
sqlite3 $INSTALL_DIR/data/tokens.db << 'SQL_END'
CREATE TABLE IF NOT EXISTS tokens (
    id INTEGER PRIMARY KEY,
    short_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    expires_at TEXT NOT NULL
);
INSERT OR REPLACE INTO tokens (id, short_token, refresh_token, expires_at)
VALUES (1,
    '${SHORT_TOKEN}',
    '${REFRESH_TOKEN}',
    '${EXPIRES_AT}'
);
SQL_END
SAVE_TOKENS

if [ $? -eq 0 ]; then
    echo "✓ Tokens saved to SQLite"
else
    echo "⚠️  Could not save to SQLite (SQLite may not be installed)"
fi
echo ""

# Step 7: Test manual upload
echo "[Step 7] Testing data upload..."
echo ""

TEST_TIMESTAMP=$(date '+%Y/%m/%d_%H:%M:%S')

UPLOAD_RESPONSE=$(sshpass -p "${SSH_PASSWORD}" ssh -o StrictHostKeyChecking=no "${SSH_USER}@${SERVER_IP}" << TEST_UPLOAD_EOF
UPLOAD_RESP=\$(curl -s -X POST "http://${MAC_IP}:8080/api/client/upload-data" \\
  -H "Content-Type: application/json" \\
  -H "X-Client-Token: '${SHORT_TOKEN}'" \\
  -d '{
    "data": {
      "client_uuid": "RHEL9-Server-${SERVER_IP}",
      "date": "'${TEST_TIMESTAMP}'",
      "hostname": "RHEL9-Server",
      "operasystem": "Red Hat Enterprise Linux 9.4",
      "kernel": "5.14.0-xxx.el9.x86_64",
      "ip": "${SERVER_IP}",
      "dnf_conf_gpgcheck": "gpgcheck=1",
      "redhat_repo_gpgcheck": "gpgcheck = 1",
      "pass_max_days": "30",
      "pass_min_days": "1",
      "pass_min_len": "14",
      "pass_warn_age": "7",
      "inactive": "30",
      "gid": "0",
      "tmout": "180",
      "crypto_policies": "DEFAULT:NO-SHA1:NO-WEAKMAC",
      "ntp_server": "server ntp.example.com iburst"
    }
  }')

echo "\$UPLOAD_RESP" | jq .
TEST_UPLOAD_EOF
)

echo "Upload response:"
echo "${UPLOAD_RESPONSE}"

RECORD_ID=$(echo "${UPLOAD_RESPONSE}" | jq -r '.record_id // empty')

if [ -n "${RECORD_ID}" ] && [ "${RECORD_ID}" != "null" ]; then
    echo ""
    echo "✅ Upload test successful! Record ID: ${RECORD_ID}"
else
    echo ""
    echo "⚠️  Upload may have failed or returned unexpected response"
fi
echo ""

# Step 8: Verify in database
echo "[Step 8] Verifying data in MySQL database..."
echo ""

echo "Clients table:"
mysql -u root -p"${SSH_PASSWORD}" -h localhost system_hardening -e "
SELECT 
    id,
    client_uuid,
    device_name,
    ip_address,
    os_version,
    status,
    last_check_time,
    last_upload_time,
    created_at
FROM clients 
WHERE device_name LIKE '%RHEL9%'
ORDER BY id DESC;"

echo ""
echo "SystemCheck table (latest 5 records):"
mysql -u root -p"${SSH_PASSWORD}" -h localhost system_hardening -e "
SELECT 
    id,
    hostname,
    operasystem,
    kernel,
    ip,
    date,
    client_uuid
FROM systemcheck 
WHERE hostname LIKE '%RHEL9%'
ORDER BY id DESC 
LIMIT 5;"

echo ""

# Step 9: Summary
echo "============================================"
echo "Installation & Test Complete!"
echo "============================================"
echo ""
echo "✅ Successfully completed:"
echo "  • Binary installed on RHEL 9"
echo "  • Shell script installed"
echo "  • Configuration updated (server_url: http://${MAC_IP}:8080)"
echo "  • Client registered (UUID: ${CLIENT_UUID})"
echo "  • Token saved to SQLite"
echo "  • Test data uploaded (Record ID: ${RECORD_ID:-'N/A'})"
echo ""
echo "📁 Installed location: root@${SERVER_IP}:${INSTALL_DIR}"
echo "📄 Config file: ${INSTALL_DIR}/config.yaml"
echo "💾 SQLite DB: ${INSTALL_DIR}/data/tokens.db"
echo ""
echo "📊 Database verification:"
CLIENT_COUNT=$(mysql -u root -p"${SSH_PASSWORD}" -h localhost system_hardening -N -e "SELECT COUNT(*) FROM clients WHERE device_name LIKE '%RHEL9%'")
CHECK_COUNT=$(mysql -u root -p"${SSH_PASSWORD}" -h localhost system_hardening -N -e "SELECT COUNT(*) FROM systemcheck WHERE hostname LIKE '%RHEL9%'")
echo "  • Clients: ${CLIENT_COUNT} record(s)"
echo "  • SystemChecks: ${CHECK_COUNT} record(s)"
echo ""
echo "🔧 To monitor client logs on RHEL 9:"
echo "   tail -f ${INSTALL_DIR}/logs/client.log"
echo ""
echo "🚀 To run daily checks manually on RHEL 9:"
echo "   ${INSTALL_DIR}/bin/linux-hardening-client"
echo ""
echo "⏰ Client runs daily at midnight (default schedule)"
echo ""
echo "============================================"
echo "✅ All operations completed successfully!"
echo "============================================"
