export interface SlackConfig {
  token: string;
  defaultChannel?: string;
  [key: string]: unknown;
}

export interface SlackMessage {
  channel: string;
  text: string;
  thread_ts?: string;
}

export interface SlackResponse {
  ok: boolean;
  error?: string;
  ts?: string;
  message?: any;
}
