#!/usr/bin/env node

import { spawn } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

console.log('启动 Agent...\n');

const agent = spawn('node', [join(__dirname, 'packages/cli/dist/index.js')], {
  cwd: __dirname,
  stdio: ['pipe', 'inherit', 'inherit'],
  env: {
    ...process.env,
    OPENROUTER_API_KEY: process.env.OPENROUTER_API_KEY || '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'
  }
});

// 等待 Agent 启动
setTimeout(() => {
  console.log('\n--- 发送测试输入 ---\n');
  agent.stdin.write('查看设备 DEV-001 的详情\n');
}, 5000);

// 等待响应后退出
setTimeout(() => {
  console.log('\n--- 发送退出命令 ---\n');
  agent.stdin.write('退出\n');
}, 30000);

agent.on('close', (code) => {
  console.log(`\nAgent 退出，代码: ${code}`);
});

agent.on('error', (error) => {
  console.error('Agent 错误:', error);
});
