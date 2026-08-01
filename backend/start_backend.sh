#!/bin/bash
# ============================================================
# 启动修复版后端服务
# ============================================================

cd /Users/yeung/Projects/system_hardening/backend

# 检查是否已有运行的进程
if pgrep -f "backend_app" > /dev/null; then
    echo "🚫 检测到已运行的后端进程，正在停止..."
    pkill -f backend_app
    sleep 3
fi

# 确保使用最新编译的版本
echo "✅ 使用最新的 backend_app_new"

# 启动新后端
nohup ./backend_app_new > backend.log 2>&1 &
PID=$!

echo "🚀 后端服务已启动 (PID: $PID)"
sleep 3

# 验证进程是否正常运行
if pgrep -f "backend_app_new" > /dev/null; then
    echo "✅ 后端服务运行正常"
else
    echo "❌ 后端服务启动失败"
    exit 1
fi
