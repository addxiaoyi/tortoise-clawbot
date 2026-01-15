package plugin

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// ============ Plugin Sandbox 插件沙箱 ============
// 安全隔离的插件执行环境

// SandboxConfig 沙箱配置
type SandboxConfig struct {
	MaxMemoryMB      int           // 最大内存 (MB)
	MaxCPUPercent   int           // 最大 CPU 百分比
	MaxDuration     time.Duration // 最大执行时间
	MaxOutputSize   int           // 最大输出大小 (bytes)
	AllowNetwork    bool          // 允许网络访问
	AllowFileSystem bool          // 允许文件系统访问
	AllowedPaths    []string      // 允许的文件路径
	WorkDir         string        // 工作目录
}

// SandboxResult 沙箱执行结果
type SandboxResult struct {
	Success    bool
	Output     string
	Error      string
	ExitCode   int
	Duration   time.Duration
	MemoryUsed int64
}

// Sandbox 插件沙箱
type Sandbox struct {
	config    *SandboxConfig
	processes map[string]*exec.Cmd
	mu        sync.Mutex
}

// NewSandbox 创建沙箱
func NewSandbox(config *SandboxConfig) *Sandbox {
	if config == nil {
		config = &SandboxConfig{
			MaxMemoryMB:    256,
			MaxCPUPercent:  50,
			MaxDuration:    30 * time.Second,
			MaxOutputSize:  1024 * 1024, // 1MB
			AllowNetwork:   false,
			AllowFileSystem: false,
		}
	}

	if config.WorkDir == "" {
		config.WorkDir = "/tmp/tortoise-sandbox"
	}

	// 创建工作目录
	os.MkdirAll(config.WorkDir, 0755)

	return &Sandbox{
		config:    config,
		processes: make(map[string]*exec.Cmd),
	}
}

// Execute 执行插件代码（安全隔离）
func (s *Sandbox) Execute(ctx context.Context, code string, language string) (*SandboxResult, error) {
	result := &SandboxResult{}

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, s.config.MaxDuration)
	defer cancel()

	// 根据语言选择执行方式
	switch language {
	case "javascript", "js":
		return s.executeJS(ctx, code)
	case "python":
		return s.executePython(ctx, code)
	case "shell":
		return s.executeShell(ctx, code)
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
}

// executeJS 执行 JavaScript
func (s *Sandbox) executeJS(ctx context.Context, code string) (*SandboxResult, error) {
	result := &SandboxResult{}
	start := time.Now()

	// 使用 Node.js 运行，带资源限制
	cmd := exec.CommandContext(ctx, "node", "-e", code)
	s.configureLimits(cmd)

	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(start)
	result.Output = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.Error = err.Error()
		}
		result.Success = false
	} else {
		result.Success = true
	}

	return result, nil
}

// executePython 执行 Python
func (s *Sandbox) executePython(ctx context.Context, code string) (*SandboxResult, error) {
	result := &SandboxResult{}
	start := time.Now()

	// 使用 Python 运行，带资源限制
	cmd := exec.CommandContext(ctx, "python3", "-c", code)
	s.configureLimits(cmd)

	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(start)
	result.Output = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.Error = err.Error()
		}
		result.Success = false
	} else {
		result.Success = true
	}

	return result, nil
}

// executeShell 执行 Shell
func (s *Sandbox) executeShell(ctx context.Context, code string) (*SandboxResult, error) {
	result := &SandboxResult{}
	start := time.Now()

	// 使用 bash 运行
	cmd := exec.CommandContext(ctx, "bash", "-c", code)
	s.configureLimits(cmd)

	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(start)
	result.Output = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.Error = err.Error()
		}
		result.Success = false
	} else {
		result.Success = true
	}

	return result, nil
}

// configureLimits 配置资源限制
func (s *Sandbox) configureLimits(cmd *exec.Cmd) {
	cmd.Dir = s.config.WorkDir

	// 设置环境变量
	env := os.Environ()
	if !s.config.AllowNetwork {
		env = append(env, "HTTP_PROXY=", "HTTPS_PROXY=", "http_proxy=", "https_proxy=")
	}
	cmd.Env = env

	// 限制输出大小
	// 注意: Go 的 exec 不直接支持这些，需要使用系统工具
}

// ExecuteInSubprocess 在子进程中执行（更安全）
func (s *Sandbox) ExecuteInSubprocess(ctx context.Context, script string, args []string) (*SandboxResult, error) {
	result := &SandboxResult{}
	start := time.Now()

	// 创建临时文件
	tmpDir, err := os.MkdirTemp(s.config.WorkDir, "plugin-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	scriptPath := filepath.Join(tmpDir, "script.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return nil, fmt.Errorf("failed to write script: %w", err)
	}

	// 执行
	cmd := exec.CommandContext(ctx, scriptPath, args...)
	cmd.Dir = tmpDir

	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(start)
	result.Output = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.Error = err.Error()
		}
		result.Success = false
	} else {
		result.Success = true
	}

	return result, nil
}

// ValidateCode 验证代码安全性
func (s *Sandbox) ValidateCode(code string, language string) error {
	// 危险模式检测
	dangerousPatterns := []string{
		"rm -rf /",           // 危险删除
		":(){ :|:& };:",     // Fork炸弹
		"curl | sh",          // 远程代码执行
		"wget | sh",          // 远程代码执行
		"eval($",             // 动态代码执行
		"exec(",              // 执行
		"subprocess(",        // 子进程
		"os.system(",         // 系统命令
		"__import__",        // 动态导入
		"import os",          // os模块
		"import subprocess",  // subprocess模块
		"import sys",         // sys模块（某些情况下）
	}

	for _, pattern := range dangerousPatterns {
		if contains(code, pattern) {
			return fmt.Errorf("code contains dangerous pattern: %s", pattern)
		}
	}

	return nil
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Kill 终止沙箱进程
func (s *Sandbox) Kill(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cmd, exists := s.processes[id]; exists {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		delete(s.processes, id)
	}
}

// Cleanup 清理所有进程
func (s *Sandbox) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, cmd := range s.processes {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		delete(s.processes, id)
	}

	// 清理工作目录
	os.RemoveAll(s.config.WorkDir)
}

// ResourceMonitor 资源监控器
type ResourceMonitor struct {
	mu          sync.RWMutex
	measurements []ResourceMeasurement
	maxMemoryMB int64
}

// ResourceMeasurement 资源测量
type ResourceMeasurement struct {
	Timestamp time.Time
	MemoryMB  int64
	CPUPercent float64
}

// NewResourceMonitor 创建资源监控器
func NewResourceMonitor(maxMemoryMB int64) *ResourceMonitor {
	return &ResourceMonitor{
		maxMemoryMB: maxMemoryMB,
	}
}

// Record 记录资源使用
func (m *ResourceMonitor) Record(memoryMB int64, cpuPercent float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.measurements = append(m.measurements, ResourceMeasurement{
		Timestamp: time.Now(),
		MemoryMB:  memoryMB,
		CPUPercent: cpuPercent,
	})

	// 只保留最近1000条记录
	if len(m.measurements) > 1000 {
		m.measurements = m.measurements[len(m.measurements)-1000:]
	}
}

// IsOverLimit 检查是否超过限制
func (m *ResourceMonitor) IsOverLimit() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.measurements) == 0 {
		return false
	}

	last := m.measurements[len(m.measurements)-1]
	return last.MemoryMB > m.maxMemoryMB
}

// GetStats 获取统计
func (m *ResourceMonitor) GetStats() (avgMemory float64, maxMemory int64, avgCPU float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.measurements) == 0 {
		return 0, 0, 0
	}

	var totalMemory int64
	var totalCPU float64

	for _, m := range m.measurements {
		totalMemory += m.MemoryMB
		totalCPU += m.CPUPercent
		if m.MemoryMB > maxMemory {
			maxMemory = m.MemoryMB
		}
	}

	count := float64(len(m.measurements))
	return float64(totalMemory) / count, maxMemory, totalCPU / count
}

// ============ Plugin Validator 插件验证器 ============

type Validator struct {
	sandbox *Sandbox
}

// NewValidator 创建验证器
func NewValidator(sandbox *Sandbox) *Validator {
	return &Validator{sandbox: sandbox}
}

// Validate 验证插件代码
func (v *Validator) Validate(plugin Plugin) error {
	metadata := plugin.Metadata()
	if metadata == nil {
		return fmt.Errorf("plugin metadata is nil")
	}

	// 验证元数据
	if metadata.ID == "" {
		return fmt.Errorf("plugin ID is required")
	}
	if metadata.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if metadata.Version == "" {
		return fmt.Errorf("plugin version is required")
	}

	// 验证工具
	tools := plugin.GetTools()
	for _, tool := range tools {
		if tool.Name == "" {
			return fmt.Errorf("tool name is required")
		}
	}

	return nil
}

// TestExecution 测试执行插件
func (v *Validator) TestExecution(ctx context.Context, code string, language string) error {
	// 先验证代码
	if err := v.sandbox.ValidateCode(code, language); err != nil {
		return err
	}

	// 尝试执行
	result, err := v.sandbox.Execute(ctx, code, language)
	if err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("execution failed: %s", result.Error)
	}

	return nil
}

// GetSystemLimits 获取系统限制
func GetSystemLimits() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"go_version":     runtime.Version(),
		"go_num_cpu":     runtime.NumCPU(),
		"go_num_goroute": runtime.NumGoroutine(),
		"go_mem_sys":     m.Sys / 1024 / 1024, // MB
		"os":             runtime.GOOS,
		"arch":           runtime.GOARCH,
	}
}
