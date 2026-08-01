#!/bin/bash
set -e

echo "=============================================="
echo "Linux Hardening Client Uninstaller"
echo "=============================================="

INSTALL_DIR="/opt/linux-hardening-client"

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

echo ""
echo "=============================================="
echo "Uninstallation Complete!"
echo "=============================================="
echo ""
echo "Client has been completely removed."
echo ""
