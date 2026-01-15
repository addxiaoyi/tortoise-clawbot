# Tasks

- [x] Task 1: 创建内容过滤核心模块 content-filter.ts
  - [x] SubTask 1.1: 实现敏感词库数据结构
  - [x] SubTask 1.2: 实现赌博类敏感词检测函数
  - [x] SubTask 1.3: 实现炒股类敏感词检测函数
  - [x] SubTask 1.4: 实现内容过滤主函数
  - [x] SubTask 1.5: 实现过滤策略配置（严格/宽松模式）

- [x] Task 2: 创建敏感词库 keywords.ts
  - [x] SubTask 2.1: 定义赌博类敏感词列表
  - [x] SubTask 2.2: 定义炒股类敏感词列表
  - [x] SubTask 2.3: 提供词库导出接口

- [x] Task 3: 集成内容过滤到应用框架
  - [x] SubTask 3.1: 在应用入口添加定位声明展示
  - [x] SubTask 3.2: 在关于页面添加定位说明
  - [x] SubTask 3.3: 添加用户协议确认弹窗

- [x] Task 4: 添加内容过滤 UI 组件
  - [x] SubTask 4.1: 创建过滤提示 Toast 组件
  - [x] SubTask 4.2: 创建协议确认弹窗组件

# Task Dependencies
- Task 2 依赖 Task 1（需要先有过滤逻辑才能定义词库结构）
- Task 3 依赖 Task 1 和 Task 2
- Task 4 依赖 Task 1
