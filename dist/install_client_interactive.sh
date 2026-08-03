#!/bin/bash
# ==========================================
# Interactive Installation Script for Linux Hardening Client
# Automatically detects network configuration and prompts for server info
# Detects system hostname and IP address automatically
# Supports two modes:
#   - Development mode: Runs from project root on macOS
#   - Server mode: Runs after extracting zip package on RHEL9
# ==========================================

set -e

echo ""
echo "=========================================="
echo "  Linux Hardening Client Installation"
echo "=========================================="
echo ""

# Get system information automatically (always from current host)
LOCAL_HOSTNAME=$(hostname)
PRIMARY_IP=$(hostname -I | awk '{print $1}')

# Auto-detect if we're in development or server mode
# Server mode: script is extracted to /tmp, /root/tmp, or any directory containing package files
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "🔍 Detection mode: $SCRIPT_DIR"

# Check if running from a package extraction location (server mode)
if [ -f "${SCRIPT_DIR}/linux-hardening-client" ] && [ -f "${SCRIPT_DIR}/System_Check-1.2.sh" ]; then
    echo "✅ Running in SERVER MODE (extracted from zip package)"
    IN_SERVER_MODE=true
else
    echo "⚠️  Running in DEVELOPMENT MODE (from source code)"
    IN_SERVER_MODE=false
fi

# Check if backend server argument provided, otherwise prompt
if [ -n "$1" ]; then
    SERVER_URL="$1"
else
    echo "Backend Server Configuration:"
    echo "-------------------------------"
    read -p "Enter backend server URL (e.g., http://10.60.1.191:8080): " SERVER_URL
    
    # Validate URL format
    if [[ ! "$SERVER_URL" =~ ^https?:// ]]; then
        echo "❌ Invalid URL format. Please include http:// or https://"
        exit 1
    fi
fi

echo ""
echo "System Information:"
echo "-------------------"
echo "Hostname:   ${LOCAL_HOSTNAME}"
echo "IP Address: ${PRIMARY_IP}"
echo "Server URL: ${SERVER_URL}"
echo ""

# Default installation path
INSTALL_DIR="/opt/linux-hardening-client"
DATA_DIR="${INSTALL_DIR}/data"
LOGS_DIR="${INSTALL_DIR}/logs"
SCRIPTS_DIR="${INSTALL_DIR}/scripts"
BIN_DIR="${INSTALL_DIR}/bin"

# Create directories
echo "Creating installation directories..."
mkdir -p "${DATA_DIR}" "${LOGS_DIR}" "${SCRIPTS_DIR}" "${BIN_DIR}"

# Copy binary - detect based on mode
BINARY_PATH=""
if [ "$IN_SERVER_MODE" = true ]; then
    # Server mode: binary is in same directory as this script (unzipped)
    BINARY_PATH="${SCRIPT_DIR}/linux-hardening-client"
    echo "📦 Binary path: ${BINARY_PATH}"
    if [ ! -f "${BINARY_PATH}" ]; then
        echo "❌ Error: Binary not found at ${BINARY_PATH}"
        echo "   Make sure you extracted the ZIP package first!"
        exit 1
    fi
else
    # Development mode: binary is in ../bin/
    BINARY_PATH="../bin/linux-hardening-client"
    echo "💻 Dev mode binary path: ${BINARY_PATH}"
    if [ ! -f "${BINARY_PATH}" ]; then
        echo "⚠️  Warning: Binary not found at ${BINARY_PATH}"
        echo "   Please run create_package.sh first!"
        exit 1
    fi
fi

cp "${BINARY_PATH}" "${BIN_DIR}/linux-hardening-client"
chmod +x "${BIN_DIR}/linux-hardening-client"
echo "✅ Binary installed to ${BIN_DIR}/linux-hardening-client"

# Copy shell script - detect based on mode
if [ "$IN_SERVER_MODE" = true ]; then
    # Server mode: script is in same directory as this script (unzipped)
    SCRIPT_PATH="${SCRIPT_DIR}/System_Check-1.2.sh"
    echo "📋 Script path: ${SCRIPT_PATH}"
    if [ ! -f "${SCRIPT_PATH}" ]; then
        echo "❌ Error: System_Check-1.2.sh not found at ${SCRIPT_PATH}"
        exit 1
    fi
else
    # Development mode: script is in dist/ directory
    SCRIPT_PATH="${SCRIPT_DIR}/dist/System_Check-1.2.sh"
    if [ ! -f "${SCRIPT_PATH}" ]; then
        echo "❌ Error: System_Check-1.2.sh not found at ${SCRIPT_PATH}"
        exit 1
    fi
fi

cp "${SCRIPT_PATH}" "${SCRIPTS_DIR}/System_Check-1.2.sh"
chmod +x "${SCRIPTS_DIR}/System_Check-1.2.sh"
echo "✅ Shell script installed to ${SCRIPTS_DIR}/System_Check-1.2.sh"

# Install systemd service file
echo ""
echo "Installing systemd service..."

# Detect the absolute path of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CURRENT_BINARY="${BIN_DIR}/linux-hardening-client"

cat > /etc/systemd/system/linux-hardening-client.service << SERVICE_EOF
[Unit]
Description=Linux Hardening Client
After=network.target
Wants=network.target

[Service]
Type=simple
ExecStart=${CURRENT_BINARY}
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal
User=root
Group=root

[Install]
WantedBy=multi-user.target
SERVICE_EOF

systemctl daemon-reload
systemctl enable linux-hardening-client
systemctl restart linux-hardening-client
echo "✅ Systemd service installed and started"

# Configure application settings
echo ""
echo "Configuring client..."
CONFIG_FILE="${INSTALL_DIR}/config.yaml"

# Generate config with auto-detected values
# device_name and ip_address are automatically detected from the current system
cat > "${CONFIG_FILE}" << CONFIG_EOF
server_url: ${SERVER_URL}
local_db_path: ${DATA_DIR}/tokens.json
device_name: ${LOCAL_HOSTNAME}
ip_address: ${PRIMARY_IP}
script_path: ${SCRIPTS_DIR}/System_Check-1.2.sh
CONFIG_EOF

echo "✅ Configuration saved to ${CONFIG_FILE}"
cat "${CONFIG_FILE}"

# Show next steps
echo ""
echo "=========================================="
echo "Installation Complete!"
echo "=========================================="
echo ""
echo "Next Steps:"
echo ""
echo "1. Register the client with the backend:"
echo "   a. Request temporary token:"
echo "      curl -X POST ${SERVER_URL}/api/client/request-temp-token \\ "
echo "        -H \"Content-Type: application/json\" \\ "
echo "        -d \"{\\\"device_name\\\":\\\"${LOCAL_HOSTNAME}\\\",\\\"ip_address\\\":\\\"${PRIMARY_IP}\\\"}\""
echo ""
echo "   b. Use the returned temp_token to register and obtain tokens"
echo "   c. Save tokens to tokens.json file:"
echo "      echo '{\"short_token\": \"YOUR_SHORT_TOKEN\", \"refresh_token\": \"YOUR_REFRESH_TOKEN\", \"expires_at\": \"2026-08-15T00:00:00Z\"}' > ${DATA_DIR}/tokens.json"
echo ""
echo "2. Monitor the service:"
echo "   systemctl status linux-hardening-client"
echo "   journalctl -u linux-hardening-client -f"
echo "   tail -f ${LOGS_DIR}/client.log"
echo ""
echo "3. Stop/start the service:"
echo "   systemctl stop linux-hardening-client"
echo "   systemctl start linux-hardening-client"
echo ""
echo "Configuration File Location: ${CONFIG_FILE}"
echo ""
