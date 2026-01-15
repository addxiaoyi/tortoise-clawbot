import type { SkillPlugin } from "../plugins/new-core/types";
import { CanvasPlugin } from "../plugins/new-core/canvas/plugin";
import { CodeReviewPlugin } from "../plugins/new-core/code-review/plugin";
import { DebuggingPlugin } from "../plugins/new-core/debugging/plugin";
import { DocumentationPlugin } from "../plugins/new-core/documentation/plugin";
import { GitWorkflowPlugin } from "../plugins/new-core/git-workflow/plugin";
import { GitHubPlugin } from "../plugins/new-core/github/plugin";
import { NotionPlugin } from "../plugins/new-core/notion/plugin";
import { OfficePlugin } from "../plugins/new-core/office/plugin";
import { PlanningPlugin } from "../plugins/new-core/planning/plugin";
import { SecurityPlugin } from "../plugins/new-core/security/plugin";
import { SlackPlugin } from "../plugins/new-core/slack/plugin";
import { TestingPlugin } from "../plugins/new-core/testing/plugin";
import { WebBuilderPlugin } from "../plugins/new-core/web-builder/plugin";

export const SKILL_IDS = [
  "github",
  "slack",
  "notion",
  "canvas",
  "code-review",
  "debugging",
  "documentation",
  "git-workflow",
  "office",
  "planning",
  "security",
  "testing",
  "web-builder",
] as const;

export type SkillId = (typeof SKILL_IDS)[number];

const factories: Record<SkillId, () => SkillPlugin> = {
  github: () => new GitHubPlugin(),
  slack: () => new SlackPlugin(),
  notion: () => new NotionPlugin(),
  canvas: () => new CanvasPlugin(),
  "code-review": () => new CodeReviewPlugin(),
  debugging: () => new DebuggingPlugin(),
  documentation: () => new DocumentationPlugin(),
  "git-workflow": () => new GitWorkflowPlugin(),
  office: () => new OfficePlugin(),
  planning: () => new PlanningPlugin(),
  security: () => new SecurityPlugin(),
  testing: () => new TestingPlugin(),
  "web-builder": () => new WebBuilderPlugin(),
};

export function isSkillId(value: string): value is SkillId {
  return (SKILL_IDS as readonly string[]).includes(value);
}

export function createSkillPlugin(skill: SkillId): SkillPlugin {
  return factories[skill]();
}
