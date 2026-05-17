import { useState } from 'react';

export default function Exams() {
  const [exams] = useState([
    { id: 1, title: '高等数学期末考试', date: '2026-01-15', status: '已完成', score: 92 },
    { id: 2, title: '大学物理实验', date: '2026-01-20', status: '进行中', score: null },
    { id: 3, title: '计算机网络', date: '2026-02-01', status: '未开始', score: null },
  ]);

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">考试管理</h1>
        <button className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
          创建考试
        </button>
      </div>

      <div className="bg-white rounded-lg shadow">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">考试名称</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">日期</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">分数</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {exams.map((exam) => (
              <tr key={exam.id}>
                <td className="px-6 py-4 whitespace-nowrap">{exam.title}</td>
                <td className="px-6 py-4 whitespace-nowrap">{exam.date}</td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`px-2 py-1 rounded-full text-xs ${
                    exam.status === '已完成' ? 'bg-green-100 text-green-800' :
                    exam.status === '进行中' ? 'bg-yellow-100 text-yellow-800' :
                    'bg-gray-100 text-gray-800'
                  }`}>
                    {exam.status}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  {exam.score !== null ? `${exam.score}分` : '-'}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <button className="text-blue-600 hover:text-blue-800 mr-3">查看</button>
                  <button className="text-red-600 hover:text-red-800">删除</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
