# 依赖更新脚本

## 功能说明

本脚本用于自动更新项目的依赖，包括：

- `tortoise` 目录（前端桌面应用）
- `openclaw-main` 目录（后端核心）

## 使用方法

### 运行脚本

在项目根目录执行：

```bash
node scripts/update-dependencies.mjs
```

### 脚本执行流程

1. **更新 tortoise 目录**：
   - 运行 `npm update` 更新依赖
   - 运行 `npm install` 安装依赖
   - 运行 `npm run build` 构建项目

2. **更新 openclaw-main 目录**：
   - 运行 `pnpm update` 更新依赖
   - 运行 `pnpm install` 安装依赖
   - 运行 `pnpm run build` 构建项目（需要 WSL 环境）

## 注意事项

1. **环境要求**：
   - Node.js 22.12.0 或更高版本
   - npm 9.0.0 或更高版本
   - pnpm 10.0.0 或更高版本
   - 构建 openclaw-main 需要 WSL（Windows Subsystem for Linux）环境

2. **网络要求**：
   - 脚本需要访问 npm 注册表和 GitHub 仓库
   - 建议使用稳定的网络连接

3. **错误处理**：
   - 脚本具有重试机制，当命令执行失败时会自动重试最多 3 次
   - 如果构建失败，脚本会继续执行其他任务

## 自动化建议

### 使用 GitHub Actions 自动化

在 `.github/workflows/` 目录创建一个工作流文件，例如 `dependencies.yml`：

```yaml
name: Update Dependencies

on:
  schedule:
    - cron: '0 0 * * 0'  # 每周日执行
  workflow_dispatch:  # 手动触发

jobs:
  update-dependencies:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
          cache: 'pnpm'
      - run: npm install -g pnpm
      - run: node scripts/update-dependencies.mjs
```

### 使用系统定时任务

#### Windows（任务计划程序）

1. 打开「任务计划程序」
2. 创建新任务
3. 设置触发器（例如每周执行）
4. 设置操作：`node.exe`，参数：`scripts/update-dependencies.mjs`，起始于：项目根目录

#### Linux / macOS（cron）

在 crontab 中添加：

```bash
0 0 * * 0 cd /path/to/nanobot && node scripts/update-dependencies.mjs
```

## 依赖更新策略

1. **安全更新**：优先更新具有安全补丁的依赖
2. **补丁更新**：更新补丁版本号（例如 1.0.0 → 1.0.1）
3. **次要更新**：更新次要版本号（例如 1.0.0 → 1.1.0）
4. **主要更新**：谨慎更新主要版本号（例如 1.0.0 → 2.0.0），可能包含破坏性变更

## 故障排除

### 常见错误

1. **网络连接失败**：
   - 检查网络连接
   - 尝试使用代理
   - 稍后再试

2. **构建失败**：
   - 检查环境要求
   - 确保 WSL 已正确安装（仅 openclaw-main）
   - 查看详细错误信息

3. **依赖冲突**：
   - 运行 `npm ls` 或 `pnpm ls` 查看依赖树
   - 手动解决冲突

## 版本历史

- **1.0.0**：初始版本，支持基本依赖更新功能
- **1.1.0**：添加错误处理和重试机制
- **1.2.0**：优化执行顺序，先更新前端依赖
