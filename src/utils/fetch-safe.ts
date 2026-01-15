/**
 * 出站 fetch 的保守默认项：不跟随重定向，降低 token 被带到非预期主机的风险。
 * 各调用方再合并 method / headers / body 等。
 */
export const FETCH_NO_REDIRECT = { redirect: "error" as const };
