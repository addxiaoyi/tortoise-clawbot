import {
  classifyHermesInvokeError,
  invokeHermesSkillTool,
} from "../hermes/runtime/skill-runner";
import type { HermesInvokeOptions } from "../hermes/runtime/types";
import type { TohelpOpenClawApi } from "./openclaw-api-types";

/** Stable codes for log filtering / metrics (not HTTP status). */
export const classifyInvokeError = classifyHermesInvokeError;

export type InvokeSkillToolOptions = HermesInvokeOptions;

/**
 * Compatibility shim for the legacy bridge name. New code should call
 * `invokeHermesSkillTool` from `src/hermes/runtime/skill-runner` directly.
 */
export async function invokeSkillTool(
  api: TohelpOpenClawApi,
  skill: string,
  tool: string,
  args: Record<string, unknown>,
  options?: InvokeSkillToolOptions,
): Promise<unknown> {
  const compatApi = {
    ...api,
    config: api.pluginConfig ?? api.config,
    logger: {
      ...api.logger,
      info: (message: string, ...rest: unknown[]) => {
        const normalized = message.startsWith("[hermes]")
          ? message.replace(/^\[hermes\]/, "[tohelp]")
          : message;
        api.logger.info(normalized, ...rest);
      },
      error: (message: string, error?: Error, ...rest: unknown[]) => {
        const normalized = message.startsWith("[hermes]")
          ? message.replace(/^\[hermes\]/, "[tohelp]")
          : message;
        api.logger.error(normalized, error, ...rest);
      },
    },
  };

  try {
    return await invokeHermesSkillTool(
      compatApi,
      skill,
      tool,
      args,
      options,
    );
  } catch (error) {
    if (
      error instanceof Error &&
      error.message.startsWith("[hermes_invoke_skill]")
    ) {
      throw new Error(
        error.message.replace(/^\[hermes_invoke_skill\]/, "[tohelp_invoke_skill]"),
        { cause: error.cause },
      );
    }
    throw error;
  }
}
