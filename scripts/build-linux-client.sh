#!/bin/bash
# Linux 客户端交叉编译与打包脚本
set -e

VERSION="${1:-1.1.0}"
BASE_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CLIENT_DIR="$BASE_DIR/linux-client"
TEMP_DIR="$BASE_DIR/.temp_build_v${VERSION}"
PACKAGE_NAME="linux-hardening-client_v${VERSION}.zip"

echo "======================================"
echo "Linux 加固客户端 v${VERSION} 构建开始"
echo "======================================"

go version || { echo "❌ Go 未安装！"; exit 1; }

echo "[2/4] 编译 Linux 客户端二进制文件..."
mkdir -p "$TEMP_DIR"
cd "$CLIENT_DIR"
GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X main.version=${VERSION}" \
    -o "$TEMP_DIR/linux-hardening-client" .

if [ ! -f "$TEMP_DIR/linux-hardening-client" ]; then
    echo "❌ 编译失败！"; exit 1
fi

chmod +x "$TEMP_DIR/linux-hardening-client"
echo "✅ 二进制文件已编译 (MD5: $(md5sum $TEMP_DIR/linux-hardening-client | awk '{print $1}'))"

echo "[3/4] 准备安装包文件..."
mkdir -p "$TEMP_DIR/linux-hardening-client_v${VERSION}"

# **关键**: 所有文件都从 dist 目录复制（避免 dist 目录下的其他文件混入）
cp "$TEMP_DIR/linux-hardening-client" "$TEMP_DIR/linux-hardening-client_v${VERSION}/"
cp "$BASE_DIR/dist/System_Check-1.2.sh" "$TEMP_DIR/linux-hardening-client_v${VERSION}/"
cp "$BASE_DIR/dist/install_client_interactive.sh" "$TEMP_DIR/linux-hardening-client_v${VERSION}/"
cp "$CLIENT_DIR/uninstall_server.sh" "$TEMP_DIR/linux-hardening-client_v${VERSION}/uninstall.sh"
cp "$BASE_DIR/dist/config.example.yaml" "$TEMP_DIR/linux-hardening-client_v${VERSION}/"
cp "$CLIENT_DIR/README.md" "$TEMP_DIR/linux-hardening-client_v${VERSION}/README_Client.md"
cp "$BASE_DIR/dist/README.md" "$TEMP_DIR/linux-hardening-client_v${VERSION}/README.md"

chmod +x "$TEMP_DIR/linux-hardening-client_v${VERSION}/linux-hardening-client"
chmod +x "$TEMP_DIR/linux-hardening-client_v${VERSION}/System_Check-1.2.sh"
chmod +x "$TEMP_DIR/linux-hardening-client_v${VERSION}/install_client_interactive.sh"

echo "[4/4] 打包为 ZIP (仅包含子目录)..."
cd "$TEMP_DIR"
rm -rf linux-hardening-client
zip -r "$BASE_DIR/$PACKAGE_NAME" linux-hardening-client_v${VERSION}

cd "$BASE_DIR"
rm -rf "$TEMP_DIR"

echo ""
echo "======================================"
echo "✅ 构建完成！"
echo "======================================"
echo "安装包：$PACKAGE_NAME ($(ls -lh $BASE_DIR/$PACKAGE_NAME | awk '{print $5}'))"
echo ""
echo "ZIP 包内容:"
unzip -l $BASE_DIR/$PACKAGE_NAME
echo ""
echo "📦 解压后唯一子目录：linux-hardening-client_v${VERSION}"
echo "📁 没有嵌套或重复的文件"
echo ""
echo "下一步操作: 部署到测试服务器进行端到端测试"
echo "======================================"
