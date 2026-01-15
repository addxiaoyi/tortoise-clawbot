// 账号相关
export interface Account {
  id: string
  username: string
  password: string
  platform: Platform
  cookies?: Record<string, string>
  status: 'active' | 'inactive' | 'error'
  lastLogin?: string
  examCount: number
  successRate: number
  createdAt: string
}

export type Platform = 
  | 'chaoxing'      // 学习通
  | 'icve'          // 智慧职教
  | 'yuketang'      // 雨课堂
  | 'edx'           // 学堂在线
  | 'mooc'          // 中国大学MOOC
  | 'zhihuishu'     // 智慧树
  | 'tencentschool' // 腾讯课堂
  | 'dingtalk'      // 钉钉云课堂
  | 'generic'       // 通用平台

export const PlatformNames: Record<Platform, string> = {
  chaoxing: '学习通',
  icve: '智慧职教',
  yuketang: '雨课堂',
  edx: '学堂在线',
  mooc: '中国大学MOOC',
  zhihuishu: '智慧树',
  tencentschool: '腾讯课堂',
  dingtalk: '钉钉云课堂',
  generic: '通用平台',
}

// 考试相关
export interface Exam {
  id: string
  accountId: string
  title: string
  course: string
  platform: Platform
  url: string
  status: ExamStatus
  scheduledTime?: string
  scheduledType?: 'once' | 'daily' | 'weekly'
  scheduledCron?: string
  progress: number
  score?: number
  duration?: number
  answers?: AnswerRecord[]
  startTime?: string
  endTime?: string
  error?: string
  createdAt: string
}

export type ExamStatus = 
  | 'pending'     // 待执行
  | 'running'     // 执行中
  | 'completed'   // 已完成
  | 'failed'      // 失败
  | 'cancelled'   // 已取消

export interface AnswerRecord {
  question: string
  answer: string
  isCorrect?: boolean
  source: 'manual' | 'search' | 'cache' | 'ai'
  confidence?: number
}

// 任务调度相关
export interface ScheduledTask {
  id: string
  name: string
  examId: string
  accountId: string
  cronExpression: string
  enabled: boolean
  nextRun?: string
  lastRun?: string
  runCount: number
}

// 通知相关
export interface Notification {
  id: string
  type: 'dingtalk' | 'wechat' | 'feishu' | 'webhook'
  config: DingtalkConfig | WechatConfig | FeishuConfig | WebhookConfig
  events: NotificationEvent[]
  enabled: boolean
}

export type NotificationEvent = 
  | 'exam.start'
  | 'exam.complete'
  | 'exam.fail'
  | 'account.error'
  | 'daily.report'

export interface DingtalkConfig {
  webhook: string
  secret?: string
}

export interface WechatConfig {
  corpId: string
  corpSecret: string
  agentId: string
}

export interface FeishuConfig {
  webhook: string
  secret?: string
}

export interface WebhookConfig {
  url: string
  headers?: Record<string, string>
}

// 代理相关
export interface Proxy {
  id: string
  url: string
  type: 'http' | 'https' | 'socks5'
  username?: string
  password?: string
  status: 'active' | 'inactive' | 'error'
  successRate: number
  lastUsed?: string
  usedCount: number
}

// 统计相关
export interface Stats {
  totalExams: number
  completedExams: number
  failedExams: number
  successRate: number
  avgScore: number
  avgDuration: number
  platformStats: PlatformStat[]
  dailyStats: DailyStat[]
}

export interface PlatformStat {
  platform: Platform
  count: number
  successRate: number
}

export interface DailyStat {
  date: string
  completed: number
  failed: number
  avgScore: number
}

// 题库相关
export interface Question {
  id: string
  content: string
  type: 'single' | 'multiple' | 'judge' | 'blank' | 'essay'
  options?: QuestionOption[]
  answer?: string | string[]
  source?: string
  difficulty?: number
  tags?: string[]
  createdAt: string
}

export interface QuestionOption {
  key: string
  content: string
}

// 搜索结果
export interface SearchResult {
  question: string
  answer: string
  source: string
  url: string
  confidence: number
}

// API 响应
export interface ApiResponse<T = unknown> {
  success: boolean
  data?: T
  error?: string
  message?: string
}

// 分页
export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}
