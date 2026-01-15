/**
 * 考试题库集成模块
 * 结合本地题库和在线搜索，提供智能答题支持
 */

import { OnlineQuestionBank, Question, SearchResult, createOnlineQuestionBank } from "./OnlineQuestionBank";

// ============ 本地题库存储 ============

export interface LocalQuestion extends Question {
  subject: string;
  grade?: string;
  difficulty: 1 | 2 | 3;
  frequency?: number;
  lastUsed?: number;
}

export interface LocalQuestionBank {
  name: string;
  questions: LocalQuestion[];
  lastUpdated: number;
}

// ============ 题库管理器 ============

export class ExamQuestionBank {
  private onlineBank: OnlineQuestionBank;
  private localBanks = new Map<string, LocalQuestionBank>();
  private defaultLocalBank: LocalQuestionBank;

  constructor(config?: Parameters<typeof createOnlineQuestionBank>[0]) {
    this.onlineBank = createOnlineQuestionBank(config);
    this.defaultLocalBank = {
      name: "默认题库",
      questions: [],
      lastUpdated: Date.now(),
    };
    this.localBanks.set("default", this.defaultLocalBank);
  }

  // ============ 本地题库操作 ============

  /**
   * 添加本地题库
   */
  addLocalBank(id: string, bank: LocalQuestionBank): void {
    this.localBanks.set(id, bank);
  }

  /**
   * 从 JSON 导入题库
   */
  importFromJSON(id: string, json: string | object): number {
    const data = typeof json === "string" ? JSON.parse(json) : json;
    const questions: LocalQuestion[] = Array.isArray(data)
      ? data
      : data.questions || [];

    this.localBanks.set(id, {
      name: data.name || id,
      questions,
      lastUpdated: Date.now(),
    });

    return questions.length;
  }

  /**
   * 导出题库为 JSON
   */
  exportToJSON(bankId: string = "default"): string {
    const bank = this.localBanks.get(bankId);
    if (!bank) throw new Error(`题库 ${bankId} 不存在`);
    return JSON.stringify(bank, null, 2);
  }

  /**
   * 添加题目到本地题库
   */
  addQuestion(
    bankId: string,
    question: Omit<LocalQuestion, "id" | "frequency" | "lastUsed">
  ): LocalQuestion {
    const bank = this.localBanks.get(bankId) || this.defaultLocalBank;

    const newQuestion: LocalQuestion = {
      ...question,
      id: `local-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      frequency: 0,
      lastUsed: undefined,
    };

    bank.questions.push(newQuestion);
    bank.lastUpdated = Date.now();

    this.localBanks.set(bankId, bank);
    return newQuestion;
  }

  /**
   * 在本地题库搜索
   */
  searchLocal(question: string, bankId?: string): LocalQuestion[] {
    const normalizedQuery = this.normalizeQuestion(question);
    const keywords = this.extractKeywords(question);

    const banks = bankId
      ? [this.localBanks.get(bankId)].filter(Boolean) as LocalQuestionBank[]
      : Array.from(this.localBanks.values());

    const results: (LocalQuestion & { score: number })[] = [];

    for (const bank of banks) {
      for (const q of bank.questions) {
        const qNorm = this.normalizeQuestion(q.question);
        const similarity = this.stringSimilarity(normalizedQuery, qNorm);

        // 检查关键词匹配
        const keywordMatches = keywords.filter((kw) =>
          qNorm.includes(kw) || q.question.includes(kw)
        ).length;

        const score = similarity * 0.7 + (keywordMatches / keywords.length) * 0.3;

        if (score > 0.3 || keywordMatches > 0) {
          results.push({ ...q, score });
        }
      }
    }

    return results
      .sort((a, b) => b.score - a.score)
      .slice(0, 10)
      .map(({ score, ...q }) => q);
  }

  /**
   * 获取题目的答案
   */
  getAnswer(questionId: string, bankId?: string): string | string[] | undefined {
    const banks = bankId
      ? [this.localBanks.get(bankId)].filter(Boolean) as LocalQuestionBank[]
      : Array.from(this.localBanks.values());

    for (const bank of banks) {
      const q = bank.questions.find((q) => q.id === questionId);
      if (q) {
        // 更新使用频率
        q.frequency = (q.frequency || 0) + 1;
        q.lastUsed = Date.now();
        return q.answer;
      }
    }
    return undefined;
  }

  // ============ 智能搜索 ============

  /**
   * 智能搜索 - 优先本地，再联网
   */
  async smartSearch(
    question: string,
    options?: {
      preferLocal?: boolean;
      fallbackOnline?: boolean;
      bankId?: string;
    }
  ): Promise<{
    answer: string | string[];
    confidence: number;
    source: "local" | "online" | "merged";
    question?: LocalQuestion;
    onlineResult?: SearchResult;
  }> {
    const { preferLocal = true, fallbackOnline = true, bankId } = options || {};

    // 1. 先搜索本地题库
    const localResults = this.searchLocal(question, bankId);

    if (localResults.length > 0) {
      const best = localResults[0];
      return {
        answer: best.answer!,
        confidence: Math.min(0.95, 0.5 + localResults[0].score * 0.5),
        source: "local",
        question: best,
      };
    }

    // 2. 联网搜索
    if (fallbackOnline) {
      const onlineResult = await this.onlineBank.search(question);

      if (onlineResult.matchedQuestions.length > 0) {
        const best = onlineResult.matchedQuestions[0];
        return {
          answer: best.answer!,
          confidence: onlineResult.confidence,
          source: "online",
          onlineResult,
        };
      }
    }

    // 3. 都没找到
    return {
      answer: "",
      confidence: 0,
      source: "merged",
    };
  }

  /**
   * 批量智能搜索
   */
  async smartSearchBatch(
    questions: string[],
    options?: {
      preferLocal?: boolean;
      fallbackOnline?: boolean;
      bankId?: string;
      concurrency?: number;
    }
  ): Promise<ReturnType<typeof this.smartSearch>[]> {
    const { concurrency = 5 } = options || {};

    const results: ReturnType<typeof this.smartSearch>[] = [];
    for (let i = 0; i < questions.length; i += concurrency) {
      const batch = questions.slice(i, i + concurrency);
      const batchResults = await Promise.all(
        batch.map((q) => this.smartSearch(q, options))
      );
      results.push(...batchResults);
    }
    return results;
  }

  // ============ 题库统计 ============

  /**
   * 获取题库统计信息
   */
  getStatistics(bankId?: string): {
    totalQuestions: number;
    byType: Record<string, number>;
    bySubject: Record<string, number>;
    byDifficulty: Record<number, number>;
    mostUsed: LocalQuestion[];
  } {
    const banks = bankId
      ? [this.localBanks.get(bankId)].filter(Boolean) as LocalQuestionBank[]
      : Array.from(this.localBanks.values());

    const stats = {
      totalQuestions: 0,
      byType: {} as Record<string, number>,
      bySubject: {} as Record<string, number>,
      byDifficulty: {} as Record<number, number>,
      mostUsed: [] as LocalQuestion[],
    };

    const allQuestions: LocalQuestion[] = [];

    for (const bank of banks) {
      stats.totalQuestions += bank.questions.length;
      allQuestions.push(...bank.questions);

      for (const q of bank.questions) {
        stats.byType[q.type] = (stats.byType[q.type] || 0) + 1;
        stats.bySubject[q.subject] = (stats.bySubject[q.subject] || 0) + 1;
        stats.byDifficulty[q.difficulty] = (stats.byDifficulty[q.difficulty] || 0) + 1;
      }
    }

    stats.mostUsed = allQuestions
      .filter((q) => q.frequency && q.frequency > 0)
      .sort((a, b) => (b.frequency || 0) - (a.frequency || 0))
      .slice(0, 10);

    return stats;
  }

  /**
   * 获取题库列表
   */
  listBanks(): { id: string; name: string; count: number }[] {
    return Array.from(this.localBanks.entries()).map(([id, bank]) => ({
      id,
      name: bank.name,
      count: bank.questions.length,
    }));
  }

  // ============ 辅助方法 ============

  private normalizeQuestion(question: string): string {
    return question
      .toLowerCase()
      .replace(/[^\w\u4e00-\u9fa5]/g, "")
      .trim();
  }

  private extractKeywords(question: string): string[] {
    const cleaned = question.replace(
      /^(以下|请|请问|帮我|帮我查|查找|搜索|找一下|答案是|正确的|下面哪个|下列)/gi,
      ""
    ).replace(/\?{1,}$/, "").trim();

    const keywords: string[] = [];
    const chineseMatch = cleaned.match(/[\u4e00-\u9fa5]{2,}/g);
    if (chineseMatch) keywords.push(...chineseMatch.slice(0, 5));

    return keywords;
  }

  private stringSimilarity(a: string, b: string): number {
    if (a === b) return 1;
    if (a.length === 0 || b.length === 0) return 0;

    const longer = a.length > b.length ? a : b;
    const shorter = a.length > b.length ? b : a;

    if (longer.includes(shorter)) {
      return shorter.length / longer.length;
    }

    const editDistance = this.levenshteinDistance(a, b);
    return (longer.length - editDistance) / longer.length;
  }

  private levenshteinDistance(a: string, b: string): number {
    const matrix: number[][] = [];

    for (let i = 0; i <= b.length; i++) matrix[i] = [i];
    for (let j = 0; j <= a.length; j++) matrix[0][j] = j;

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

  /**
   * 清除在线搜索缓存
   */
  clearOnlineCache(): void {
    this.onlineBank.clearCache();
  }
}

// ============ 工厂函数 ============

export function createExamQuestionBank(
  config?: Parameters<typeof createOnlineQuestionBank>[0]
): ExamQuestionBank {
  return new ExamQuestionBank(config);
}

// ============ 开源题库资源列表 ============

export const OPEN_SOURCE_QUESTION_BANKS = [
  {
    name: "Open Trivia Database",
    url: "https://opentdb.com",
    description: "免费的综合知识题库，支持多类别",
    categories: ["基础知识", "科学", "历史", "地理", "文学", "艺术", "体育"],
    api: true,
    license: "CC BY-SA 4.0",
  },
  {
    name: "JService Jeopardy",
    url: "https://jservice.io",
    description: "Jeopardy 知识问答题库",
    categories: ["文学", "科学", "历史", "地理", "艺术", "神话", "体育"],
    api: true,
    license: "Public Domain",
  },
  {
    name: "OpenBook QA",
    url: "https://allenai.org/datasets/open-book-qa",
    description: "AI2 开放书本问答数据集",
    categories: ["科学常识", "学校科目"],
    api: false,
    license: "CC BY-SA",
  },
  {
    name: "CommonsenseQA",
    url: "https://www.tau-nlp.org/commonsenseqa",
    description: "常识问答数据集",
    categories: ["常识推理"],
    api: false,
    license: "Apache 2.0",
  },
  {
    name: "MedQA (USMLE)",
    url: "https://github.com/jind11/MedQA",
    description: "医学考试题库",
    categories: ["医学", "临床"],
    api: false,
    license: "Apache 2.0",
  },
] as const;

/**
 * 示例用法:
 *
 * ```typescript
 * import { createExamQuestionBank, OPEN_SOURCE_QUESTION_BANKS } from './ExamQuestionBank';
 *
 * // 创建题库实例
 * const questionBank = createExamQuestionBank({
 *   enableOnlineSearch: true,
 *   cacheResults: true,
 * });
 *
 * // 添加本地题库
 * questionBank.importFromJSON('math', [
 *   {
 *     id: 'q1',
 *     type: 'single',
 *     question: '1+1=?',
 *     options: [{ key: 'A', value: '1' }, { key: 'B', value: '2' }],
 *     answer: '2',
 *     subject: '数学',
 *     difficulty: 1,
 *   },
 * ]);
 *
 * // 智能搜索答案
 * const result = await questionBank.smartSearch('1+1等于多少');
 * console.log(result.answer); // '2'
 *
 * // 批量搜索
 * const questions = ['什么是AI?', '机器学习是什么?'];
 * const results = await questionBank.smartSearchBatch(questions);
 *
 * // 获取统计
 * const stats = questionBank.getStatistics();
 * console.log(stats);
 *
 * // 导出题库
 * const json = questionBank.exportToJSON('math');
 * ```
 */
