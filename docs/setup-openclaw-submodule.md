# 将 `openclaw-main` 配置为 Git 子模块（可选）

把上游 [OpenClaw](https://github.com/openclaw/openclaw) 以 **子模块** 形式放在 `openclaw-main/`，便于：

- `git submodule update --remote openclaw-main` 跟踪上游更新  
- 减小本仓库与上游的拷贝漂移  

## 风险与前提

- **会删除现有 `openclaw-main/` 目录**（脚本强制要求你先备份本地改动）。  
- 必须在 **Git 仓库根目录**（含 `.git`）执行。  
- 执行前请 **提交或暂存** 本仓库其它修改，避免与子模块提交混在一起难以回滚。

## 一键初始化（推荐读完上文后再做）

1. 备份：若 `openclaw-main` 内有未同步的 fork 修改，请先 push 到远程或复制到别处。  
2. 设置确认环境变量（**必须**）：  
   - Bash：`export TOHELP_CONFIRM_SUBMODULE_INIT=1`  
   - PowerShell：`$env:TOHELP_CONFIRM_SUBMODULE_INIT = '1'`  
3. 在仓库根目录执行：  
   - **Windows**：`powershell -ExecutionPolicy Bypass -File scripts/init-openclaw-submodule.ps1`  
   - **Unix**：`bash scripts/init-openclaw-submodule.sh`  

4. 安装上游依赖（按 OpenClaw 文档，一般为在 `openclaw-main` 内 `pnpm install`）。  
5. **`npm run doctor`** 应仍全部 `ok`。

## 手动步骤（脚本失败时）

```bash
# 在仓库根目录，已备份 openclaw-main
git rm -rf openclaw-main   # 若该目录曾被提交为普通目录
rm -rf openclaw-main       # 若未纳入版本控制则直接删除
git submodule add https://github.com/openclaw/openclaw.git openclaw-main
git submodule update --init --recursive
git add .gitmodules openclaw-main
git commit -m "chore: vendor OpenClaw as submodule openclaw-main"
```

若提示 `already exists in the index`，先执行 `git rm -rf openclaw-main` 再 `git submodule add`。

## 日常更新上游

```bash
git submodule update --remote openclaw-main
cd openclaw-main && git status   # 确认指向的 commit
cd .. && git add openclaw-main && git commit -m "chore: bump openclaw submodule"
```

或在父仓库根目录：

```bash
git -C openclaw-main fetch origin
git -C openclaw-main checkout main   # 或所需分支/标签
git add openclaw-main
git commit -m "chore: bump openclaw submodule"
```

## CI 注意

GitHub Actions 默认 **浅克隆** 可能不拉子模块。若使用子模块，在 workflow 的 checkout 步骤增加：

```yaml
- uses: actions/checkout@v4
  with:
    submodules: recursive
```

当前本仓库 CI **未假设** 一定存在子模块；若你启用子模块并希望在 CI 中跑依赖 `openclaw-main` 的步骤，请按上式调整。
