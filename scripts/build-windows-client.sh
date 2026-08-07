#!/bin/bash
# ============================================================================
# Windows 客户端交叉编译与 NSIS 打包脚本
# ============================================================================

set -e

VERSION="${1:-1.1.0}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
WINDOWS_CLIENT_DIR="$PROJECT_ROOT/windows-client"
DIST_WIN_DIR="$PROJECT_ROOT/dist_win"

echo "============================================================"
echo "Windows 客户端 v${VERSION} 编译与 NSIS 打包"
echo "============================================================"
echo ""

# 检查工具
echo "[1/4] 检查工具..."

if ! command -v cargo &> /dev/null; then
    echo "❌ Cargo 未安装！请先安装 Rust: https://rustup.rs/"
    exit 1
fi

echo "✅ Cargo 已安装：$(cargo --version)"

# 添加 Windows 目标
if ! rustup target list | grep -q "x86_64-pc-windows-gnu"; then
    echo "⚠️ Windows 目标未添加！正在添加..."
    rustup target add x86_64-pc-windows-gnu
fi

echo "✅ Rust 已配置好 Windows 交叉编译"

# 更新版本号
echo ""
echo "[2/4] 更新版本号为 v${VERSION}..."

sed -i.bak "s/^version = \".*\"$/version = \"$VERSION\"/" "$WINDOWS_CLIENT_DIR/Cargo.toml"
rm -f "$WINDOWS_CLIENT_DIR/Cargo.toml.bak"

sed -i.bak "s/!define APP_VERSION \".*\"/!define APP_VERSION \"$VERSION\"/" "$WINDOWS_CLIENT_DIR/installer/windows_client.nsi"
rm -f "$WINDOWS_CLIENT_DIR/installer/windows_client.nsi.bak"

echo "✅ 已更新 Cargo.toml 和 windows_client.nsi"

# 交叉编译
echo ""
echo "[3/4] 编译 Windows 客户端 (x86_64-pc-windows-gnu)..."

cd "$WINDOWS_CLIENT_DIR"

echo "   正在编译..."
cargo build --release --target x86_64-pc-windows-gnu

BINARY_PATH="target/x86_64-pc-windows-gnu/release/windows_hardening_client.exe"

if [ -f "$BINARY_PATH" ]; then
    FILE_SIZE=$(stat -f%z "$BINARY_PATH" 2>/dev/null || stat -c%s "$BINARY_PATH" 2>/dev/null)
    FILE_SIZE_MB=$(echo "scale=2; $FILE_SIZE / 1048576" | bc)
    echo "✅ 编译成功!"
    echo "   二进制文件：$BINARY_PATH (${FILE_SIZE_MB} MB)"
else
    echo "❌ 编译失败，未找到二进制文件：$BINARY_PATH"
    exit 1
fi

cd "$SCRIPT_DIR"

# NSIS 打包
echo ""
echo "[4/4] NSIS 打包..."

mkdir -p "$DIST_WIN_DIR"

NSIS_CMD=""
if command -v makensis &> /dev/null; then
    NSIS_CMD="makensis"
elif which nsis > /dev/null 2>&1; then
    NSIS_CMD="/Applications/NSIS/makensis"
fi

if [ -z "$NSIS_CMD" ]; then
    echo "⚠️ NSIS 未找到！请手动在 Windows 机器上编译："
    echo "   cd windows-client/installer"
    echo "   makensis windows_client.nsi"
    echo ""
    echo "当前生成的文件:"
    echo "  ✅ Cargo.toml (已更新为 v${VERSION})"
    echo "  ✅ windows_client.nsi (已更新)"
    echo "  ✅ 二进制文件：$BINARY_PATH"
else
    echo "✅ 找到 NSIS: $NSIS_CMD"
    
    cd "$WINDOWS_CLIENT_DIR/installer"
    $NSIS_CMD windows_client.nsi
    
    if [ $? -eq 0 ]; then
        INSTALLER_EXE=$(ls -lt "$DIST_WIN_DIR"/SystemHardening_WindowsClient_Setup_*.exe 2>/dev/null | head -1 | awk '{print $NF}')
        
        if [ -n "$INSTALLER_EXE" ] && [ -f "$INSTALLER_EXE" ]; then
            FILE_SIZE=$(stat -f%z "$INSTALLER_EXE" 2>/dev/null || stat -c%s "$INSTALLER_EXE" 2>/dev/null)
            FILE_SIZE_MB=$(echo "scale=2; $FILE_SIZE / 1048576" | bc)
            
            echo ""
            echo "=========================================================="
            echo "🎉 Windows 客户端 NSIS 安装包构建完成!"
            echo "=========================================================="
            echo "文件名：$(basename "$INSTALLER_EXE")"
            echo "路径：$INSTALLER_EXE"
            echo "大小：${FILE_SIZE_MB} MB"
            echo "版本：$VERSION"
            
            # 计算 MD5
            if md5 -q "$INSTALLER_EXE" 2>/dev/null; then
                MD5=$(md5 -q "$INSTALLER_EXE")
            elif md5sum "$INSTALLER_EXE" 2>/dev/null | awk '{print $1}' | head -1; then
                MD5=$(md5sum "$INSTALLER_EXE" | awk '{print $1}')
            else
                MD5="(无法计算)"
            fi
            echo "MD5: $MD5"
            echo ""
            echo "📦 此安装包包含："
            echo "  ✅ Windows 客户端程序"
            echo "  ✅ config.yaml 配置文件模板"
            echo "  ✅ systemd 服务注册（Service Name: SystemHardeningWinClient）"
            echo "  ✅ 自动启动配置"
            echo "  ✅ 卸载程序"
            echo ""
        fi
    else
        echo "❌ NSIS 编译失败!"
        exit 1
    fi
fi

echo ""
echo "============================================================"
echo "构建完成！"
echo "============================================================"
