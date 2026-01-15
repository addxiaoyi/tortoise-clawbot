/**
 * 题库模块使用示例
 */

import {
  createOnlineQuestionBank,
  createExamQuestionBank,
  OPEN_SOURCE_QUESTION_BANKS,
} from "../src/index.ts";

// ============ 示例 1: 基础在线题库搜索 ============

async function basicOnlineSearch() {
  console.log("=== 示例 1: 在线题库搜索 ===\n");

  const onlineBank = createOnlineQuestionBank({
    enableOnlineSearch: true,
    searchTimeout: 15000,
    cacheResults: true,
  });

  // 搜索一个问题
  const result = await onlineBank.search("What is artificial intelligence?");

  console.log("原始问题:", result.question);
  console.log("匹配数量:", result.matchedQuestions.length);
  console.log("置信度:", result.confidence);
  console.log("来源:", result.source);

  if (result.matchedQuestions.length > 0) {
    console.log("\n最佳匹配:");
    console.log("  题目:", result.matchedQuestions[0].question);
    console.log("  答案:", result.matchedQuestions[0].answer);
    console.log("  来源:", result.matchedQuestions[0].source);
  }
}

// ============ 示例 2: 考试题库集成 ============

async function examBankIntegration() {
  console.log("\n=== 示例 2: 考试题库集成 ===\n");

  const examBank = createExamQuestionBank({
    enableOnlineSearch: true,
    cacheResults: true,
  });

  // 导入本地题库
  const importedCount = examBank.importFromJSON("math", [
    {
      name: "数学题库",
      questions: [
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
        {
          id: "q2",
          type: "single",
          question: "2 + 2 = ?",
          options: [
            { key: "A", value: "2" },
            { key: "B", value: "3" },
            { key: "C", value: "4" },
            { key: "D", value: "5" },
          ],
          answer: "C",
          subject: "数学",
          difficulty: 1,
        },
        {
          id: "q3",
          type: "single",
          question: "什么是机器学习?",
          options: [
            { key: "A", value: "一种编程语言" },
            { key: "B", value: "让计算机从数据中学习的技术" },
            { key: "C", value: "一种硬件设备" },
            { key: "D", value: "一个游戏" },
          ],
          answer: "B",
          subject: "计算机科学",
          difficulty: 2,
        },
      ],
    },
  ]);

  console.log(`导入 ${importedCount} 道题目\n`);

  // 智能搜索
  const result1 = await examBank.smartSearch("1+1等于多少");
  console.log("搜索 '1+1等于多少':");
  console.log("  答案:", result1.answer);
  console.log("  来源:", result1.source);
  console.log("  置信度:", result1.confidence);

  const result2 = await examBank.smartSearch("机器学习是什么?");
  console.log("\n搜索 '机器学习是什么?':");
  console.log("  答案:", result2.answer);
  console.log("  来源:", result2.source);
  console.log("  置信度:", result2.confidence);

  // 获取统计
  const stats = examBank.getStatistics("math");
  console.log("\n题库统计:");
  console.log("  总题数:", stats.totalQuestions);
  console.log("  按类型:", stats.byType);
  console.log("  按科目:", stats.bySubject);
  console.log("  按难度:", stats.byDifficulty);
}

// ============ 示例 3: 批量搜索 ============

async function batchSearch() {
  console.log("\n=== 示例 3: 批量搜索 ===\n");

  const examBank = createExamQuestionBank({
    enableOnlineSearch: true,
  });

  const questions = [
    "What is Python?",
    "What is machine learning?",
    "What is the capital of France?",
  ];

  const results = await examBank.smartSearchBatch(questions, {
    concurrency: 3,
  });

  results.forEach((result, index) => {
    console.log(`问题 ${index + 1}:`, questions[index]);
    console.log("  答案:", result.answer || "(未找到)");
    console.log("  来源:", result.source);
    console.log("  置信度:", result.confidence.toFixed(2));
    console.log();
  });
}

// ============ 示例 4: 开源题库资源 ============

function listOpenSourceBanks() {
  console.log("\n=== 示例 4: 开源题库资源 ===\n");

  console.log("可用的开源题库:\n");

  OPEN_SOURCE_QUESTION_BANKS.forEach((bank) => {
    console.log(`📚 ${bank.name}`);
    console.log(`   网址: ${bank.url}`);
    console.log(`   描述: ${bank.description}`);
    console.log(`   类别: ${bank.categories.join(", ")}`);
    console.log(`   API: ${bank.api ? "✅ 支持" : "❌ 不支持"}`);
    console.log(`   许可: ${bank.license}`);
    console.log();
  });
}

// ============ 运行所有示例 ============

async function main() {
  console.log("🎯 题库模块演示\n");
  console.log("=".repeat(50));

  try {
    await basicOnlineSearch();
    await examBankIntegration();
    await batchSearch();
    listOpenSourceBanks();

    console.log("=".repeat(50));
    console.log("✅ 所有示例执行完成!");
  } catch (error) {
    console.error("❌ 执行出错:", error);
  }
}

main();
