#!/bin/bash
# Terminal Agent 启动脚本
cd "$(dirname "$0")"
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi
node packages/cli/dist/index.js "$@"
