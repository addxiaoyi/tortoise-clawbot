export type CoreAgentId = 'hermes' | 'openclaw' | 'myclaw' | 'nanobot';

export interface CoreAgentDef {
  id: CoreAgentId;
  name: string;
  icon: string;
  color: string;
  description: string;
  advantages: string[];
  disadvantages: string[];
  capabilities: string[];
  defaultSkills: string[];
  modelProviders: string[];
  supportedChannels: string[];
  exclusiveSkillCount: number;
}

export const CORE_AGENTS: Record<CoreAgentId, CoreAgentDef> = {
  hermes: {
    id: 'hermes',
    name: 'Hermes Agent',
    icon: '🪽',
    color: '#5B8CFF',
    description: '本地优先、弱依赖 OpenClaw 的 Hermes Agent Runtime',
    advantages: [
      'Hermes-first：runtime、gateway、MCP、memory 都由本项目自管',
      '弱依赖 OpenClaw：保留兼容层但不再要求作为运行前置',
      '更适合桌面宿主：便于 Tauri/ClawX 风格控制台直接管理本地 agent',
      '兼容现有 tohelp_* MCP 工具名，迁移成本更低',
    ],
    disadvantages: [
      '仍处于重构迁移期',
      'OpenClaw 生态的部分现成能力需要通过兼容层接入',
    ],
    capabilities: ['runtime', 'gateway', 'mcp', 'memory', 'skill-registry', 'local-host'],
    defaultSkills: ['gateway', 'memory-setup', 'coding-agent'],
    modelProviders: ['OpenAI', 'Anthropic', 'Google', 'DeepSeek', 'Qwen', 'GLM', 'Ollama', 'LM Studio'],
    supportedChannels: ['本地桌面端', 'MCP 客户端', '兼容 OpenClaw Gateway'],
    exclusiveSkillCount: 6,
  },
  openclaw: {
    id: 'openclaw',
    name: 'OpenClaw',
    icon: '🐉',
    color: '#007AFF',
    description: '成熟的开源 AI Agent 系统',
    advantages: [
      '完整插件生态：Gateway、GitHub、Slack、Notion、Memory 等',
      '成熟稳定：经过大量用户验证',
      '社区活跃：文档完善、插件丰富',
      '企业级功能：Confluence、SharePoint 集成',
    ],
    disadvantages: [
      '学习曲线较陡',
      '资源占用相对较高',
    ],
    capabilities: ['gateway', 'github', 'slack', 'notion', 'memory', 'discord', 'confluence', 'sharepoint'],
    defaultSkills: ['gateway', 'github', 'slack', 'notion-page', 'memory-setup'],
    modelProviders: ['OpenAI', 'Anthropic', 'Google', 'DeepSeek', 'Qwen'],
    supportedChannels: ['Slack', 'Discord', 'Notion', 'GitHub', 'Memory', 'Gateway', 'Confluence', 'SharePoint'],
    exclusiveSkillCount: 19,
  },
  myclaw: {
    id: 'myclaw',
    name: 'myclaw',
    icon: '🐢',
    color: '#34C759',
    description: '多渠道集成的 AI 助手',
    advantages: [
      '多渠道集成：Telegram、飞书、企业微信、WhatsApp',
      'Gateway 模式：完整渠道编排 + cron + heartbeat',
      '部署简单',
      '国内生态完善：飞书、钉钉、企业微信',
    ],
    disadvantages: [
      '插件生态相对较小',
      '社区活跃度较低',
    ],
    capabilities: ['telegram', 'feishu', 'wecom', 'dingtalk', 'lark', 'gateway'],
    defaultSkills: ['telegram', 'feishu-doc', 'feishu-drive'],
    modelProviders: ['OpenAI', 'DeepSeek', 'Qwen', 'Moonshot', 'Volcengine'],
    supportedChannels: ['Telegram', '飞书', '企业微信', 'WhatsApp', '钉钉', 'Lark'],
    exclusiveSkillCount: 16,
  },
  nanobot: {
    id: 'nanobot',
    name: 'nanobot',
    icon: '🤖',
    color: '#AF52DE',
    description: 'AI同事工作流系统',
    advantages: [
      'AI同事工作流：多Agent协作',
      '经济基准测试：真实专业任务评估',
      '多模型支持：GPT-4o、Claude、Gemini、Qwen、GLM-4',
      '沙箱安全：隔离执行环境',
    ],
    disadvantages: [
      '项目较新',
      '文档完善度待提升',
    ],
    capabilities: ['multi-agent', 'context-tiered', 'sandbox', 'a2a-protocol', 'heartbeat-schedule'],
    defaultSkills: ['multi-agent-orchestration', 'a2a-protocol', 'context-tiered'],
    modelProviders: ['OpenAI', 'Anthropic', 'Google', 'DeepSeek', 'Qwen', 'GLM'],
    supportedChannels: ['多Agent内部协作'],
    exclusiveSkillCount: 14,
  },
};

export const CORE_AGENT_LIST = Object.values(CORE_AGENTS);

export function getCoreAgent(id: CoreAgentId): CoreAgentDef | undefined {
  return CORE_AGENTS[id];
}
