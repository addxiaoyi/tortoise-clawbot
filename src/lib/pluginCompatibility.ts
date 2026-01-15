import { CoreAgentId } from './coreAgents';

export type CompatibilityLabel = '🪽 仅 Hermes' | '🔵 仅 OpenClaw' | '🟢 仅 myclaw' | '🟣 仅 nanobot' | '✅ 全平台';

interface SkillCompatibility {
  supported: CoreAgentId[];
}

export const PLUGIN_COMPATIBILITY: Record<string, SkillCompatibility> = {
  runtime: { supported: ['hermes'] },
  'skill-registry': { supported: ['hermes'] },
  'local-host': { supported: ['hermes'] },

  gateway: { supported: ['hermes', 'openclaw'] },
  github: { supported: ['openclaw'] },
  slack: { supported: ['openclaw'] },
  discord: { supported: ['openclaw'] },
  'notion-page': { supported: ['openclaw'] },
  'notion-database': { supported: ['openclaw'] },
  'notion-wiki': { supported: ['openclaw'] },
  'memory-setup': { supported: ['hermes', 'openclaw'] },
  'agent-memory': { supported: ['openclaw'] },
  'self-improvement': { supported: ['openclaw'] },
  confluence: { supported: ['openclaw'] },
  'confluence-cloud': { supported: ['openclaw'] },
  sharepoint: { supported: ['openclaw'] },
  langchain: { supported: ['openclaw'] },
  llamaindex: { supported: ['openclaw'] },
  autogen: { supported: ['openclaw'] },
  crewai: { supported: ['openclaw'] },
  dspy: { supported: ['openclaw'] },
  traceloop: { supported: ['openclaw'] },

  telegram: { supported: ['myclaw'] },
  'feishu-doc': { supported: ['myclaw'] },
  'feishu-drive': { supported: ['myclaw'] },
  'feishu-perm': { supported: ['myclaw'] },
  'feishu-wiki': { supported: ['myclaw'] },
  'feishu-meeting': { supported: ['myclaw'] },
  'feishu-calendar': { supported: ['myclaw'] },
  'feishu-approval': { supported: ['myclaw'] },
  'wecom-doc': { supported: ['myclaw'] },
  'wecom-meeting': { supported: ['myclaw'] },
  'dingtalk-doc': { supported: ['myclaw'] },
  'dingtalk-drive': { supported: ['myclaw'] },
  'dingtalk-meeting': { supported: ['myclaw'] },
  'lark-doc': { supported: ['myclaw'] },
  'lark-sheet': { supported: ['myclaw'] },
  'lark-base': { supported: ['myclaw'] },

  'multi-agent-orchestration': { supported: ['nanobot'] },
  'a2a-protocol': { supported: ['nanobot'] },
  'heartbeat-schedule': { supported: ['nanobot'] },
  'context-tiered': { supported: ['nanobot'] },
  'sandbox-security': { supported: ['nanobot'] },
  'mongodb-change-stream': { supported: ['nanobot'] },
  'knowledge-沉淀': { supported: ['nanobot'] },
  'agent-team-orchestration': { supported: ['nanobot'] },
  'agent-spawner': { supported: ['nanobot'] },
  'agent-swarm': { supported: ['nanobot'] },
  'agent-orchestrator': { supported: ['nanobot'] },
  'agent-registry': { supported: ['nanobot'] },
  'agent-dispatch': { supported: ['nanobot'] },
  'context-engine': { supported: ['nanobot'] },

  'tavily-search': { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
  'duckduckgo-search': { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
  'google-search': { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
  screenshot: { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
  'ocr-text': { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
  'video-subtitle': { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
  videocut: { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
  'lmstudio-local': { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
  'ollama-local': { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
  'deepseek-api': { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
  'qwen-api': { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
  'moonshot-api': { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
  'volcengine-api': { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
  'glm-api': { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
  langfuse: { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
  langsmith: { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
  'coding-agent': { supported: ['hermes', 'openclaw', 'myclaw', 'nanobot'] },
};

export function getCompatibilityLabel(skillId: string, selectedCore: CoreAgentId | null): CompatibilityLabel {
  const compatibility = PLUGIN_COMPATIBILITY[skillId];

  if (!compatibility) {
    return '✅ 全平台';
  }

  const supported = compatibility.supported;

  if (supported.length === 1) {
    if (supported[0] === 'hermes') return '🪽 仅 Hermes';
    if (supported[0] === 'openclaw') return '🔵 仅 OpenClaw';
    if (supported[0] === 'myclaw') return '🟢 仅 myclaw';
    if (supported[0] === 'nanobot') return '🟣 仅 nanobot';
  }

  if (selectedCore && supported.includes(selectedCore)) {
    return '✅ 全平台';
  }

  return '✅ 全平台';
}

export function isSkillCompatible(skillId: string, coreId: CoreAgentId): boolean {
  const compatibility = PLUGIN_COMPATIBILITY[skillId];
  if (!compatibility) return true;
  return compatibility.supported.includes(coreId);
}

export function getExclusiveSkillsForCore(coreId: CoreAgentId): string[] {
  return Object.entries(PLUGIN_COMPATIBILITY)
    .filter(([_, compat]) => compat.supported.length === 1 && compat.supported[0] === coreId)
    .map(([skillId]) => skillId);
}

export function getCrossPlatformSkills(): string[] {
  return Object.entries(PLUGIN_COMPATIBILITY)
    .filter(([_, compat]) => compat.supported.length > 1)
    .map(([skillId]) => skillId);
}
