#!/usr/bin/env node

import { execSync } from 'child_process';
import { readFileSync, writeFileSync } from 'fs';
import { join, dirname } from 'path';

// 读取 package.json 文件
function readPackageJson(dir) {
  const packageJsonPath = join(dir, 'package.json');
  return JSON.parse(readFileSync(packageJsonPath, 'utf8'));
}

// 写入 package.json 文件
function writePackageJson(dir, data) {
  const packageJsonPath = join(dir, 'package.json');
  writeFileSync(packageJsonPath, JSON.stringify(data, null, 2));
}

// 执行命令（带重试机制）
function runCommand(command, cwd, maxRetries = 3) {
  let retries = 0;
  while (retries < maxRetries) {
    try {
      console.log(`执行命令: ${command}`);
      execSync(command, { cwd, stdio: 'inherit' });
      return true;
    } catch (error) {
      retries++;
      if (retries >= maxRetries) {
        console.error(`命令执行失败，已尝试 ${maxRetries} 次: ${command}`);
        console.error(error.message);
        return false;
      }
      console.warn(`命令执行失败，正在重试（${retries}/${maxRetries}）: ${command}`);
      // 等待一段时间后重试
      setTimeout(() => {}, 2000);
    }
  }
  return false;
}

// 更新依赖
function updateDependencies(dir, manager) {
  console.log(`\n=== 更新 ${dir} 目录的依赖 ===`);
  
  // 检查目录是否存在
  try {
    readFileSync(join(dir, 'package.json'));
  } catch (error) {
    console.error(`错误: 无法找到 ${dir} 目录的 package.json 文件`);
    return;
  }
  
  // 更新依赖
  let success = true;
  if (manager === 'pnpm') {
    success = runCommand('pnpm update', dir);
  } else if (manager === 'npm') {
    success = runCommand('npm update', dir);
  }
  
  // 安装依赖
  if (success) {
    if (manager === 'pnpm') {
      success = runCommand('pnpm install', dir);
    } else if (manager === 'npm') {
      success = runCommand('npm install', dir);
    }
  }
  
  // 构建项目（如果有构建脚本）
  if (success) {
    const packageJson = readPackageJson(dir);
    if (packageJson.scripts && packageJson.scripts.build) {
      console.log('\n=== 构建项目 ===');
      if (manager === 'pnpm') {
        success = runCommand('pnpm run build', dir);
      } else if (manager === 'npm') {
        success = runCommand('npm run build', dir);
      }
    }
  }
  
  if (success) {
    console.log(`\n=== ${dir} 目录的依赖更新完成 ===`);
  } else {
    console.error(`\n=== ${dir} 目录的依赖更新失败 ===`);
  }
}

// 主函数
function main() {
  console.log('=== 开始更新项目依赖 ===');
  
  // 更新 tortoise 目录的依赖（使用 npm）
  updateDependencies('d:\\qwq\\nanobot\\tortoise', 'npm');
  
  // 更新 openclaw-main 目录的依赖（使用 pnpm）
  updateDependencies('d:\\qwq\\nanobot\\openclaw-main', 'pnpm');
  
  console.log('\n=== 依赖更新任务完成 ===');
}

// 运行主函数
main();
