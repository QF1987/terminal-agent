// ============================================================
// index.ts - 工具包入口（Go CLI 版本）
// ============================================================
import type { AgentTool } from "@mariozechner/pi-agent-core";
export { createGoCLITools } from "./go-cli.js";

export function createDeviceManagementTools(): AgentTool[] {
  const { createGoCLITools } = require("./go-cli.js");
  return createGoCLITools();
}
