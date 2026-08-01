#!/bin/bash
# 测试 Linux 加固检查 API

echo "=== 测试 Linux 加固检查 API ==="
echo ""

# 假设 token 已经存储在环境变量中
TOKEN=${1:-""}

if [ -z "$TOKEN" ]; then
    echo "请提供认证 token 作为参数运行此脚本:"
    echo "  ./test_linux_checks.sh <your_token>"
    exit 1
fi

echo "正在测试列表接口..."
curl -X GET \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/linux-checks?page=1&pageSize=10 \
  | jq '.'

echo ""
echo "=================================="
echo ""
echo "正在测试详情接口 (ID: 1)..."
curl -X GET \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/linux-checks/1 \
  | jq '.'
