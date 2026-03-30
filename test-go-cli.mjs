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

// 测试用例
const testCases = [
  // "查看设备 DEV-001 的详情",
  "用Go CLI列出华北地区的设备",
  // "查看最近的告警信息",
  // "退出"
];

let testIndex = 0;

// 等待 Agent 启动
setTimeout(() => {
  console.log('\n--- 测试自然语言输入 ---\n');
  agent.stdin.write(testCases[0] + '\n');
}, 3000);

// 每 15 秒发送下一个测试用例
const interval = setInterval(() => {
  testIndex++;
  if (testIndex < testCases.length) {
    console.log(`\n--- 测试用例 ${testIndex + 1} ---\n`);
    agent.stdin.write(testCases[testIndex] + '\n');
  } else {
    agent.stdin.write('退出\n');
    clearInterval(interval);
  }
}, 15000);

// 60 秒后强制退出
setTimeout(() => {
  agent.kill();
  clearInterval(interval);
}, 60000);

agent.on('close', (code) => {
  console.log(`\nAgent 退出，代码: ${code}`);
});
