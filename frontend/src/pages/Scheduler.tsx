import { useState } from 'react';

export default function Scheduler() {
  const [schedules] = useState([
    { id: 1, title: '每日学习计划', time: '08:00', days: ['周一', '周二', '周三', '周四', '周五'] },
    { id: 2, title: '晚间复习', time: '20:00', days: ['周一', '周三', '周五'] },
    { id: 3, title: '周末练习', time: '10:00', days: ['周六', '周日'] },
  ]);

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">学习计划</h1>
        <button className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
          添加计划
        </button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {schedules.map((schedule) => (
          <div key={schedule.id} className="bg-white rounded-lg shadow p-4">
            <div className="flex justify-between items-start mb-4">
              <h3 className="font-semibold text-lg">{schedule.title}</h3>
              <span className="text-blue-600 font-medium">{schedule.time}</span>
            </div>
            <div className="flex flex-wrap gap-2">
              {['周一', '周二', '周三', '周四', '周五', '周六', '周日'].map((day) => (
                <span
                  key={day}
                  className={`px-2 py-1 rounded text-xs ${
                    schedule.days.includes(day)
                      ? 'bg-blue-100 text-blue-800'
                      : 'bg-gray-100 text-gray-400'
                  }`}
                >
                  {day}
                </span>
              ))}
            </div>
            <div className="mt-4 flex gap-2">
              <button className="flex-1 px-3 py-2 bg-gray-100 rounded hover:bg-gray-200">
                编辑
              </button>
              <button className="flex-1 px-3 py-2 bg-red-100 text-red-600 rounded hover:bg-red-200">
                删除
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
