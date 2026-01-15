// File Manager 技能实现
package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FileManagerSkill 文件管理器技能
type FileManagerSkill struct {
	basePath string
}

// NewFileManagerSkill 创建文件管理器技能
func NewFileManagerSkill(basePath string) *FileManagerSkill {
	return &FileManagerSkill{basePath: basePath}
}

// Execute 执行技能
func (s *FileManagerSkill) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Action string `json:"action"`
		Path   string `json:"path"`
		Content string `json:"content,omitempty"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	switch params.Action {
	case "read":
		return s.readFile(params.Path)
	case "write":
		return s.writeFile(params.Path, params.Content)
	case "delete":
		return s.deleteFile(params.Path)
	case "list":
		return s.listFiles(params.Path)
	case "exists":
		return s.fileExists(params.Path)
	case "mkdir":
		return s.createDir(params.Path)
	default:
		return json.Marshal(map[string]interface{}{
			"error": "Unknown action: " + params.Action,
		})
	}
}

// readFile 读取文件
func (s *FileManagerSkill) readFile(path string) (json.RawMessage, error) {
	fullPath := s.resolvePath(path)

	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return json.Marshal(map[string]interface{}{
			"error":   err.Error(),
			"success": false,
		})
	}

	return json.Marshal(map[string]interface{}{
		"success": true,
		"path":    path,
		"content": string(content),
		"size":    len(content),
	})
}

// writeFile 写入文件
func (s *FileManagerSkill) writeFile(path, content string) (json.RawMessage, error) {
	fullPath := s.resolvePath(path)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return json.Marshal(map[string]interface{}{
			"error":   err.Error(),
			"success": false,
		})
	}

	if err := ioutil.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return json.Marshal(map[string]interface{}{
			"error":   err.Error(),
			"success": false,
		})
	}

	return json.Marshal(map[string]interface{}{
		"success": true,
		"path":    path,
		"size":    len(content),
	})
}

// deleteFile 删除文件
func (s *FileManagerSkill) deleteFile(path string) (json.RawMessage, error) {
	fullPath := s.resolvePath(path)

	if err := os.Remove(fullPath); err != nil {
		return json.Marshal(map[string]interface{}{
			"error":   err.Error(),
			"success": false,
		})
	}

	return json.Marshal(map[string]interface{}{
		"success": true,
		"path":    path,
	})
}

// listFiles 列出文件
func (s *FileManagerSkill) listFiles(path string) (json.RawMessage, error) {
	fullPath := s.resolvePath(path)

	entries, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return json.Marshal(map[string]interface{}{
			"error":   err.Error(),
			"success": false,
		})
	}

	var files []map[string]interface{}
	for _, entry := range entries {
		info, _ := entry.Info()
		files = append(files, map[string]interface{}{
			"name":    entry.Name(),
			"is_dir":  entry.IsDir(),
			"size":    info.Size(),
			"mod_time": info.ModTime().Unix(),
		})
	}

	return json.Marshal(map[string]interface{}{
		"success": true,
		"path":    path,
		"files":   files,
	})
}

// fileExists 检查文件是否存在
func (s *FileManagerSkill) fileExists(path string) (json.RawMessage, error) {
	fullPath := s.resolvePath(path)

	_, err := os.Stat(fullPath)
	exists := err == nil

	return json.Marshal(map[string]interface{}{
		"success": true,
		"path":    path,
		"exists":  exists,
	})
}

// createDir 创建目录
func (s *FileManagerSkill) createDir(path string) (json.RawMessage, error) {
	fullPath := s.resolvePath(path)

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return json.Marshal(map[string]interface{}{
			"error":   err.Error(),
			"success": false,
		})
	}

	return json.Marshal(map[string]interface{}{
		"success": true,
		"path":    path,
	})
}

// resolvePath 解析路径
func (s *FileManagerSkill) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(s.basePath, path)
}

// Metadata 返回技能元数据
func (s *FileManagerSkill) Metadata() *SkillMetadata {
	return &SkillMetadata{
		Name:        "file_manager",
		Description: "Read, write, and manage files",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"description": "Action to perform: read, write, delete, list, exists, mkdir",
					"enum":        []string{"read", "write", "delete", "list", "exists", "mkdir"},
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "File or directory path",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "File content (for write action)",
				},
			},
			"required": []string{"action", "path"},
		},
	}
}
