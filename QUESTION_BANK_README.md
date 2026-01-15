# 📚 在线题库联网搜索模块

多账号自动考试系统的题库联网搜索模块，支持多个开源题库 API。

## 🎯 功能特性

### 1. 在线题库搜索
- **Open Trivia Database** - 免费综合知识题库
- **JService Jeopardy** - 知识问答题库
- **Quiz API** - 综合题库

### 2. 本地题库管理
- 导入/导出题库（JSON 格式）
- 题目分类（科目、难度、类型）
- 搜索历史记录

### 3. 智能搜索策略
- 优先搜索本地题库
- 自动联网搜索
- 结果缓存
- 置信度评估

## 📦 模块结构

```
src/
├── OnlineQuestionBank.ts    # 在线题库搜索
├── ExamQuestionBank.ts      # 考试题库集成
└── index.ts                # 统一导出
```

## 🚀 快速开始

### 安装依赖

确保已安装 zod（用于类型验证）:

```bash
npm install zod
```

### 基本使用

```typescript
import { createExamQuestionBank } from "./src/index";

// 创建题库实例
const examBank = createExamQuestionBank({
  enableOnlineSearch: true,
  cacheResults: true,
});

// 智能搜索答案
const result = await examBank.smartSearch("什么是人工智能?");

console.log(result.answer);      // 答案
console.log(result.confidence);  // 置信度 (0-1)
console.log(result.source);     // 来源 (local/online)
```

### 导入本地题库

```typescript
// 从 JSON 导入
examBank.importFromJSON("math", [
  {
    id: "q1",
    type: "single",
    question: "1 + 1 = ?",
    options: [
      { key: "A", value: "1" },
      { key: "B", value: "2" },
      { key: "C", value: "3" },
      { key: "D", value: "4" },
    ],
    answer: "B",
    subject: "数学",
    difficulty: 1,
  },
]);

// 导出题库
const json = examBank.exportToJSON("math");
```

### 批量搜索

```typescript
const questions = [
  "问题1",
  "问题2",
  "问题3",
];

const results = await examBank.smartSearchBatch(questions, {
  concurrency: 5,  // 并发数
});

// 获取统计
const stats = examBank.getStatistics();
console.log(stats.totalQuestions);
```

## 🔧 配置选项

```typescript
interface QuestionBankConfig {
  enableOnlineSearch: boolean;  // 启用联网搜索
  searchTimeout: number;        // 搜索超时（毫秒）
  fallbackToLocal: boolean;     // 本地未找到时联网搜索
  cacheResults: boolean;        // 缓存搜索结果
  cacheExpiry: number;          // 缓存过期时间（毫秒）
  preferredSources: string[];   // 优先使用的题库
}
```

## 🌐 开源题库资源

| 题库 | 网址 | 类别 | API | 许可 |
|------|------|------|-----|------|
| Open Trivia DB | https://opentdb.com | 综合 | ✅ | CC BY-SA 4.0 |
| JService Jeopardy | https://jservice.io | 知识问答 | ✅ | Public Domain |
| OpenBook QA | https://allenai.org | 科学常识 | ❌ | CC BY-SA |
| CommonsenseQA | https://tau-nlp.org | 常识推理 | ❌ | Apache 2.0 |
| MedQA | https://github.com | 医学 | ❌ | Apache 2.0 |

## 📝 API 参考

### OnlineQuestionBank

```typescript
// 创建实例
const bank = createOnlineQuestionBank(config);

// 搜索题目
const result = await bank.search("question text");

// 批量搜索
const results = await bank.searchBatch(["q1", "q2", "q3"]);

// 清除缓存
bank.clearCache();
```

### ExamQuestionBank

```typescript
// 创建实例
const bank = createExamQuestionBank(config);

// 智能搜索
const result = await bank.smartSearch("question", options);

// 批量智能搜索
const results = await bank.smartSearchBatch(questions, options);

// 本地题库操作
bank.addLocalBank(id, bank);
bank.importFromJSON(id, json);
const json = bank.exportToJSON(id);
bank.addQuestion(bankId, question);

// 搜索本地
const localResults = bank.searchLocal("question", bankId);

// 获取答案
const answer = bank.getAnswer(questionId, bankId);

// 统计信息
const stats = bank.getStatistics(bankId);

// 列出题库
const banks = bank.listBanks();
```

## 🎨 题库类别 (Open Trivia DB)

```typescript
import { QUESTION_CATEGORIES } from "./src/index";

// 基础知识
QUESTION_CATEGORIES.opentdb[0]  // { id: 9, name: "General Knowledge" }

// 科学
QUESTION_CATEGORIES.opentdb[6]  // { id: 17, name: "Science & Nature" }

// 历史
QUESTION_CATEGORIES.opentdb[9]  // { id: 23, name: "History" }
```

## 📊 类型定义

```typescript
interface Question {
  id: string;
  type: "single" | "multiple" | "true_false" | "fill" | "essay";
  question: string;
  options?: { key: string; value: string }[];
  answer?: string | string[];
  explanation?: string;
  source?: string;
  tags?: string[];
}

interface LocalQuestion extends Question {
  subject: string;
  grade?: string;
  difficulty: 1 | 2 | 3;
  frequency?: number;
  lastUsed?: number;
}

interface SearchResult {
  question: string;
  matchedQuestions: Question[];
  confidence: number;
  source: string;
}
```

## 🔄 工作流程

```
┌─────────────┐
│  用户问题    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 本地题库搜索 │
└──────┬──────┘
       │
  ┌────┴────┐
  │ 找到?   │
  └────┬────┘
   Yes │ No
   ┌───┴───┐
   ▼       ▼
┌──────┐ ┌─────────────┐
│返回本地│ │ 联网搜索   │
│答案   │ └──────┬──────┘
└──────┘        │
          ┌─────┴─────┐
          │ 找到?    │
          └─────┬─────┘
           Yes │ No
           ┌───┴───┐
           ▼       ▼
      ┌────────┐ ┌────────┐
      │返回联网│ │返回未找到│
      │答案   │ └────────┘
      └────────┘
```

## 📄 许可

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！
