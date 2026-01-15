/**
 * 在线题库联网搜索模块
 * 集成多个开源题库 API，支持题目自动搜索和答案匹配
 */

import { z } from "zod";

// ============ 类型定义 ============

export interface Question {
  id: string;
  type: "single" | "multiple" | "true_false" | "fill" | "essay";
  question: string;
  options?: { key: string; value: string }[];
  answer?: string | string[];
  explanation?: string;
  source?: string;
  tags?: string[];
}

export interface SearchResult {
  question: string;
  matchedQuestions: Question[];
  confidence: number;
  source: string;
}

export interface QuestionBankConfig {
  enableOnlineSearch: boolean;
  searchTimeout: number;
  fallbackToLocal: boolean;
  cacheResults: boolean;
  cacheExpiry: number; // milliseconds
  preferredSources: string[];
}

// ============ 开源题库 API 配置 ============

const QUESTION_BANK_APIS = {
  // Open Trivia Database - 免费开源题库
  opentdb: {
    name: "Open Trivia Database",
    baseUrl: "https://opentdb.com",
    category: {
      general: 9,
      science: 17,
      history: 23,
      geography: 22,
      literature: 10,
      math: 19,
    },
    supportsCategories: true,
  },

  // Quiz API - 综合题库
  quizapi: {
    name: "Quiz API",
    baseUrl: "https://quizapi.io/api/v1",
    supportsDifficulty: true,
    supportsTags: true,
  },

  // Jeopardy - 知识问答
  jeopardy: {
    name: "JService Jeopardy",
    baseUrl: "https://jservice.io/api",
    randomEndpoint: "/random",
  },
} as const;

// ============ 本地缓存 ============

class SearchCache {
  private cache = new Map<string, { data: SearchResult; expiry: number }>();

  get(key: string, maxAge: number): SearchResult | null {
    const entry = this.cache.get(key);
    if (!entry) return null;
    if (Date.now() > entry.expiry) {
      this.cache.delete(key);
      return null;
    }
    return entry.data;
  }

  set(key: string, data: SearchResult, maxAge: number): void {
    this.cache.set(key, {
      data,
      expiry: Date.now() + maxAge,
    });
  }

  clear(): void {
    this.cache.clear();
  }
}

// ============ API 响应类型 ============

const OpenTDBResponseSchema = z.object({
  response_code: z.number(),
  results: z.array(
    z.object({
      question: z.string(),
      correct_answer: z.string(),
      incorrect_answers: z.array(z.string()),
      category: z.string(),
      difficulty: z.string(),
      type: z.string(),
    })
  ),
});

const QuizAPIResponseSchema = z.object({
  id: z.number().optional(),
  question: z.string(),
  description: z.string().optional(),
  answer: z.string().optional(),
  multiple_correct_answers: z.string().optional(),
  answers: z
    .object({
      answer_a: z.string().optional(),
      answer_b: z.string().optional(),
      answer_c: z.string().optional(),
      answer_d: z.string().optional(),
    })
    .optional(),
  correct_answers: z
    .object({
      answer_a_correct: z.string().optional(),
      answer_b_correct: z.string().optional(),
      answer_c_correct: z.string().optional(),
      answer_d_correct: z.string().optional(),
    })
    .optional(),
  category: z.string().optional(),
  difficulty: z.string().optional(),
});

// ============ 题库搜索引擎 ============

export class OnlineQuestionBank {
  private cache: SearchCache;
  private config: QuestionBankConfig;
  private httpClient: typeof fetch;

  constructor(config: Partial<QuestionBankConfig> = {}) {
    this.cache = new SearchCache();
    this.config = {
      enableOnlineSearch: true,
      searchTimeout: 10000,
      fallbackToLocal: true,
      cacheResults: true,
      cacheExpiry: 24 * 60 * 60 * 1000, // 24 hours
      preferredSources: ["opentdb", "quizapi", "jeopardy"],
      ...config,
    };
    this.httpClient = fetch;
  }

  /**
   * 搜索题目
   */
  async search(question: string): Promise<SearchResult> {
    // 检查缓存
    const cacheKey = this.normalizeQuestion(question);
    if (this.config.cacheResults) {
      const cached = this.cache.get(cacheKey, this.config.cacheExpiry);
      if (cached) {
        return { ...cached, source: `[Cache] ${cached.source}` };
      }
    }

    // 并行搜索多个源
    const results = await Promise.allSettled([
      this.searchOpenTDB(question),
      this.searchQuizAPI(question),
      this.searchJeopardy(question),
    ]);

    // 合并结果
    const matchedQuestions: Question[] = [];
    let bestConfidence = 0;
    let primarySource = "Unknown";

    for (const result of results) {
      if (result.status === "fulfilled" && result.value.length > 0) {
        matchedQuestions.push(...result.value);
        const confidence = this.calculateConfidence(question, result.value);
        if (confidence > bestConfidence) {
          bestConfidence = confidence;
          primarySource = this.getSourceName(result);
        }
      }
    }

    const searchResult: SearchResult = {
      question,
      matchedQuestions,
      confidence: bestConfidence,
      source: primarySource,
    };

    // 缓存结果
    if (this.config.cacheResults) {
      this.cache.set(cacheKey, searchResult, this.config.cacheExpiry);
    }

    return searchResult;
  }

  /**
   * 批量搜索
   */
  async searchBatch(questions: string[]): Promise<SearchResult[]> {
    return Promise.all(questions.map((q) => this.search(q)));
  }

  /**
   * Open Trivia Database 搜索
   */
  private async searchOpenTDB(question: string): Promise<Question[]> {
    try {
      // 提取关键词
      const keywords = this.extractKeywords(question);

      // 尝试使用关键词搜索
      const url = new URL(`${QUESTION_BANK_APIS.opentdb.baseUrl}/api.php`);
      url.searchParams.set("amount", "10");

      if (keywords.length > 0) {
        url.searchParams.set("category", "9"); // 默认通用类别
      }

      const response = await this.fetchWithTimeout(
        url.toString(),
        this.config.searchTimeout
      );

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      const data = await response.json();
      const parsed = OpenTDBResponseSchema.safeParse(data);

      if (!parsed.success) {
        return [];
      }

      return parsed.data.results.map((item, index) => ({
        id: `opentdb-${Date.now()}-${index}`,
        type: item.type === "multiple" ? "multiple" : "single",
        question: this.decodeHTML(item.question),
        options: this.shuffleOptions([
          { key: "A", value: item.correct_answer },
          ...item.incorrect_answers.map((ans, i) => ({
            key: String.fromCharCode(66 + i),
            value: ans,
          })),
        ]),
        answer: item.correct_answer,
        source: "Open Trivia Database",
        tags: [item.category, item.difficulty],
        explanation: `类别: ${item.category}, 难度: ${item.difficulty}`,
      }));
    } catch (error) {
      console.error("OpenTDB search failed:", error);
      return [];
    }
  }

  /**
   * QuizAPI 搜索
   */
  private async searchQuizAPI(question: string): Promise<Question[]> {
    try {
      // QuizAPI 需要 API key，这里使用公开端点
      const url = new URL(`${QUESTION_BANK_APIS.quizapi.baseUrl}/questions`);
      url.searchParams.set("limit", "10");
      url.searchParams.set("tags", this.extractKeywords(question).join(","));

      const response = await this.fetchWithTimeout(
        url.toString(),
        this.config.searchTimeout
      );

      if (!response.ok) {
        return [];
      }

      const data = await response.json();
      const parsed = z.array(QuizAPIResponseSchema).safeParse(data);

      if (!parsed.success) {
        return [];
      }

      return parsed.data.map((item, index) => ({
        id: `quizapi-${Date.now()}-${index}`,
        type: item.multiple_correct_answers === "true" ? "multiple" : "single",
        question: item.question,
        options: this.formatQuizAPIAnswers(item),
        answer: item.answer || this.extractCorrectAnswer(item),
        source: "Quiz API",
        tags: [item.category || "", item.difficulty || ""].filter(Boolean),
      }));
    } catch (error) {
      console.error("QuizAPI search failed:", error);
      return [];
    }
  }

  /**
   * Jeopardy 搜索
   */
  private async searchJeopardy(question: string): Promise<Question[]> {
    try {
      const keywords = this.extractKeywords(question).slice(0, 3);
      const url = new URL(`${QUESTION_BANK_APIS.jeopardy.baseUrl}/search`);
      url.searchParams.set("str", keywords.join(" "));

      const response = await this.fetchWithTimeout(
        url.toString(),
        this.config.searchTimeout
      );

      if (!response.ok) {
        return [];
      }

      const data = await response.json();

      if (!Array.isArray(data)) {
        return [];
      }

      return data.slice(0, 10).map((item: any, index: number) => ({
        id: `jeopardy-${Date.now()}-${index}`,
        type: "single",
        question: item.question || item.clue,
        answer: item.answer,
        source: "JService Jeopardy",
        tags: [item.category, item.value, item.air_date].filter(Boolean),
        explanation: `类别: ${item.category} | 价值: ${item.value}`,
      }));
    } catch (error) {
      console.error("Jeopardy search failed:", error);
      return [];
    }
  }

  // ============ 辅助方法 ============

  private async fetchWithTimeout(
    url: string,
    timeout: number
  ): Promise<Response> {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);

    try {
      const response = await this.httpClient(url, {
        signal: controller.signal,
        headers: {
          Accept: "application/json",
          "User-Agent": "Tohelp-QuestionBank/1.0",
        },
      });
      return response;
    } finally {
      clearTimeout(timeoutId);
    }
  }

  private normalizeQuestion(question: string): string {
    return question
      .toLowerCase()
      .replace(/[^\w\u4e00-\u9fa5]/g, "")
      .trim();
  }

  private extractKeywords(question: string): string[] {
    // 移除常见的问题前缀
    const cleaned = question
      .replace(
        /^(以下|请|请问|帮我|帮我查|查找|搜索|找一下|答案是|正确的|下面哪个|下列)/gi,
        ""
      )
      .replace(/\?{1,}$/, "")
      .trim();

    // 提取中文和英文关键词
    const keywords: string[] = [];
    const chineseMatch = cleaned.match(/[\u4e00-\u9fa5]+/g);
    if (chineseMatch) {
      keywords.push(...chineseMatch.slice(0, 5));
    }

    const englishMatch = cleaned.match(/[a-zA-Z]{3,}/g);
    if (englishMatch) {
      keywords.push(...englishMatch.slice(0, 3));
    }

    return keywords;
  }

  private calculateConfidence(
    original: string,
    matched: Question[]
  ): number {
    if (matched.length === 0) return 0;

    const normalizedOriginal = this.normalizeQuestion(original);

    for (const q of matched) {
      const normalizedQ = this.normalizeQuestion(q.question);
      const similarity = this.stringSimilarity(normalizedOriginal, normalizedQ);

      if (similarity > 0.8) return 0.9;
      if (similarity > 0.6) return 0.7;
      if (similarity > 0.4) return 0.5;
    }

    return 0.3;
  }

  private stringSimilarity(a: string, b: string): number {
    if (a === b) return 1;
    if (a.length === 0 || b.length === 0) return 0;

    const longer = a.length > b.length ? a : b;
    const shorter = a.length > b.length ? b : a;

    if (longer.includes(shorter)) {
      return shorter.length / longer.length;
    }

    // 简单的编辑距离相似度
    const editDistance = this.levenshteinDistance(a, b);
    return (longer.length - editDistance) / longer.length;
  }

  private levenshteinDistance(a: string, b: string): number {
    const matrix: number[][] = [];

    for (let i = 0; i <= b.length; i++) {
      matrix[i] = [i];
    }
    for (let j = 0; j <= a.length; j++) {
      matrix[0][j] = j;
    }

    for (let i = 1; i <= b.length; i++) {
      for (let j = 1; j <= a.length; j++) {
        if (b.charAt(i - 1) === a.charAt(j - 1)) {
          matrix[i][j] = matrix[i - 1][j - 1];
        } else {
          matrix[i][j] = Math.min(
            matrix[i - 1][j - 1] + 1,
            matrix[i][j - 1] + 1,
            matrix[i - 1][j] + 1
          );
        }
      }
    }

    return matrix[b.length][a.length];
  }

  private decodeHTML(html: string): string {
    return html
      .replace(/&quot;/g, '"')
      .replace(/&#039;/g, "'")
      .replace(/&amp;/g, "&")
      .replace(/&lt;/g, "<")
      .replace(/&gt;/g, ">")
      .replace(/&eacute;/g, "é")
      .replace(/&ntilde;/g, "ñ");
  }

  private shuffleOptions(
    options: { key: string; value: string }[]
  ): { key: string; value: string }[] {
    return options
      .map((item) => ({ item, sort: Math.random() }))
      .sort((a, b) => a.sort - b.sort)
      .map(({ item }) => item);
  }

  private formatQuizAPIAnswers(item: z.infer<typeof QuizAPIResponseSchema>) {
    const options: { key: string; value: string }[] = [];
    const answers = item.answers;
    const correct = item.correct_answers;

    if (!answers || !correct) return options;

    const mappings = [
      { key: "A", answer: answers.answer_a, correct: correct.answer_a_correct },
      { key: "B", answer: answers.answer_b, correct: correct.answer_b_correct },
      { key: "C", answer: answers.answer_c, correct: correct.answer_c_correct },
      { key: "D", answer: answers.answer_d, correct: correct.answer_d_correct },
    ];

    for (const m of mappings) {
      if (m.answer) {
        options.push({
          key: m.key,
          value: m.answer,
        });
      }
    }

    return options;
  }

  private extractCorrectAnswer(
    item: z.infer<typeof QuizAPIResponseSchema>
  ): string {
    const correct = item.correct_answers;
    if (!correct) return "";

    const mappings = [
      { key: "A", correct: correct.answer_a_correct },
      { key: "B", correct: correct.answer_b_correct },
      { key: "C", correct: correct.answer_c_correct },
      { key: "D", correct: correct.answer_d_correct },
    ];

    const answers = item.answers;
    if (!answers) return "";

    for (const m of mappings) {
      if (m.correct === "true") {
        return answers[`answer_${m.key.toLowerCase()}` as keyof typeof answers] || "";
      }
    }

    return "";
  }

  private getSourceName(result: PromiseResult<Question[], any>): string {
    if (result.status === "fulfilled") {
      return result.value[0]?.source || "Unknown";
    }
    return "Failed";
  }

  /**
   * 清除缓存
   */
  clearCache(): void {
    this.cache.clear();
  }

  /**
   * 获取配置
   */
  getConfig(): QuestionBankConfig {
    return { ...this.config };
  }
}

// ============ 工厂函数 ============

export function createOnlineQuestionBank(
  config?: Partial<QuestionBankConfig>
): OnlineQuestionBank {
  return new OnlineQuestionBank(config);
}

// ============ 预定义题库类别 ============

export const QUESTION_CATEGORIES = {
  opentdb: [
    { id: 9, name: "General Knowledge", nameZh: "基础知识" },
    { id: 10, name: "Entertainment: Books", nameZh: "娱乐: 书籍" },
    { id: 11, name: "Entertainment: Film", nameZh: "娱乐: 电影" },
    { id: 12, name: "Entertainment: Music", nameZh: "娱乐: 音乐" },
    { id: 14, name: "Entertainment: Television", nameZh: "娱乐: 电视" },
    { id: 15, name: "Entertainment: Video Games", nameZh: "娱乐: 电子游戏" },
    { id: 17, name: "Science & Nature", nameZh: "科学与自然" },
    { id: 18, name: "Science: Computers", nameZh: "科学: 计算机" },
    { id: 19, name: "Science: Mathematics", nameZh: "科学: 数学" },
    { id: 22, name: "Geography", nameZh: "地理" },
    { id: 23, name: "History", nameZh: "历史" },
    { id: 25, name: "Art", nameZh: "艺术" },
    { id: 27, name: "Animals", nameZh: "动物" },
  ],
  jeopardy: [
    { name: "Literature", nameZh: "文学" },
    { name: "Science", nameZh: "科学" },
    { name: "History", nameZh: "历史" },
    { name: "Geography", nameZh: "地理" },
    { name: "Art & Architecture", nameZh: "艺术与建筑" },
    { name: "Mythology", nameZh: "神话" },
    { name: "Sports", nameZh: "体育" },
    { name: "Pop Culture", nameZh: "流行文化" },
  ],
} as const;

// ============ 类型导出 ============

export type SearchPromiseResult<T> = PromiseSettledResult<T>;
export type { z };

// ============ 使用示例 ============

/**
 * 示例用法:
 *
 * ```typescript
 * import { createOnlineQuestionBank } from './OnlineQuestionBank';
 *
 * const questionBank = createOnlineQuestionBank({
 *   enableOnlineSearch: true,
 *   cacheResults: true,
 *   preferredSources: ['opentdb', 'jeopardy'],
 * });
 *
 * // 搜索单个问题
 * const result = await questionBank.search('什么是人工智能？');
 * console.log(result);
 *
 * // 批量搜索
 * const questions = [
 *   'Python 是什么？',
 *   '机器学习的基本算法有哪些？',
 * ];
 * const results = await questionBank.searchBatch(questions);
 * ```
 */
