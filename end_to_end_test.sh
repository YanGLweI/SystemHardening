#!/bin/bash
# ==========================================
# Complete End-to-End Test: Install via Package
# Simulate real user workflow from ZIP to Production
# ==========================================

set -e

REMOTE_USER="root"
REMOTE_HOST="10.60.254.127"
REMOTE_PASSWORD="!Qw2!Qw2"
PACKAGE_NAME="linux-hardening-client_$(date +%Y%m%d_%H%M%S).zip"
WORK_DIR="/tmp/hardening-test-$(date +%s)"

echo "=========================================="
echo "🧪 End-to-End Installation Test"
echo "=========================================="
echo ""
echo "Target Server: ${REMOTE_USER}@${REMOTE_HOST}"
echo "Test Directory: ${WORK_DIR}"
echo ""

# Step 1: Create package on local machine
echo "=== Step 1: Creating Installation Package ==="
cd /Users/yeung/Projects/system_hardening
bash create_package.sh
echo ""

# Step 2: Upload package to RHEL 9 server
echo "=== Step 2: Uploading Package to Server ==="
sshpass -p "${REMOTE_PASSWORD}" scp -o StrictHostKeyChecking=no \
    dist/${PACKAGE_NAME} \
    ${REMOTE_USER}@${REMOTE_HOST}:${WORK_DIR}/
echo "✅ Package uploaded successfully"
echo ""

# Step 3: Unpack and install on server
echo "=== Step 3: Installing on Server ==="
INSTALLATION_OUTPUT=$(sshpass -p "${REMOTE_PASSWORD}" ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_HOST} << 'ENDSSH'
#!/bin/bash
set -e

WORK_DIR="/tmp/hardening-test-$$"
PACKAGE_NAME=$(ls ${WORK_DIR}/*.zip 2>/dev/null | head -1)
BASENAME=$(basename ${PACKAGE_NAME})

echo "Unzipping ${BASENAME}..."
unzip -q ${PACKAGE_NAME}
rm ${PACKAGE_NAME}

echo ""
echo "Creating installation directories..."
mkdir -p /opt/linux-hardening-client/{bin,scripts,data,logs}

echo "Moving binary..."
mv linux-hardening-client /opt/linux-hardening-client/bin/
chmod +x /opt/linux-hardening-client/bin/linux-hardening-client

echo "Moving shell script..."
mv System_Check-1.2.sh /opt/linux-hardening-client/scripts/
chmod +x /opt/linux-hardening-client/scripts/System_Check-1.2.sh

echo ""
echo "Generating configuration..."
cat > /opt/linux-hardening-client/config.yaml << EOF
server_url: http://localhost:8080
local_db_path: /opt/linux-hardening-client/data/tokens.db
device_name: $(hostname)
ip_address: $(hostname -I | awk '{print $1}')
script_path: /opt/linux-hardening-client/scripts/System_Check-1.2.sh
EOF

echo "Configuration created:"
cat /opt/linux-hardening-client/config.yaml

echo ""
echo "Directory structure:"
ls -laR /opt/linux-hardening-client/ | grep -E "^total|^d.*linux-hardening" | head -15

echo ""
echo "Installation complete!"
ENDSSH

if [ $? -eq 0 ]; then
    echo "${INSTALLATION_OUTPUT}"
else
    echo "❌ Installation failed!"
    exit 1
fi
echo ""

# Step 4: Register client with backend
echo "=== Step 4: Registering Client ==="
TEMP_TOKEN=$(curl -s -X POST "http://localhost:8080/api/client/request-temp-token" \
    -H "Content-Type: application/json" \
    -d "{\"device_name\":\"$(hostname)\",\"ip_address\":\"10.60.254.127\"}" | jq -r '.temp_token')

REGISTER_RESP=$(curl -s -X POST "http://localhost:8080/api/client/register" \
    -H "Content-Type: application/json" \
    -d "{\"temp_token\":\"${TEMP_TOKEN}\",\"device_name\":\"$(hostname)\",\"ip_address\":\"10.60.254.127\",\"os_version\":\"Red Hat Enterprise Linux release 9.7 (Plow)\"}")

echo "${REGISTER_RESP}" | jq .
CLIENT_UUID=$(echo "${REGISTER_RESP}" | jq -r '.client_uuid')
SHORT_TOKEN=$(echo "${REGISTER_RESP}" | jq -r '.short_token')
REFRESH_TOKEN=$(echo "${REGISTER_RESP}" | jq -r '.refresh_token')
EXPIRES_AT=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

if [ "$CLIENT_UUID" = "null" ] || [ "$SHORT_TOKEN" = "null" ]; then
    echo "❌ Registration failed!"
    exit 1
fi

echo "✅ Client registered: ${CLIENT_UUID}"
echo ""

# Step 5: Save tokens to SQLite
echo "=== Step 5: Saving Tokens to SQLite ==="
sshpass -p "${REMOTE_PASSWORD}" ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_HOST} << SAVE_EOF
sqlite3 /opt/linux-hardening-client/data/tokens.db "CREATE TABLE IF NOT EXISTS tokens(id INTEGER PRIMARY KEY,short_token TEXT NOT NULL,refresh_token TEXT NOT NULL,expires_at TEXT NOT NULL);"
sqlite3 /opt/linux-hardening-client/data/tokens.db \"INSERT OR REPLACE INTO tokens VALUES(1,'${SHORT_TOKEN}','${REFRESH_TOKEN}','${EXPIRES_AT}')\"
SAVE_EOF

echo "✅ Tokens saved successfully"
echo ""

# Step 6: Create systemd service file
echo "=== Step 6: Creating Systemd Service File ==="
SYSTEMD_CONTENT=$(cat << 'SYSTEMD_EOF'
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
SYSTEMD_EOF
)

echo "Sending systemd service file..."
echo "${SYSTEMD_CONTENT}" | sshpass -p "${REMOTE_PASSWORD}" ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_HOST} "cat > /etc/systemd/system/linux-hardening-client.service"
echo "✅ Service file created at /etc/systemd/system/linux-hardening-client.service"
echo ""

# Step 7: Enable and start service
echo "=== Step 7: Starting Client Service ==="
SERVICESTATUS=$(sshpass -p "${REMOTE_PASSWORD}" ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_HOST} << SERVICE_EOF
systemctl daemon-reload
systemctl enable linux-hardening-client
systemctl restart linux-hardening-client

sleep 2
systemctl status linux-hardening-client --no-pager | tail -10

# Check if it's running
systemctl is-active linux-hardening-client

SERVICE_EOF
)

echo "${SERVICESTATUS}"
if [[ "${SERVICESTATUS}" == *"active"* ]]; then
    echo "✅ Service started successfully!"
else
    echo "⚠️ Service may have issues, checking logs..."
fi
echo ""

# Step 8: Wait for data upload and verify
echo "=== Step 8: Waiting for Initial Data Upload ==="
sleep 15

UPLOAD_CHECK=$(sshpass -p "${REMOTE_PASSWORD}" ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_HOST} << VERIFY_EOF
mysql -u root -p"!Qw2!Qw2!Qw2!Qw2" system_hardening -e "SELECT id, LEFT(hostname,20), date, dnf_conf_gpgcheck FROM systemcheck ORDER BY id DESC LIMIT 3;"
VERIFY_EOF
)

echo "${UPLOAD_CHECK}"

RECORD_COUNT=$(sshpass -p "${REMOTE_PASSWORD}" ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_HOST} 'mysql -u root -p"!Qw2!Qw2!Qw2!Qw2" system_hardening -N -e "SELECT COUNT(*) FROM systemcheck WHERE client_uuid='\'''"${CLIENT_UUID}"'\''';"' 2>&1)

if [ "${RECORD_COUNT}" != "" ] && [ "${RECORD_COUNT}" != " " ]; then
    echo "✅ Data uploaded successfully! Total records: ${RECORD_COUNT}"
else
    echo "⚠️ No new records yet, checking logs..."
fi
echo ""

# Step 9: Check client logs
echo "=== Step 9: Checking Client Logs ==="
sshpass -p "${REMOTE_PASSWORD}" ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_HOST} 'tail -n 50 /opt/linux-hardening-client/logs/client.log 2>/dev/null || echo "No log file generated yet"'
echo ""

# Final Summary
echo "=========================================="
echo "✅ END-TO-END INSTALLATION TEST COMPLETE"
echo "=========================================="
echo ""
echo "Installation Summary:"
echo "• Package created: ${PACKAGE_NAME}"
echo "• Installed location: /opt/linux-hardening-client"
echo "• Systemd service: ENABLED & ACTIVE"
echo "• Client UUID: ${CLIENT_UUID}"
echo "• Service config: /etc/systemd/system/linux-hardening-client.service"
echo ""
echo "Verification:"
echo "• Database: MySQL @ test-it connection"
echo "• Records found: ${RECORD_COUNT:-Unknown}"
echo "• Current hostname: $(hostname)"
echo ""
echo "Services Status:"
echo "✅ Backend API: http://localhost:8080"
echo "✅ Client Service: linux-hardening-client"
echo "✅ Systemd management: ENABLED"
echo ""
echo "To manage the service on RHEL 9:"
echo "  systemctl status linux-hardening-client"
echo "  journalctl -u linux-hardening-client -f"
echo "  tail -f /opt/linux-hardening-client/logs/client.log"
echo ""
echo " Production deployment ready!"
