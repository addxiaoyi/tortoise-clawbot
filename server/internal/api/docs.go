package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// SwaggerInfo Swagger 信息
var SwaggerInfo = map[string]interface{}{
	"openapi": "3.0.0",
	"info": map[string]interface{}{
		"title":       "Tortoise AI Framework API",
		"description": "高性能 AI 代理框架 API",
		"version":     "1.0.0",
		"contact": map[string]interface{}{
			"name":  "Tortoise Team",
			"email": "support@tortoise.ai",
		},
	},
	"servers": []map[string]string{
		{
			"url":         "http://localhost:18792",
			"description": "本地开发服务器",
		},
	},
}

// API 文档
const SwaggerDoc = `{
  "openapi": "3.0.0",
  "info": {
    "title": "Tortoise AI Framework API",
    "description": "高性能 AI 代理框架 API，支持多渠道消息、AI 对话、记忆管理和插件系统",
    "version": "1.0.0",
    "contact": {
      "name": "Tortoise Team",
      "email": "support@tortoise.ai"
    }
  },
  "servers": [
    {
      "url": "http://localhost:18792",
      "description": "本地开发服务器"
    }
  ],
  "paths": {
    "/api/v1/health": {
      "get": {
        "summary": "健康检查",
        "description": "检查 API 服务器运行状态",
        "tags": ["System"],
        "responses": {
          "200": {
            "description": "服务正常",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "status": {"type": "string", "example": "healthy"},
                    "version": {"type": "string", "example": "1.0.0"},
                    "time": {"type": "string", "example": "2024-01-15T12:00:00Z"}
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/v1/sessions": {
      "get": {
        "summary": "获取会话列表",
        "tags": ["Sessions"],
        "responses": {
          "200": {
            "description": "会话列表",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {"$ref": "#/components/schemas/Session"}
                }
              }
            }
          }
        }
      },
      "post": {
        "summary": "创建会话",
        "tags": ["Sessions"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["name"],
                "properties": {
                  "name": {"type": "string", "description": "会话名称"},
                  "userId": {"type": "string", "description": "用户 ID"}
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "会话已创建",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/Session"}
              }
            }
          }
        }
      }
    },
    "/api/v1/sessions/{id}/messages": {
      "get": {
        "summary": "获取会话消息",
        "tags": ["Messages"],
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {"type": "string"}
          }
        ],
        "responses": {
          "200": {
            "description": "消息列表",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {"$ref": "#/components/schemas/Message"}
                }
              }
            }
          }
        }
      },
      "post": {
        "summary": "发送消息",
        "description": "发送消息给 AI 并获取回复",
        "tags": ["Messages"],
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {"type": "string"}
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["content"],
                "properties": {
                  "content": {"type": "string", "description": "消息内容"},
                  "model": {"type": "string", "description": "指定模型 (可选)"}
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "AI 回复",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "messageId": {"type": "string"},
                    "content": {"type": "string"}
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/v1/memories": {
      "get": {
        "summary": "获取记忆列表",
        "tags": ["Memory"],
        "parameters": [
          {
            "name": "type",
            "in": "query",
            "schema": {"type": "string", "enum": ["working", "semantic", "episodic"]}
          }
        ],
        "responses": {
          "200": {
            "description": "记忆列表"
          }
        }
      },
      "post": {
        "summary": "创建记忆",
        "tags": ["Memory"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["content"],
                "properties": {
                  "type": {"type": "string", "enum": ["working", "semantic", "episodic"]},
                  "content": {"type": "string"},
                  "importance": {"type": "integer", "minimum": 1, "maximum": 10},
                  "tags": {"type": "array", "items": {"type": "string"}}
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "记忆已创建"
          }
        }
      }
    },
    "/api/v1/memories/search": {
      "get": {
        "summary": "搜索记忆",
        "tags": ["Memory"],
        "parameters": [
          {
            "name": "q",
            "in": "query",
            "required": true,
            "schema": {"type": "string"},
            "description": "搜索关键词"
          }
        ],
        "responses": {
          "200": {
            "description": "搜索结果"
          }
        }
      }
    },
    "/api/v1/plugins": {
      "get": {
        "summary": "获取插件列表",
        "tags": ["Plugins"],
        "responses": {
          "200": {
            "description": "插件列表"
          }
        }
      }
    },
    "/api/v1/plugins/{id}/toggle": {
      "post": {
        "summary": "切换插件状态",
        "tags": ["Plugins"],
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {"type": "string"}
          }
        ],
        "responses": {
          "200": {
            "description": "状态已切换"
          }
        }
      }
    },
    "/api/v1/config": {
      "get": {
        "summary": "获取配置",
        "tags": ["Config"],
        "security": [{"ApiKeyAuth": []}],
        "responses": {
          "200": {
            "description": "完整配置"
          }
        }
      },
      "patch": {
        "summary": "更新配置",
        "tags": ["Config"],
        "security": [{"ApiKeyAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "description": "部分配置更新"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "配置已更新"
          }
        }
      }
    },
    "/api/v1/stats": {
      "get": {
        "summary": "获取统计信息",
        "tags": ["Stats"],
        "responses": {
          "200": {
            "description": "统计数据",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "sessions": {"type": "integer"},
                    "messages": {"type": "integer"},
                    "memories": {"type": "integer"},
                    "plugins": {"type": "integer"},
                    "channels": {"type": "integer"}
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/v1/ai/stats": {
      "get": {
        "summary": "获取 AI 统计",
        "tags": ["AI"],
        "responses": {
          "200": {
            "description": "AI 使用统计"
          }
        }
      }
    },
    "/metrics": {
      "get": {
        "summary": "Prometheus 指标",
        "description": "Prometheus 格式的指标数据",
        "tags": ["Monitor"],
        "responses": {
          "200": {
            "description": "Prometheus 指标",
            "content": {
              "text/plain": {
                "schema": {"type": "string"}
              }
            }
          }
        }
      }
    },
    "/ws": {
      "get": {
        "summary": "WebSocket 连接",
        "description": "建立 WebSocket 连接进行实时通信",
        "tags": ["WebSocket"],
        "responses": {
          "101": {
            "description": "切换协议成功"
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "Session": {
        "type": "object",
        "properties": {
          "id": {"type": "string"},
          "name": {"type": "string"},
          "userId": {"type": "string"},
          "createdAt": {"type": "string", "format": "date-time"},
          "updatedAt": {"type": "string", "format": "date-time"},
          "messageCount": {"type": "integer"}
        }
      },
      "Message": {
        "type": "object",
        "properties": {
          "id": {"type": "string"},
          "sessionId": {"type": "string"},
          "role": {"type": "string", "enum": ["user", "assistant", "system"]},
          "content": {"type": "string"},
          "createdAt": {"type": "string", "format": "date-time"}
        }
      },
      "Memory": {
        "type": "object",
        "properties": {
          "id": {"type": "string"},
          "type": {"type": "string", "enum": ["working", "semantic", "episodic"]},
          "content": {"type": "string"},
          "importance": {"type": "integer", "minimum": 1, "maximum": 10},
          "tags": {"type": "array", "items": {"type": "string"}},
          "createdAt": {"type": "string", "format": "date-time"}
        }
      },
      "Plugin": {
        "type": "object",
        "properties": {
          "id": {"type": "string"},
          "name": {"type": "string"},
          "version": {"type": "string"},
          "description": {"type": "string"},
          "author": {"type": "string"},
          "enabled": {"type": "boolean"},
          "status": {"type": "string", "enum": ["active", "inactive", "error"]}
        }
      }
    },
    "securitySchemes": {
      "ApiKeyAuth": {
        "type": "apiKey",
        "in": "header",
        "name": "X-API-Key"
      },
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT"
      }
    }
  },
  "tags": [
    {"name": "System", "description": "系统接口"},
    {"name": "Sessions", "description": "会话管理"},
    {"name": "Messages", "description": "消息管理"},
    {"name": "Memory", "description": "记忆管理"},
    {"name": "Plugins", "description": "插件管理"},
    {"name": "Config", "description": "配置管理"},
    {"name": "Stats", "description": "统计信息"},
    {"name": "AI", "description": "AI 服务"},
    {"name": "Monitor", "description": "监控指标"},
    {"name": "WebSocket", "description": "实时通信"}
  ]
}`

// SwaggerHandler Swagger UI 处理器
func SwaggerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 返回 Swagger JSON
		if strings.HasSuffix(r.URL.Path, ".json") || r.URL.Query().Get("format") == "json" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(SwaggerDoc))
			return
		}

		// 返回 Swagger UI HTML
		html := `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Tortoise AI Framework API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css">
  <style>
    body { background: #1a1a2e; }
    .swagger-ui .topbar { display: none; }
    .swagger-ui .info .title { color: #4ade80; font-size: 2em; }
    .swagger-ui .info description { color: #9ca3af; }
    .swagger-ui .scheme-container { background: #16213e; padding: 20px; border-radius: 8px; }
    .swagger-ui .btn.authorize { background: #4ade80; border-color: #4ade80; color: #000; }
    .swagger-ui select { background: #16213e; color: #fff; border-color: #374151; }
    .swagger-ui input { background: #16213e; color: #fff; border-color: #374151; }
    .swagger-ui .model { color: #60a5fa; }
    .swagger-ui .parameter__name { color: #f472b6; }
    .swagger-ui .opblock-tag { color: #fbbf24; border-color: #374151; }
    .swagger-ui .opblock-summary { color: #e5e7eb; }
    .swagger-ui .tab-item { color: #9ca3af; }
    .swagger-ui .tab-item.active { color: #4ade80; border-color: #4ade80; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      SwaggerUIBundle({
        url: "?format=json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
        layout: "StandaloneLayout",
        docExpansion: "list"
      });
    };
  </script>
</body>
</html>`

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}
}

// RegisterDocsRoutes 注册文档路由
func RegisterDocsRoutes(r *mux.Router) {
	r.HandleFunc("/docs", SwaggerHandler())
	r.HandleFunc("/docs.json", SwaggerHandler())
	log.Printf("📚 API 文档已注册: http://localhost:18792/docs")
}
