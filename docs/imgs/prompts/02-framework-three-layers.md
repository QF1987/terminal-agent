---
illustration_id: 02
type: framework
style: notion
---

三层架构图 - Terminal Agent 技术架构

STRUCTURE: hierarchical (自下而上三层堆叠)

NODES:
- 底层 core（数据层）: 包含"设备模型" "故障日志" "模拟数据(50+)" "内存存储" 四个子模块
- 中层 tools（工具层）: 包含"查设备列表" "单机状态" "故障日志" "统计分析" "修改配置" "远程重启" 六个工具图标
- 顶层 cli（入口层）: 包含"LLM" "工具" "系统提示词" 三个组件，标注"Agent = LLM + Tools + Prompt"

RELATIONSHIPS:
- cli 层向下调用 tools 层（箭头标注"tool-calling loop"）
- tools 层向下访问 core 层（箭头标注"数据读写"）
- 右侧标注依赖：pi-agent-core（运行时）、pi-ai（LLM抽象层）、TypeBox（参数校验）

LABELS:
- 每层标题: "core 数据层" "tools 工具层" "cli 入口层"
- 核心数据: "50+ 台设备" "6 个工具" "< 2000 行代码"
- 底部标注: "LangChain 太重？这套够轻。"

COLORS: 三层用不同色块区分（浅蓝、浅绿、浅橙），箭头用深灰色，关键数字用强调色

STYLE: Notion style — 手绘线条感的架构图，圆角矩形作为节点容器，手绘箭头连接，整体干净简洁但有手绘温度。

Include a subtle watermark "灵枢" positioned at bottom-right.

Clean composition with generous white space. Simple or no background. Main elements centered or positioned by content needs.
Text should be large and prominent with handwritten-style fonts. Keep minimal, focus on keywords.

ASPECT: 16:9
