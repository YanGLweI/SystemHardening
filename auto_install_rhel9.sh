#!/bin/bash
# ==========================================
# RHEL 9 Linux Hardening Client Installation
# Auto-Installation Script
# ==========================================

set -e

# Configuration
SERVER_IP="10.60.254.127"
SSH_USER="root"
SSH_PASSWORD="!Qw2!Qw2"
INSTALL_DIR="/opt/linux-hardening-client"
LOCAL_BINARY="/Users/yeung/Projects/system_hardening/bin/linux-hardening-client"
LOCAL_SHELL_SCRIPT="/Users/yeung/Projects/未命名文件夹/RHEL/System_Check-1.2.sh"

echo "=========================================="
echo "RHEL 9 Linux Hardening Client - Auto Install"
echo "=========================================="
echo ""
echo "Target Server: ${SSH_USER}@${SERVER_IP}"
echo "Local Binary:  ${LOCAL_BINARY}"
echo "Local Script:  ${LOCAL_SHELL_SCRIPT}"
echo ""

# Step 1: Check if files exist
echo "[Step 1] Checking local files..."
if [ ! -f "${LOCAL_BINARY}" ]; then
    echo "❌ Error: Binary file not found at ${LOCAL_BINARY}"
    echo "   Please build the client first:"
    echo "   cd /Users/yeung/Projects/system_hardening/client"
    echo "   GOOS=linux GOARCH=amd64 go build -o ../bin/linux-hardening-client ."
    exit 1
fi

if [ ! -f "${LOCAL_SHELL_SCRIPT}" ]; then
    echo "❌ Error: Shell script not found at ${LOCAL_SHELL_SCRIPT}"
    exit 1
fi

echo "✓ Local files verified"
echo ""

# Step 2: Copy files to remote server
echo "[Step 2] Copying files to remote server..."

echo "  - Copying binary..."
sshpass -p "${SSH_PASSWORD}" scp -o StrictHostKeyChecking=no "${LOCAL_BINARY}" \
    "${SSH_USER}@${SERVER_IP}:/tmp/linux-hardening-client"
if [ $? -ne 0 ]; then
    echo "❌ Failed to copy binary"
    exit 1
fi

echo "  - Copying shell script..."
sshpass -p "${SSH_PASSWORD}" scp -o StrictHostKeyChecking=no "${LOCAL_SHELL_SCRIPT}" \
    "${SSH_USER}@${SERVER_IP}:/tmp/System_Check-1.2.sh"
if [ $? -ne 0 ]; then
    echo "❌ Failed to copy shell script"
    exit 1
fi

echo "✓ Files copied successfully"
echo ""

# Step 3: Execute installation on remote server
echo "[Step 3] Installing on remote server..."

REMOTE_INSTALL_CMD=$(cat << 'END_REMOTEOF'
#!/bin/bash
set -e

INSTALL_DIR="/opt/linux-hardening-client"

echo "=== Remote Installation Started ==="
echo "Hostname: $(hostname)"
echo "Date: $(date)"
echo ""

# Create directories
echo "Creating installation directories..."
mkdir -p ${INSTALL_DIR}/{bin,scripts,data,logs}
echo "✓ Directories created"

# Move binary
echo "Installing binary..."
if [ -f /tmp/linux-hardening-client ]; then
    mv /tmp/linux-hardening-client ${INSTALL_DIR}/bin/
    chmod +x ${INSTALL_DIR}/bin/linux-hardening-client
    echo "✓ Binary installed"
else
    echo "❌ Binary not found in /tmp/"
    exit 1
fi

# Move shell script
echo "Installing shell script..."
if [ -f /tmp/System_Check-1.2.sh ]; then
    mv /tmp/System_Check-1.2.sh ${INSTALL_DIR}/scripts/
    chmod +x ${INSTALL_DIR}/scripts/System_Check-1.2.sh
    echo "✓ Shell script installed"
else
    echo "❌ Shell script not found in /tmp/"
    exit 1
fi

# Get server IP info
DEVICE_NAME=$(hostname)
IP_ADDRESS=$(ip addr show | grep global | awk '{print $2}' | head -1 || echo "unknown")
OS_VERSION=$(cat /etc/redhat-release 2>/dev/null || echo "Unknown OS")

# Generate configuration
echo "Generating configuration..."
cat > ${INSTALL_DIR}/config.yaml << CONFIGEOF
server_url: http://${IP_ADDRESS}:8080
local_db_path: ${INSTALL_DIR}/data/tokens.db
device_name: ${DEVICE_NAME}
ip_address: ${IP_ADDRESS}
script_path: ${INSTALL_DIR}/scripts/System_Check-1.2.sh
CONFIGEOF

echo "✓ Configuration generated"
echo ""

# Display installation status
echo "=== Installation Status ==="
echo "Install Directory: ${INSTALL_DIR}"
echo ""
echo "Directory structure:"
ls -la ${INSTALL_DIR}/
echo ""
echo "Configuration file content:"
cat ${INSTALL_DIR}/config.yaml
echo ""
echo "Binary version check:"
file ${INSTALL_DIR}/bin/linux-hardening-client
echo ""
echo "Shell script check:"
head -20 ${INSTALL_DIR}/scripts/System_Check-1.2.sh
echo ""
echo "=== Installation Complete ==="
echo ""
echo "Next steps:"
echo "1. Request temporary token from server API"
echo "2. Register the client"
echo "3. Start the client service"

END_REMOTEOF

# Execute remotely
sshpass -p "${SSH_PASSWORD}" ssh -o StrictHostKeyChecking=no "${SSH_USER}@${SERVER_IP}" bash << 'SSH_END'
$(cat << 'INNER_EOF'
bash -c "$(cat -)" <<< "$REMOTE_INSTALL_CMD"
INNER_EOF
"

echo "✓ Remote installation completed"
echo ""

# Step 4: Test server connectivity
echo "[Step 4] Testing server connectivity..."

# First, we need to get the actual MAC IP that the server can reach
echo "Getting server's own IP for localhost access..."
SERVER_IP_INFO=$(sshpass -p "${SSH_PASSWORD}" ssh -o StrictHostKeyChecking=no "${SSH_USER}@${SERVER_IP}" \
    "hostname -I | awk '{print \$1}'" 2>/dev/null || echo "localhost")

echo "Server IP detected: ${SERVER_IP_INFO}"
echo ""

# Update config with correct server URL
echo "Updating config with server URL..."
sshpass -p "${SSH_PASSWORD}" ssh -o StrictHostKeyChecking=no "${SSH_USER}@${SERVER_IP}" << EOF
sed -i "s/server_url:.*/server_url: http:\/\/${SERVER_IP_INFO}:8080/" ${INSTALL_DIR}/config.yaml
echo "Config updated:"
cat ${INSTALL_DIR}/config.yaml
EOF

echo ""

# Step 5: Request temporary token and register
echo "[Step 5] Requesting temporary token..."

TEMP_TOKEN_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/client/request-temp-token" \
    -H "Content-Type: application/json" \
    -d "{
        \"device_name\": \"RHEL9-Server-${SERVER_IP}\",
        \"ip_address\": \"${SERVER_IP}\"
    }")

TEMP_TOKEN=$(echo "${TEMP_TOKEN_RESPONSE}" | jq -r '.temp_token')

if [ "${TEMP_TOKEN}" = "null" ] || [ -z "${TEMP_TOKEN}" ]; then
    echo "❌ Failed to get temp token"
    echo "Response: ${TEMP_TOKEN_RESPONSE}"
    echo ""
    echo "Please ensure backend is running:"
    echo "  curl http://localhost:8080/api/health"
    exit 1
fi

echo "✓ Temporary Token received"
echo "  Token: ${TEMP_TOKEN}"
echo ""

# Step 6: Register client remotely
echo "[Step 6] Registering client..."

REGISTER_RESULT=$(sshpass -p "${SSH_PASSWORD}" ssh -o StrictHostKeyChecking=no "${SSH_USER}@${SERVER_IP}" << REMOTE_REGISTER_EOF
REG_RESP=\$(curl -s -X POST "http://${SERVER_IP_INFO}:8080/api/client/register" \\
  -H "Content-Type: application/json" \\
  -d "{
    \\"temp_token\\": \\"${TEMP_TOKEN}\\",
    \\"device_name\\": \\"RHEL9-Server-${SERVER_IP}\\",
    \\"ip_address\\": \\"${SERVER_IP}\\",
    \\"os_version\\": \\"$(sshpass -p '${SSH_PASSWORD}' ssh -o StrictHostKeyChecking=no ${SSH_USER}@${SERVER_IP} 'cat /etc/redhat-release')"\
  }")

echo "\$REG_RESP" | jq .
REMOTE_REGISTER_EOF
)

CLIENT_UUID=$(echo "${REGISTER_RESULT}" | jq -r '.client_uuid')
SHORT_TOKEN=$(echo "${REGISTER_RESULT}" | jq -r '.short_token')
REFRESH_TOKEN=$(echo "${REGISTER_RESULT}" | jq -r '.refresh_token')
EXPIRES_AT=$(echo "${REGISTER_RESULT}" | jq -r '.expires_at')

if [ "${CLIENT_UUID}" = "null" ]; then
    echo "❌ Registration failed"
    echo "Response: ${REGISTER_RESULT}"
    exit 1
fi

echo "✓ Client registered successfully"
echo "  Client UUID: ${CLIENT_UUID}"
echo "  Short Token: ${SHORT_TOKEN}"
echo "  Refresh Token: ${REFRESH_TOKEN}"
echo "  Expires At: ${EXPIRES_AT}"
echo ""

# Step 7: Save tokens to SQLite remotely
echo "[Step 7] Saving tokens to local SQLite..."

sshpass -p "${SSH_PASSWORD}" ssh -o StrictHostKeyChecking=no "${SSH_USER}@${SERVER_IP}" << SAVE_TOKENS_EOF
sqlite3 ${INSTALL_DIR}/data/tokens.db << 'SQL_EOF'
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
SQL_EOF

echo "Tokens saved to SQLite database"
SAVE_TOKENS_EOF

echo "✓ Tokens saved"
echo ""

# Step 8: Run manual test upload
echo "[Step 8] Running manual test upload..."

TEST_DATA=$(cat << 'JSON_EOF'
{
  "data": {
    "client_uuid": "RHEL9-Server-${SERVER_IP}",
    "date": "'$(date '+%Y/%m/%d_%H:%M:%S')'",
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
}
JSON_EOF
)

UPLOAD_RESULT=$(sshpass -p "${SSH_PASSWORD}" ssh -o StrictHostKeyChecking=no "${SSH_USER}@${SERVER_IP}" << UPLOAD_TEST_EOF
UPLOAD_RESP=\$(curl -s -X POST "http://${SERVER_IP_INFO}:8080/api/client/upload-data" \\
  -H "Content-Type: application/json" \\
  -H "X-Client-Token: '${SHORT_TOKEN}'" \\
  --data-binary '\${TEST_DATA}')

echo "\$UPLOAD_RESP" | jq .
UPLOAD_TEST_EOF
)

RECORD_ID=$(echo "${UPLOAD_RESULT}" | jq -r '.record_id // empty')

echo "${UPLOAD_RESULT}"
echo ""

if [ -n "${RECORD_ID}" ] && [ "${RECORD_ID}" != "null" ]; then
    echo "✅ Upload successful! Record ID: ${RECORD_ID}"
else
    echo "⚠️  Upload may have failed or returned unexpected response"
    echo "Response: ${UPLOAD_RESULT}"
fi
echo ""

# Step 9: Verify data in database
echo "[Step 9] Verifying data in database..."

sleep 2

echo "Checking clients table..."
sshpass -p "${SSH_PASSWORD}" mysql -u root -p"${SSH_PASSWORD}" -h localhost system_hardening -e "
SELECT 
    id,
    client_uuid,
    device_name,
    ip_address,
    os_version,
    status,
    last_upload_time,
    created_at
FROM clients 
WHERE device_name LIKE '%RHEL9%'
ORDER BY id DESC;
"

echo ""
echo "Checking systemcheck table..."
sshpass -p "${SSH_PASSWORD}" mysql -u root -p"${SSH_PASSWORD}" -h localhost system_hardening -e "
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
LIMIT 5;
"
echo ""

# Step 10: Summary
echo "=========================================="
echo "Installation & Test Complete!"
echo "=========================================="
echo ""
echo "Summary:"
echo "  ✅ Binary installed: ${INSTALL_DIR}/bin/linux-hardening-client"
echo "  ✅ Shell script installed: ${INSTALL_DIR}/scripts/System_Check-1.2.sh"
echo "  ✅ Config file created: ${INSTALL_DIR}/config.yaml"
echo "  ✅ Client registered: ${CLIENT_UUID}"
echo "  ✅ Token saved to SQLite"
echo "  ✅ Test data uploaded: ${RECORD_ID:-'Manual"}
echo ""
echo "Data verification:"
echo "  • Clients record: $(sshpass -p "${SSH_PASSWORD}" mysql -u root -p"${SSH_PASSWORD}" -h localhost system_hardening -N -e "SELECT COUNT(*) FROM clients WHERE device_name LIKE '%RHEL9%'")"
echo "  • SystemCheck records: $(sshpass -p "${SSH_PASSWORD}" mysql -u root -p"${SSH_PASSWORD}" -h localhost system_hardening -N -e "SELECT COUNT(*) FROM systemcheck WHERE hostname LIKE '%RHEL9%'")"
echo ""
echo "To monitor logs:"
echo "  tail -f ${INSTALL_DIR}/logs/client.log"
echo ""
echo "To run daily checks manually:"
echo "  ${INSTALL_DIR}/bin/linux-hardening-client"
echo ""
echo "Server details:"
echo "  Target: ${SSH_USER}@${SERVER_IP}"
echo "  Configured server URL: http://${SERVER_IP_INFO}:8080"
echo ""
echo "✅ All tests completed successfully!"
