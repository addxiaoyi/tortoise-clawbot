# 插件市场优化 Spec

## Why
当前插件市场存在两个关键问题：
1. 插件参数定制功能不完善，用户无法直观地调整插件参数
2. 插件描述质量较低，使用简单翻译，缺乏详细信息帮助用户决策

## What Changes
- 实现完善的插件参数定制界面
- 增强插件描述，提供详细、准确、用户友好的信息

## Impact
- Affected specs: 插件市场用户体验
- Affected code: `src/components/EmbeddedPluginDetail.tsx`, `src/pages/SkillsPage.tsx`

## ADDED Requirements
### Requirement: 插件参数定制界面
系统应提供直观的参数定制界面：

#### Scenario: 参数类型支持
- **THEN** 支持文本输入（string）
- **THEN** 支持数字输入（number）带范围验证
- **THEN** 支持布尔开关（boolean）
- **THEN** 支持下拉选择（select）

#### Scenario: 参数验证
- **WHEN** 用户输入参数值
- **THEN** 实时验证输入有效性
- **THEN** 显示验证错误提示
- **THEN** 禁用无效输入的保存

#### Scenario: 参数描述
- **WHEN** 用户查看参数
- **THEN** 显示参数名称、描述、默认值
- **THEN** 显示可接受的值范围

### Requirement: 插件描述增强
插件描述应包含以下信息：

#### Scenario: 描述内容
- **THEN** 功能概述：插件的核心功能
- **THEN** 使用场景：适合什么情况下使用
- **THEN** 兼容性：支持哪些核心 Agent
- **THEN** 安装要求：需要什么依赖
- **THEN** 使用说明：如何配置和使用

#### Scenario: 描述来源
- **THEN** 通过联网搜索获取真实信息
- **THEN** 编写原创、准确、用户友好的描述
- **THEN** 避免纯机器翻译

## MODIFIED Requirements
- 扩展现有 `SKILL_COPY` 数据结构
- 增强 `EmbeddedPluginDetail` 组件

## REMOVED Requirements
- 无
