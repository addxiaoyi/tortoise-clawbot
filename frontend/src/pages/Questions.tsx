import { useState } from 'react';

export default function Questions() {
  const [questions] = useState([
    { id: 1, type: '选择题', difficulty: '简单', count: 50, subject: '数学' },
    { id: 2, type: '填空题', difficulty: '中等', count: 30, subject: '物理' },
    { id: 3, type: '解答题', difficulty: '困难', count: 20, subject: '化学' },
  ]);

  const [selectedDifficulty, setSelectedDifficulty] = useState('all');

  const filteredQuestions = selectedDifficulty === 'all'
    ? questions
    : questions.filter(q => q.difficulty === selectedDifficulty);

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">题库管理</h1>
        <button className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
          添加题目
        </button>
      </div>

      <div className="flex gap-2 mb-6">
        {['all', '简单', '中等', '困难'].map((level) => (
          <button
            key={level}
            onClick={() => setSelectedDifficulty(level)}
            className={`px-4 py-2 rounded-lg ${
              selectedDifficulty === level
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            {level === 'all' ? '全部' : level}
          </button>
        ))}
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {filteredQuestions.map((q) => (
          <div key={q.id} className="bg-white rounded-lg shadow p-4">
            <div className="flex justify-between items-start mb-2">
              <span className="text-lg font-semibold">{q.type}</span>
              <span className={`px-2 py-1 rounded text-xs ${
                q.difficulty === '简单' ? 'bg-green-100 text-green-800' :
                q.difficulty === '中等' ? 'bg-yellow-100 text-yellow-800' :
                'bg-red-100 text-red-800'
              }`}>
                {q.difficulty}
              </span>
            </div>
            <div className="text-sm text-gray-500 mb-2">
              科目: {q.subject} | 数量: {q.count}题
            </div>
            <div className="flex gap-2">
              <button className="flex-1 px-3 py-2 bg-blue-100 text-blue-600 rounded hover:bg-blue-200">
                练习
              </button>
              <button className="flex-1 px-3 py-2 bg-gray-100 rounded hover:bg-gray-200">
                编辑
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
