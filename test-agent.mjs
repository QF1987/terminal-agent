#!/usr/bin/env node

import { spawn } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const agent = spawn('node', [join(__dirname, 'packages/cli/dist/index.js')], {
  cwd: __dirname,
  stdio: ['pipe', 'pipe', 'pipe'],
  env: {
    ...process.env,
    OPENROUTER_API_KEY: process.env.OPENROUTER_API_KEY || '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'
  }
});

// 收集输出
let output = '';
agent.stdout.on('data', (data) => {
  const text = data.toString();
  output += text;
  process.stdout.write(text);
});

agent.stderr.on('data', (data) => {
  process.stderr.write(data);
});

// 等待 Agent 启动
setTimeout(() => {
  console.log('\n--- 测试自然语言输入 ---\n');
  agent.stdin.write('列出华南地区所有设备\n');
}, 3000);

// 等待响应后退出
setTimeout(() => {
  agent.stdin.write('退出\n');
}, 30000);

agent.on('close', (code) => {
  console.log(`\nAgent 退出，代码: ${code}`);
});
