# start-dev.ps1 - OpenClaw Windows Developer Launch Script

if (-not (Test-Path "$PSScriptRoot\openclaw-main\package.json")) {
    Write-Host "Error: openclaw-main\ missing. Run: npm run doctor" -ForegroundColor Red
    exit 1
}

# 1. 设置配置路径 (指向仓库内的配置文件)
$env:OPENCLAW_CONFIG_PATH = "$PSScriptRoot\.openclaw-dev\openclaw.json"

# 2. 加载 .env 文件 (如果存在)
$EnvFile = "$PSScriptRoot\.env"
if (Test-Path $EnvFile) {
    Write-Host "Loading environment from .env..." -ForegroundColor Gray
    Get-Content $EnvFile | ForEach-Object {
        if ($_ -match '^\s*([^#=]+)\s*=\s*(.*)$') {
            $name = $matches[1]
            $value = $matches[2]
            [Environment]::SetEnvironmentVariable($name, $value, "Process")
        }
    }
}

# 2b. 网关 token：openclaw.json 使用 ${OPENCLAW_GATEWAY_TOKEN}，未设置则无法解析配置
if (-not $env:OPENCLAW_GATEWAY_TOKEN -or $env:OPENCLAW_GATEWAY_TOKEN.Trim() -eq '') {
    Write-Host "Warning: OPENCLAW_GATEWAY_TOKEN not set. Using dev-only placeholder; set in .env for real use." -ForegroundColor Yellow
    $env:OPENCLAW_GATEWAY_TOKEN = 'dev-only-tohelp-gateway-token-change-me'
}

# 3. 跳过不必要的初始化 (加快开发启动)
if (-not $env:OPENCLAW_SKIP_CHANNELS) { $env:OPENCLAW_SKIP_CHANNELS = '1' }
if (-not $env:CLAWDBOT_SKIP_CHANNELS) { $env:CLAWDBOT_SKIP_CHANNELS = '1' }
$env:OPENCLAW_NO_VERSION_CHECK = '1'

# 4. 设置默认 API Base URL (如果未在 .env 中设置)
if (-not $env:ANTHROPIC_BASE_URL) { $env:ANTHROPIC_BASE_URL = "https://api.anthropic.com" }
if (-not $env:OPENAI_BASE_URL) { $env:OPENAI_BASE_URL = "https://api.openai.com/v1" }
if (-not $env:GOOGLE_BASE_URL) { $env:GOOGLE_BASE_URL = "https://generativelanguage.googleapis.com" }
if (-not $env:CUSTOM_LOCAL_BASE_URL) { $env:CUSTOM_LOCAL_BASE_URL = "http://localhost:11434/v1" }
if (-not $env:OPENROUTER_BASE_URL) { $env:OPENROUTER_BASE_URL = "https://openrouter.ai/api/v1" }
if (-not $env:GROQ_BASE_URL) { $env:GROQ_BASE_URL = "https://api.groq.com/openai/v1" }
if (-not $env:PERPLEXITY_BASE_URL) { $env:PERPLEXITY_BASE_URL = "https://api.perplexity.ai" }
if (-not $env:MISTRAL_BASE_URL) { $env:MISTRAL_BASE_URL = "https://api.mistral.ai/v1" }
if (-not $env:TOGETHER_BASE_URL) { $env:TOGETHER_BASE_URL = "https://api.together.xyz/v1" }
if (-not $env:DEEPSEEK_BASE_URL) { $env:DEEPSEEK_BASE_URL = "https://api.deepseek.com" }
if (-not $env:CUSTOM_2_BASE_URL) { $env:CUSTOM_2_BASE_URL = "http://localhost:1234/v1" }

# 5. 检查 API Key (如果未设置，提示用户或设置临时占位符)
if (-not $env:ANTHROPIC_API_KEY) {
    Write-Host "Warning: ANTHROPIC_API_KEY is not set. Using placeholder." -ForegroundColor Yellow
    $env:ANTHROPIC_API_KEY = "sk-placeholder-anthropic"
}
if (-not $env:OPENAI_API_KEY) {
    Write-Host "Warning: OPENAI_API_KEY is not set. Using placeholder." -ForegroundColor Yellow
    $env:OPENAI_API_KEY = "sk-placeholder-openai"
}
if (-not $env:GOOGLE_API_KEY) {
    Write-Host "Warning: GOOGLE_API_KEY is not set. Using placeholder." -ForegroundColor Yellow
    $env:GOOGLE_API_KEY = "sk-placeholder-google"
}
if (-not $env:CUSTOM_LOCAL_API_KEY) {
    Write-Host "Warning: CUSTOM_LOCAL_API_KEY is not set. Using default 'ollama'." -ForegroundColor Yellow
    $env:CUSTOM_LOCAL_API_KEY = "ollama"
}
if (-not $env:OPENROUTER_API_KEY) {
    Write-Host "Warning: OPENROUTER_API_KEY is not set." -ForegroundColor Yellow
    $env:OPENROUTER_API_KEY = "sk-placeholder-openrouter"
}
if (-not $env:GROQ_API_KEY) {
    Write-Host "Warning: GROQ_API_KEY is not set." -ForegroundColor Yellow
    $env:GROQ_API_KEY = "gsk-placeholder-groq"
}
if (-not $env:PERPLEXITY_API_KEY) {
    Write-Host "Warning: PERPLEXITY_API_KEY is not set." -ForegroundColor Yellow
    $env:PERPLEXITY_API_KEY = "pplx-placeholder-perplexity"
}
if (-not $env:MISTRAL_API_KEY) {
    Write-Host "Warning: MISTRAL_API_KEY is not set." -ForegroundColor Yellow
    $env:MISTRAL_API_KEY = "sk-placeholder-mistral"
}
if (-not $env:TOGETHER_API_KEY) {
    Write-Host "Warning: TOGETHER_API_KEY is not set." -ForegroundColor Yellow
    $env:TOGETHER_API_KEY = "sk-placeholder-together"
}
if (-not $env:DEEPSEEK_API_KEY) {
    Write-Host "Warning: DEEPSEEK_API_KEY is not set." -ForegroundColor Yellow
    $env:DEEPSEEK_API_KEY = "sk-placeholder-deepseek"
}
if (-not $env:CUSTOM_2_API_KEY) {
    Write-Host "Warning: CUSTOM_2_API_KEY is not set." -ForegroundColor Yellow
    $env:CUSTOM_2_API_KEY = "sk-placeholder-custom-2"
}

# 4. 启动网关
Write-Host "Starting OpenClaw Gateway..." -ForegroundColor Cyan
Write-Host "Config: $env:OPENCLAW_CONFIG_PATH" -ForegroundColor Gray

# 必须进入项目根目录，否则 pnpm 找不到 package.json
Set-Location "$PSScriptRoot\openclaw-main"
node "scripts\run-node.mjs" --dev gateway
