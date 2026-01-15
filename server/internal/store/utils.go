package store

import (
	"strings"
)

// containsIgnoreCase 忽略大小写的包含检查
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
