# Tortoise 插件开发指南

## 概述

Tortoise 插件系统允许扩展核心功能。插件可以添加新工具、渠道或处理逻辑。

## 插件接口

```go
package plugin

import "context"

// Plugin 是所有 Tortoise 插件必须实现的接口
type Plugin interface {
    // Name 返回插件名称
    Name() string
    
    // Version 返回插件版本
    Version() string
    
    // Description 返回插件描述
    Description() string
    
    // Init 初始化插件
    Init(ctx context.Context, config map[string]interface{}) error
    
    // Execute 执行插件逻辑
    Execute(ctx context.Context, req *Request) (*Response, error)
    
    // Tools 返回插件提供的工具列表
    Tools() []ToolDefinition
    
    // Cleanup 清理插件资源
    Cleanup() error
}
```

## 创建插件

### 1. 创建插件项目

```bash
mkdir -p my-plugin
cd my-plugin
go mod init github.com/tortoise/my-plugin
```

### 2. 实现插件接口

```go
package main

import (
    "context"
    
    "github.com/tortoise/server/plugin"
)

type MyPlugin struct{}

func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) Version() string {
    return "1.0.0"
}

func (p *MyPlugin) Description() string {
    return "My custom plugin for Tortoise"
}

func (p *MyPlugin) Init(ctx context.Context, config map[string]interface{}) error {
    // 初始化逻辑
    return nil
}

func (p *MyPlugin) Execute(ctx context.Context, req *plugin.Request) (*plugin.Response, error) {
    return &plugin.Response{
        Success: true,
        Data:    map[string]interface{}{"result": "ok"},
    }, nil
}

func (p *MyPlugin) Tools() []plugin.ToolDefinition {
    return []plugin.ToolDefinition{
        {
            Name:        "my_tool",
            Description: "A custom tool",
            Parameters: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "input": map[string]interface{}{
                        "type": "string",
                    },
                },
                "required": []string{"input"},
            },
        },
    }
}

func (p *MyPlugin) Cleanup() error {
    return nil
}

func main() {}
```

### 3. 导出插件

```go
// ExportPlugin 导出插件实例供 Tortoise 加载
func ExportPlugin() plugin.Plugin {
    return &MyPlugin{}
}
```

## 工具定义

```go
type ToolDefinition struct {
    Name        string                 // 工具名称
    Description string                 // 工具描述
    Parameters  map[string]interface{} // JSON Schema 参数定义
}
```

### 参数示例

```go
Parameters: map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "query": map[string]interface{}{
            "type":        "string",
            "description": "Search query",
        },
        "limit": map[string]interface{}{
            "type":    "integer",
            "default": 10,
        },
    },
    "required": []string{"query"},
}
```

## 请求与响应

```go
// Request 插件请求
type Request struct {
    Tool      string                 // 工具名称
    Arguments map[string]interface{} // 调用参数
    Context   map[string]interface{} // 上下文信息
}

// Response 插件响应
type Response struct {
    Success bool                   // 是否成功
    Data    interface{}           // 返回数据
    Error   string                // 错误信息
}
```

## 安装插件

### 本地安装

```bash
# 编译插件
go build -o my-plugin.so -buildmode=plugin my-plugin/

# 移动到插件目录
mv my-plugin.so ~/.tortoise/plugins/
```

### API 安装

```bash
curl -X POST http://localhost:18792/api/v1/plugins/install \
  -H "Content-Type: application/json" \
  -d '{"source": "github", "repo": "user/my-plugin"}'
```

## 内置工具

Tortoise 提供以下内置工具：

### Web Search

```go
{
    "name": "web_search",
    "description": "Search the web for information",
    "parameters": {
        "type": "object",
        "properties": {
            "query": {"type": "string"},
            "num_results": {"type": "integer", "default": 5}
        },
        "required": ["query"]
    }
}
```

### Calculator

```go
{
    "name": "calculator",
    "description": "Perform mathematical calculations",
    "parameters": {
        "type": "object",
        "properties": {
            "expression": {"type": "string"}
        },
        "required": ["expression"]
    }
}
```

### File System

```go
{
    "name": "file_read",
    "description": "Read file contents",
    "parameters": {
        "type": "object",
        "properties": {
            "path": {"type": "string"}
        },
        "required": ["path"]
    }
}
```

## 调试

### 日志

使用 `tracing` 或 `zerolog` 输出日志：

```go
import "github.com/rs/zerolog/log"

func (p *MyPlugin) Execute(ctx context.Context, req *plugin.Request) (*plugin.Response, error) {
    log.Info().Str("tool", req.Tool).Msg("Executing tool")
    // ...
}
```

### 测试

```go
func TestMyPlugin(t *testing.T) {
    p := &MyPlugin{}
    
    resp, err := p.Execute(context.Background(), &plugin.Request{
        Tool:      "my_tool",
        Arguments: map[string]interface{}{"input": "test"},
    })
    
    if err != nil {
        t.Fatal(err)
    }
    
    if !resp.Success {
        t.Fatalf("Expected success, got error: %s", resp.Error)
    }
}
```
