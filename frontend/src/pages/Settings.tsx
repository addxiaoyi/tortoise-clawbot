export default function Settings() {
  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">设置</h1>

      <div className="space-y-6">
        {/* 基本设置 */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">基本设置</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                显示名称
              </label>
              <input
                type="text"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="输入显示名称"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                邮箱
              </label>
              <input
                type="email"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="输入邮箱地址"
              />
            </div>
          </div>
        </div>

        {/* 通知设置 */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">通知设置</h2>
          <div className="space-y-3">
            {[
              { label: '邮件通知', checked: true },
              { label: '桌面通知', checked: true },
              { label: '短信通知', checked: false },
            ].map((item) => (
              <label key={item.label} className="flex items-center">
                <input
                  type="checkbox"
                  defaultChecked={item.checked}
                  className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                />
                <span className="ml-2 text-gray-700">{item.label}</span>
              </label>
            ))}
          </div>
        </div>

        {/* 外观设置 */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">外观设置</h2>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              主题
            </label>
            <select className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500">
              <option value="light">浅色模式</option>
              <option value="dark">深色模式</option>
              <option value="system">跟随系统</option>
            </select>
          </div>
        </div>

        {/* 保存按钮 */}
        <div className="flex justify-end">
          <button className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
            保存设置
          </button>
        </div>
      </div>
    </div>
  );
}
