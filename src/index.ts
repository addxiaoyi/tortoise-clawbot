/**
 * 题库模块索引
 * 统一导出所有题库相关功能
 */

export {
  // 在线题库
  OnlineQuestionBank,
  createOnlineQuestionBank,
  QUESTION_CATEGORIES,
  type Question,
  type SearchResult,
  type QuestionBankConfig,
} from "./OnlineQuestionBank";

export {
  // 考试题库
  ExamQuestionBank,
  createExamQuestionBank,
  OPEN_SOURCE_QUESTION_BANKS,
  type LocalQuestion,
  type LocalQuestionBank,
} from "./ExamQuestionBank";
