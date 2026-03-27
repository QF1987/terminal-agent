import { Agent } from "@mariozechner/pi-agent-core";
import { getModel } from "@mariozechner/pi-ai";

const model = getModel("openrouter", "google/gemini-2.0-flash-001");
console.log("Model:", model?.id, model?.provider);

const agent = new Agent({
  initialState: {
    systemPrompt: "你是一个助手，回答简洁。",
    model,
    tools: []
  }
});

agent.subscribe((event) => {
  console.log("EVENT:", event.type);
  if (event.type === "message_end") {
    const msg = event.message;
    if (msg.role === "assistant") {
      const texts = msg.content.filter(c => c.type === "text").map(c => c.text);
      console.log("RESPONSE:", texts.join(""));
    }
  }
  if (event.type === "agent_end") {
    console.log("AGENT DONE");
  }
});

try {
  console.log("Calling prompt...");
  await agent.prompt("你好，用一句话回复");
  console.log("Prompt done");
} catch (e) {
  console.error("Error:", e.message);
  console.error(e.stack);
}
