# Skills 技能系统

Terminal Agent 支持通过 Skills 扩展 Agent 能力，无需修改核心代码。

## 目录结构

```
skills/
├── your-skill-name/
│   ├── SKILL.md      ← 必需：技能描述文件
│   └── tools.mjs     ← 可选：附加工具
```

## SKILL.md 格式

```markdown
# 技能名称

简要描述这个技能的作用。

## 触发条件

描述什么时候应该激活这个技能。

## 使用说明

1. 第一步做什么
2. 第二步做什么
3. ...
```

## 附加工具（可选）

如果技能需要新的工具，在同目录创建 `tools.mjs`：

```javascript
import { Type } from "@sinclair/typebox";

export default function() {
  return [
    {
      name: "my_tool",
      label: "我的工具",
      description: "工具描述",
      parameters: Type.Object({
        param: Type.String({ description: "参数说明" })
      }),
      async execute(toolCallId, input) {
        // 执行逻辑
        return {
          content: [{ type: "text", text: "结果" }],
          details: {}
        };
      }
    }
  ];
}
```

## 注意事项

- SKILL.md 的标题会成为技能名称
- 技能描述会被注入到 Agent 的 system prompt
- 工具名称不要和核心工具冲突
- 保持描述简洁，避免占用过多 token
