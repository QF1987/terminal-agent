#!/usr/bin/env node
// ============================================================
// cli/src/index.ts - 命令行入口（主程序）
// ============================================================
// 这是整个程序的起点
// 负责：打印欢迎信息 → 创建 Agent → 读取用户输入 → 调用 Agent → 显示结果
// 类似一个聊天应用，但用命令行实现
// ============================================================
// Shebang：告诉系统这个文件用 node 执行
// 如果把文件设为可执行（chmod +x），可以直接运行 ./index.js

// ─── Node.js 内置模块 ─────────────────────────────────────
import readline from "node:readline";
// readline：逐行读取输入的模块
// 用于实现命令行交互（用户输入一行 → 处理一行）

import { stdin as input, stdout as output } from "node:process";
// stdin：标准输入（键盘）
// stdout：标准输出（终端屏幕）
// 重命名成 input/output，更直观

import { createTerminalAgent } from "./agent.js";
// 导入创建 Agent 的函数

// ─── 打印欢迎信息 ─────────────────────────────────────────
// console.log()：向终端输出一行文本
function printBanner() {
  console.log("╔══════════════════════════════════════╗");
  console.log("║      🤖 终端管理助手 Demo            ║");
  console.log("║      管理对象：自助购药机             ║");
  console.log("╚══════════════════════════════════════╝");
  console.log("");
  console.log("输入自然语言即可提问，例如：");
  console.log("  • 列出华东地区所有设备");
  console.log("  • 查看设备 SH-PD-001 的状态");
  console.log("  • 最近一周故障最多的是哪些设备");
  console.log("  • 把设备 BJ-PD-002 的屏幕亮度改成 80");
  console.log("  • 统计各区域设备情况");
  console.log("");
  console.log("输入 '退出' 或 Ctrl+C 结束会话。");
  console.log("");
}

// ─── 主函数 ─────────────────────────────────────────────
// async function：异步函数，可以使用 await
async function main() {
  // 1. 打印欢迎信息
  printBanner();

  // 2. 检查 API Key
  // process.env：环境变量对象
  // || 逻辑或：左边是 falsy（空字符串也是 falsy）就用右边的
  const apiKey = process.env.OPENROUTER_API_KEY || process.env.OPENAI_API_KEY;
  if (!apiKey) {
    console.log("⚠️  未检测到 OPENROUTER_API_KEY 或 OPENAI_API_KEY，将以纯工具测试模式运行。");
    console.log("");
  }

  // 3. 创建 Agent 实例
  // await：等待异步操作完成
  const agent = await createTerminalAgent();

  // 4. 状态标记：是否正在处理中
  let waiting = false;

  // ─── Agent 事件订阅 ─────────────────────────────────
  // agent.subscribe()：注册事件监听器
  // Agent 会发出各种事件：消息开始、消息结束、Agent 结束等
  let accumulatedContent = "";  // 累积助手回复内容（用于流式输出）

  agent.subscribe((event: any) => {
    // 事件：助手消息内容更新（流式输出）
    if (event.type === "message_update" && event.message?.role === "assistant") {
      const texts = event.message.content
        ?.filter((c: any) => c.type === "text")
        ?.map((c: any) => c.text) || [];

      if (texts.length > 0) {
        const newContent = texts.join("\n");
        if (newContent.length > accumulatedContent.length) {
          accumulatedContent = newContent;
        }
      }
    }

    // 事件：Agent 处理结束 - 打印累积的内容
    if (event.type === "agent_end") {
      if (event.error) {
        process.stdout.write(`\n助手> 处理失败：${event.error}\n`);
      } else if (accumulatedContent) {
        process.stdout.write(`\n\n助手> ${accumulatedContent}\n`);
      }
      accumulatedContent = "";
      waiting = false;
      console.log("");      // 空一行
      // 用 setImmediate 延迟显示提示符，确保消息先打印出来
      setImmediate(() => {
        rl.prompt();
      });
    }
  });

  // ─── 创建 readline 接口 ─────────────────────────────
  // createInterface()：创建一个读取行的接口
  // { input, output }：指定输入源和输出目标
  // prompt: "你> "：用户输入前显示的提示符
  const rl = readline.createInterface({ input, output, prompt: "你> " });

  // 显示第一次提示符
  rl.prompt();

  // ─── 监听用户输入 ─────────────────────────────────
  // rl.on("line", callback)：每输入一行就触发一次
  rl.on("line", async (line) => {
    // .trim()：去掉首尾空白
    const message = line.trim();

    // 空行：直接显示提示符，不做处理
    if (!message) {
      rl.prompt();
      return;  // 提前返回，不执行后面的代码
    }

    // 退出命令
    // ["退出", "exit", "quit"].includes()：检查数组是否包含某个值
    // .toLowerCase()：转小写（不区分大小写）
    if (["退出", "exit", "quit"].includes(message.toLowerCase())) {
      rl.close();  // 关闭 readline 接口
      return;
    }

    // 正在处理中：拒绝新输入
    if (waiting) {
      console.log("正在处理中，请稍候...");
      return;
    }

    // 5. 调用 Agent 处理用户消息
    waiting = true;  // 标记为处理中

    try {
      // agent.prompt()：发送消息给 Agent，等待处理
      await agent.prompt({
        role: "user",           // 消息角色：用户
        content: message,       // 消息内容
        timestamp: Date.now()   // 时间戳（毫秒）
      });
    } catch (error) {
      // 捕获异常
      // instanceof Error：检查 error 是不是 Error 类型
      // 是的话取 error.message，不是的话转成字符串
      const detail = error instanceof Error ? error.message : String(error);
      console.log(`\n助手> 处理失败：${detail}`);

      // 没有 API Key 时给提示
      if (!apiKey) {
        console.log("  💡 请设置 OPENROUTER_API_KEY 环境变量。");
      }

      waiting = false;
      console.log("");
      rl.prompt();
    }
  });

  // ─── 监听关闭事件 ─────────────────────────────────
  // readline 接口关闭时（用户输入"退出"或 Ctrl+C）
  rl.on("close", () => {
    console.log("\n会话已结束。再见！👋");
    process.exit(0);  // 退出程序，0 表示正常退出
  });
}

// ─── 启动程序 ─────────────────────────────────────────
// 调用 main()，捕获任何未处理的异常
main().catch((error) => {
  // console.error()：输出到标准错误（通常也是终端，但语义上是错误信息）
  console.error(`启动失败：${error instanceof Error ? error.message : String(error)}`);
  process.exit(1);  // 退出程序，1 表示异常退出
});
