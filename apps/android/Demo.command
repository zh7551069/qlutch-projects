#!/usr/bin/env bash
# Finder 双击入口：起模拟器 → 编译安装 → 看效果。
# Terminal 执行 .command 不会加载 ~/.zshrc，PATH 由 demo.sh 自己拼，
# 所以这里不用 source 任何 shell 配置。
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")" || exit 1

SCRIPT="$(cd ../.. && pwd)/tools/android/demo.sh"
if [ ! -x "$SCRIPT" ]; then
  echo "找不到 $SCRIPT"
  echo "先执行：chmod +x tools/android/demo.sh"
  read -n 1 -s -r -p "按任意键关闭…"
  exit 1
fi

"$SCRIPT" "$@"
status=$?

echo
if [ "$status" -ne 0 ]; then
  echo "退出码 $status —— 上面有报错，按任意键关闭窗口"
  read -n 1 -s -r
fi
exit "$status"
