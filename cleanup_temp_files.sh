#!/bin/bash
# ============================================================
# 清理项目中的临时文件和多余文件
# ============================================================

echo "=========================================="
echo " 开始清理项目中的临时文件"
echo "=========================================="

# 获取项目根目录
PROJECT_ROOT="/Users/yeung/Projects/system_hardening"

cd "$PROJECT_ROOT"

echo ""
echo "步骤 1: 清理 dist 目录的备份文件..."
if [ -d "dist" ]; then
    # 只保留最新的 zip 文件（可选，如果需要的话）
    echo "   注意：保留所有安装包，不做删除"
else
    echo "   ⚠️  dist 目录不存在"
fi

echo ""
echo "步骤 2: 清理 backend 目录的临时日志文件..."
if [ -f "backend/backend.log.old" ]; then
    rm -f backend/backend.log.old
    echo "   ✅ 已删除 backend/backend.log.old"
fi

if [ -f "backend/new_backend.log" ]; then
    rm -f backend/new_backend.log
    echo "   ✅ 已删除 backend/new_backend.log"
fi

if [ -f "backend/backend_output.log" ]; then
    rm -f backend/backend_output.log
    echo "   ✅ 已删除 backend/backend_output.log"
fi

echo ""
echo "步骤 3: 清理客户端二进制备份..."
if [ -f "backend/backend_app_old" ]; then
    rm -f backend/backend_app_old
    echo "   ✅ 已删除 backend/backend_app_old"
fi

echo ""
echo "步骤 4: 检查 .qoder 缓存目录..."
if [ -d ".qoder/cache" ]; then
    echo "   ⚠️  发现 .qoder/cache 目录"
    echo "   建议手动清理或使用 git clean -fdx .qoder/"
else
    echo "   ✅ .qoder 缓存目录为空或不适用"
fi

echo ""
echo "步骤 5: 清理 .gitignore 忽略的文件..."
# 查看哪些文件会被 git clean 删除
echo "   以下文件将被清理:"
git ls-files --others --exclude-standard | head -n 20

echo ""
echo "步骤 6: 询问是否执行 git clean..."
read -p "是否执行清理？(y/n): " -n 1 -r
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    git clean -fdx
    echo "✅ Git clean 完成"
else
    echo "⚠️  跳过 git clean 操作"
fi

echo ""
echo "=========================================="
echo "🧹 清理完成！"
echo "=========================================="

# 显示剩余的重要文件
echo ""
echo "当前项目结构概览:"
echo "-------------------"
ls -lh dist/*.zip 2>/dev/null | tail -n 5 || echo "无压缩包"
ls -lh backend/*.go 2>/dev/null | wc -l | xargs -I {} echo "后端 Go 文件：{} 个"
find src -name "*.vue" 2>/dev/null | wc -l | xargs -I {} echo "前端 Vue 文件：{} 个"
