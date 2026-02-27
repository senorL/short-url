#!/bin/bash

echo "开始测试"

# 1.POST所需的 lua 脚本
cat <<EOF > post.lua
wrk.method = "POST"
wrk.body   = '{"url": "https://github.com"}'
wrk.headers["Content-Type"] = "application/json"
EOF

# 2. 压测参数
THREADS=12        # 线程数
CONNECTIONS=1000  # 并发连接数
DURATION=60s      # 接口测试的秒数


SHORTCODE="VLBH4Z"

echo ""
echo "[1/2] 正在测试 GET 接口 (短链重定向) - 压力: $CONNECTIONS 并发..."
wrk -t$THREADS -c$CONNECTIONS -d$DURATION http://localhost:8080/$SHORTCODE

echo ""
echo "2/2] 正在测试 POST 接口 (生成短链) - 压力: $CONNECTIONS 并发..."
wrk -t$THREADS -c$CONNECTIONS -d$DURATION -s post.lua http://localhost:8080/shorten

echo "结束"
echo "清理临时文件..."
rm post.lua
echo "=================================================="