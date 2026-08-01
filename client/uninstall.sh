#!/bin/bash
set -e

echo "=============================================="
echo "Linux Hardening Client Uninstaller"
echo "=============================================="

INSTALL_DIR="/opt/linux-hardening-client"
SERVER_URL="http://10.60.1.191:8080"

echo ""
echo "Removing client from: $INSTALL_DIR"
echo ""

# Stop and disable service
echo "Step 1: Stopping service..."
systemctl stop linux-hardening-client 2>/dev/null || true
systemctl disable linux-hardening-client 2>/dev/null || true
echo "✓ Service stopped"

# Remove systemd service
echo "Step 2: Removing systemd service..."
rm -f /etc/systemd/system/linux-hardening-client.service
systemctl daemon-reload
echo "✓ Systemd service removed"

# Remove installation directory
echo "Step 3: Removing installation directory..."
rm -rf "$INSTALL_DIR"
echo "✓ Installation directory removed"

# Clean up database records
echo "Step 4: Cleaning up database..."
curl -s -X DELETE "$SERVER_URL/api/client/unregister" \
    -H "Content-Type: application/json" \
    -d '{"server_url": "'$SERVER_URL'"}' 2>/dev/null || echo "⚠️ Database cleanup skipped (might need auth)"

echo ""
echo "=============================================="
echo "Uninstallation Complete!"
echo "=============================================="
echo ""
echo "Client has been completely removed."
echo "You can now reinstall with:"
echo "  bash install_client_interactive.sh $SERVER_URL"
echo ""
