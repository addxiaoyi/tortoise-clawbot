const API_BASE = (import.meta.env.VITE_CLOUD_API_BASE || "http://127.0.0.1:8080").trim();
const AUTH_BASE_PATH = (import.meta.env.VITE_CLOUD_API_BASE_PATH || "/auth").trim();

const apiBaseEl = document.querySelector("#apiBase");
const authPathEl = document.querySelector("#authPath");
const sessionStatusEl = document.querySelector("#sessionStatus");
const outputEl = document.querySelector("#output");
const busyHintEl = document.querySelector("#busyHint");
const logFilterEl = document.querySelector("#logFilter");
const loginFormEl = document.querySelector("#loginForm");
const emailEl = document.querySelector("#email");
const passwordEl = document.querySelector("#password");
const btnLogin = document.querySelector("#btnLogin");
const btnSignup = document.querySelector("#btnSignup");
const btnMe = document.querySelector("#btnMe");
const btnSignout = document.querySelector("#btnSignout");
const btnHealth = document.querySelector("#btnHealth");
const btnCopyLog = document.querySelector("#btnCopyLog");
const btnExportLog = document.querySelector("#btnExportLog");
const btnExportErrorLog = document.querySelector("#btnExportErrorLog");
const btnClearLog = document.querySelector("#btnClearLog");
const actionButtons = [btnLogin, btnSignup, btnMe, btnSignout, btnHealth];
let activeRequestCount = 0;
const logs = [];
let logFilter = "all";

if (apiBaseEl) apiBaseEl.textContent = API_BASE;
if (authPathEl) authPathEl.textContent = AUTH_BASE_PATH;

const AUTH_BASE = `${API_BASE.replace(/\/$/, "")}${AUTH_BASE_PATH.startsWith("/") ? AUTH_BASE_PATH : `/${AUTH_BASE_PATH}`}`;

function setOutput(value) {
  if (!outputEl) return;
  outputEl.textContent = value;
}

function isErrorLevel(level) {
  return level === "error";
}

function getVisibleLogs() {
  if (logFilter === "error") {
    return logs.filter((item) => isErrorLevel(item.level));
  }
  return logs;
}

function getLogText() {
  const visible = getVisibleLogs();
  if (visible.length === 0) return "暂无日志";
  return visible.map((item) => item.text).join("\n\n");
}

function nowText() {
  return new Date().toLocaleTimeString("zh-CN", { hour12: false });
}

function pushLog(message, detail, level = "info") {
  const line = `[${nowText()}] ${message}`;
  logs.push({
    level,
    text: detail ? `${line}\n${detail}` : line,
  });
  if (logs.length > 80) {
    logs.splice(0, logs.length - 80);
  }
  setOutput(getLogText());
}

function updateBusyHint() {
  if (!busyHintEl) return;
  busyHintEl.textContent = activeRequestCount > 0 ? "请求进行中，请稍候..." : "空闲";
}

function setButtonsDisabled(disabled) {
  for (const btn of actionButtons) {
    if (!btn) continue;
    btn.disabled = disabled;
  }
}

function startRequest() {
  activeRequestCount += 1;
  setButtonsDisabled(true);
  updateBusyHint();
}

function finishRequest() {
  activeRequestCount = Math.max(0, activeRequestCount - 1);
  if (activeRequestCount === 0) {
    setButtonsDisabled(false);
  }
  updateBusyHint();
}

function setSessionStatus(value) {
  if (!sessionStatusEl) return;
  sessionStatusEl.textContent = value;
}

function pretty(value) {
  try {
    return JSON.stringify(value, null, 2);
  } catch {
    return String(value);
  }
}

async function parseResponse(res) {
  const text = await res.text();
  if (!text) return null;
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

function getFriendlyErrorMessage(body) {
  if (!body || typeof body !== "object") return "";
  const message = typeof body.message === "string" ? body.message : "";
  const status = typeof body.status === "string" ? body.status : "";
  const code = typeof body.code === "string" ? body.code : "";

  if (status === "WRONG_CREDENTIALS_ERROR") return "邮箱或密码错误。";
  if (status === "EMAIL_ALREADY_EXISTS_ERROR") return "该邮箱已注册，请直接登录。";
  if (status === "FIELD_ERROR") {
    if (message) return `字段校验失败：${message}`;
    return "输入字段不符合要求。";
  }
  if (status === "SIGN_UP_NOT_ALLOWED") return "当前环境不允许注册。";
  if (code === "UNAUTHORISED") return "当前未登录或会话已失效。";
  return "";
}

function formatResult(prefix, result) {
  const friendly = getFriendlyErrorMessage(result.body);
  const extra = friendly ? `\n提示：${friendly}` : "";
  return `${prefix}：${result.status} ${result.statusText}${extra}\n${pretty(result.body)}`;
}

async function loginWithEmailPassword(email, password) {
  const loginUrl = `${AUTH_BASE}/signin`;
  const res = await fetch(loginUrl, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      rid: "emailpassword",
    },
    body: JSON.stringify({ formFields: [{ id: "email", value: email }, { id: "password", value: password }] }),
  });
  return { status: res.status, statusText: res.statusText, body: await parseResponse(res) };
}

async function signUpWithEmailPassword(email, password) {
  const signupUrl = `${AUTH_BASE}/signup`;
  const res = await fetch(signupUrl, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      rid: "emailpassword",
    },
    body: JSON.stringify({ formFields: [{ id: "email", value: email }, { id: "password", value: password }] }),
  });
  return { status: res.status, statusText: res.statusText, body: await parseResponse(res) };
}

async function fetchMe() {
  const meUrl = `${API_BASE.replace(/\/$/, "")}/api/me`;
  const res = await fetch(meUrl, {
    method: "GET",
    credentials: "include",
    headers: {
      Accept: "application/json",
    },
  });
  return { status: res.status, statusText: res.statusText, body: await parseResponse(res) };
}

async function signOut() {
  const signoutUrl = `${AUTH_BASE}/signout`;
  const res = await fetch(signoutUrl, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      rid: "session",
    },
    body: "{}",
  });
  return { status: res.status, statusText: res.statusText, body: await parseResponse(res) };
}

async function checkHealth() {
  startRequest();
  pushLog("开始检查 /health");

  const healthUrl = `${API_BASE.replace(/\/$/, "")}/health`;
  try {
    const res = await fetch(healthUrl, {
      method: "GET",
      headers: { Accept: "application/json" },
    });
    const data = await parseResponse(res);
    pushLog("检查 /health 完成", `${res.status} ${res.statusText}\n${pretty(data)}`);
  } catch (err) {
    pushLog("检查 /health 失败", err instanceof Error ? err.message : "请求失败", "error");
  } finally {
    finishRequest();
  }
}

async function onSubmitLogin(event) {
  event.preventDefault();
  const email = emailEl?.value?.trim() || "";
  const password = passwordEl?.value || "";
  if (!email || !password) {
    setOutput("请先输入邮箱和密码。");
    return;
  }

  startRequest();
  pushLog("开始登录");
  try {
    const result = await loginWithEmailPassword(email, password);
    pushLog("登录完成", formatResult("登录响应", result));
    if (result.status >= 200 && result.status < 300) {
      await refreshSessionStatus();
    }
  } catch (err) {
    pushLog("登录失败", err instanceof Error ? err.message : "登录请求失败", "error");
  } finally {
    finishRequest();
  }
}

async function onCheckMe() {
  startRequest();
  pushLog("开始读取 /api/me");
  try {
    const result = await fetchMe();
    pushLog("读取 /api/me 完成", formatResult("/api/me 响应", result));
    setSessionStatus(result.status === 200 ? "已登录" : "未登录");
  } catch (err) {
    pushLog("读取 /api/me 失败", err instanceof Error ? err.message : "请求 /api/me 失败", "error");
    setSessionStatus("检查失败");
  } finally {
    finishRequest();
  }
}

async function onSignUp() {
  const email = emailEl?.value?.trim() || "";
  const password = passwordEl?.value || "";
  if (!email || !password) {
    setOutput("请先输入邮箱和密码。");
    return;
  }

  startRequest();
  pushLog("开始注册");
  try {
    const result = await signUpWithEmailPassword(email, password);
    pushLog("注册完成", formatResult("注册响应", result));
  } catch (err) {
    pushLog("注册失败", err instanceof Error ? err.message : "注册请求失败", "error");
  } finally {
    finishRequest();
  }
}

async function onSignOut() {
  startRequest();
  pushLog("开始退出登录");
  try {
    const result = await signOut();
    pushLog("退出完成", formatResult("退出响应", result));
    await refreshSessionStatus();
  } catch (err) {
    pushLog("退出失败", err instanceof Error ? err.message : "退出请求失败", "error");
  } finally {
    finishRequest();
  }
}

async function refreshSessionStatus() {
  startRequest();
  try {
    const result = await fetchMe();
    setSessionStatus(result.status === 200 ? "已登录" : "未登录");
  } catch {
    setSessionStatus("检查失败");
  } finally {
    finishRequest();
  }
}

loginFormEl?.addEventListener("submit", onSubmitLogin);
btnSignup?.addEventListener("click", onSignUp);
btnMe?.addEventListener("click", onCheckMe);
btnSignout?.addEventListener("click", onSignOut);
btnHealth?.addEventListener("click", checkHealth);
btnClearLog?.addEventListener("click", () => {
  logs.length = 0;
  setOutput("日志已清空");
});
btnCopyLog?.addEventListener("click", async () => {
  const text = getLogText();
  try {
    await navigator.clipboard.writeText(text);
    pushLog("日志已复制到剪贴板");
  } catch (err) {
    pushLog("复制日志失败", err instanceof Error ? err.message : "浏览器不支持剪贴板写入", "error");
  }
});
btnExportLog?.addEventListener("click", () => {
  const text = getLogText();
  const blob = new Blob([text], { type: "text/plain;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  const time = new Date().toISOString().replace(/[:.]/g, "-");
  a.href = url;
  a.download = `tohelp-auth-log-${time}.txt`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
  pushLog("日志已导出", a.download);
});
btnExportErrorLog?.addEventListener("click", () => {
  const errorOnly = logs.filter((item) => isErrorLevel(item.level)).map((item) => item.text).join("\n\n");
  const text = errorOnly || "暂无错误日志";
  const blob = new Blob([text], { type: "text/plain;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  const time = new Date().toISOString().replace(/[:.]/g, "-");
  a.href = url;
  a.download = `tohelp-auth-error-log-${time}.txt`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
  pushLog("错误日志已导出", a.download);
});
logFilterEl?.addEventListener("change", (event) => {
  const value = event?.target?.value === "error" ? "error" : "all";
  logFilter = value;
  setOutput(getLogText());
  pushLog(`日志过滤已切换为：${value === "error" ? "仅错误" : "全部"}`);
});
pushLog("页面已加载，正在检查会话状态");
refreshSessionStatus();
