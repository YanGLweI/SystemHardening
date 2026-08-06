#!/bin/bash
set -e

echo "=============================================="
echo "  Linux 加固客户端卸载脚本"
echo "=============================================="
echo ""

# 从 systemd 服务文件中自动读取安装路径
SERVICE_FILE="/etc/systemd/system/linux-hardening-client.service"
if [ -f "$SERVICE_FILE" ]; then
    BINARY_PATH=$(grep '^ExecStart=' "$SERVICE_FILE" | sed 's/^ExecStart=//')
    INSTALL_DIR=$(dirname "$(dirname "$BINARY_PATH")")
else
    # 服务文件不存在，使用默认路径
    INSTALL_DIR="/opt/linux-hardening-client"
fi

echo "安装目录: $INSTALL_DIR"
echo ""

# Step 1: 停止并禁用服务
echo "Step 1: 停止服务..."
systemctl stop linux-hardening-client 2>/dev/null || true
systemctl disable linux-hardening-client 2>/dev/null || true
echo "✓ 服务已停止"

# Step 2: 移除 systemd 服务文件
echo "Step 2: 移除 systemd 服务..."
rm -f "$SERVICE_FILE"
systemctl daemon-reload
echo "✓ systemd 服务已移除"

# Step 3: 删除安装目录
echo "Step 3: 删除安装目录..."
rm -rf "$INSTALL_DIR"
echo "✓ 安装目录已删除"

echo ""
echo "=============================================="
echo "  卸载完成！"
echo "=============================================="
echo ""
echo "客户端已完全移除。"
echo ""
