package skill

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ============ Skill System Core ============

// Skill 技能接口
type Skill interface {
	Name() string
	Description() string
	Version() string
	Category() SkillCategory
	Parameters() []Parameter
	Execute(ctx context.Context, params map[string]interface{}) (*Result, error)
	Validate(params map[string]interface{}) error
}

// Parameter 参数定义
type Parameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required   bool        `json:"required"`
	Default    interface{} `json:"default,omitempty"`
	Options    []string    `json:"options,omitempty"`
}

// ParameterType 参数类型
type ParameterType string

const (
	ParamTypeString  ParameterType = "string"
	ParamTypeNumber ParameterType = "number"
	ParamTypeBoolean ParameterType = "boolean"
	ParamTypeArray  ParameterType = "array"
	ParamTypeObject ParameterType = "object"
)

// SkillCategory 技能分类
type SkillCategory string

const (
	CategoryCore        SkillCategory = "core"
	CategoryAI         SkillCategory = "ai"
	CategoryIntegration SkillCategory = "integration"
	CategoryUtility    SkillCategory = "utility"
	CategoryCustom     SkillCategory = "custom"
)

// Result 技能执行结果
type Result struct {
	Success   bool                   `json:"success"`
	Output    interface{}            `json:"output"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time             `json:"timestamp"`
}

// SkillConfig 技能配置
type SkillConfig struct {
	Name        string                 `json:"name"`
	Enabled     bool                   `json:"enabled"`
	Timeout     time.Duration          `json:"timeout"`
	Retry      int                    `json:"retry"`
	Priority   int                    `json:"priority"`
	Settings   map[string]interface{} `json:"settings"`
}

// Registry 技能注册表
type Registry struct {
	skills    map[string]Skill
	configs   map[string]*SkillConfig
	mu        sync.RWMutex
	observers []Observer
}

// Observer 技能执行观察者
type Observer interface {
	OnSkillStart(req *SkillRequest)
	OnSkillComplete(req *SkillRequest, result *Result)
	OnSkillError(req *SkillRequest, err error)
}

// NewRegistry 创建技能注册表
func NewRegistry() *Registry {
	return &Registry{
		skills:  make(map[string]Skill),
		configs: make(map[string]*SkillConfig),
	}
}

// Register 注册技能
func (r *Registry) Register(skill Skill) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := skill.Name()
	if _, exists := r.skills[name]; exists {
		return fmt.Errorf("skill %s already registered", name)
	}

	r.skills[name] = skill
	r.configs[name] = &SkillConfig{
		Name:      name,
		Enabled:   true,
		Timeout:   30 * time.Second,
		Retry:     0,
		Priority:  0,
		Settings:  make(map[string]interface{}),
	}

	return nil
}

// Unregister 注销技能
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.skills[name]; !exists {
		return fmt.Errorf("skill %s not found", name)
	}

	delete(r.skills, name)
	delete(r.configs, name)
	return nil
}

// Get 获取技能
func (r *Registry) Get(name string) (Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skill, exists := r.skills[name]
	if !exists {
		return nil, fmt.Errorf("skill %s not found", name)
	}
	return skill, nil
}

// List 列出所有技能
func (r *Registry) List() []Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]Skill, 0, len(r.skills))
	for _, s := range r.skills {
		skills = append(skills, s)
	}
	return skills
}

// ListByCategory 按分类列出技能
func (r *Registry) ListByCategory(category SkillCategory) []Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]Skill, 0)
	for _, s := range r.skills {
		if s.Category() == category {
			skills = append(skills, s)
		}
	}
	return skills
}

// Execute 执行技能
func (r *Registry) Execute(ctx context.Context, req *SkillRequest) (*Result, error) {
	r.mu.RLock()
	skill, exists := r.skills[req.SkillName]
	config := r.configs[req.SkillName]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("skill %s not found", req.SkillName)
	}

	if config == nil || !config.Enabled {
		return nil, fmt.Errorf("skill %s is disabled", req.SkillName)
	}

	// 验证参数
	if err := skill.Validate(req.Params); err != nil {
		return nil, err
	}

	// 通知观察者
	for _, obs := range r.observers {
		obs.OnSkillStart(req)
	}

	start := time.Now()

	// 执行技能
	result, err := skill.Execute(ctx, req.Params)
	if err != nil {
		for _, obs := range r.observers {
			obs.OnSkillError(req, err)
		}
		return nil, err
	}

	result.Duration = time.Since(start)
	result.Timestamp = time.Now()

	// 通知观察者
	for _, obs := range r.observers {
		obs.OnSkillComplete(req, result)
	}

	return result, nil
}

// AddObserver 添加观察者
func (r *Registry) AddObserver(obs Observer) {
	r.observers = append(r.observers, obs)
}

// GetConfig 获取技能配置
func (r *Registry) GetConfig(name string) *SkillConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.configs[name]
}

// UpdateConfig 更新技能配置
func (r *Registry) UpdateConfig(name string, cfg *SkillConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.skills[name]; !exists {
		return fmt.Errorf("skill %s not found", name)
	}

	r.configs[name] = cfg
	return nil
}

// SkillRequest 技能请求
type SkillRequest struct {
	SkillName string                 `json:"skill_name"`
	Params    map[string]interface{} `json:"params"`
}

// ToJSON 序列化为JSON
func (r *SkillRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON反序列化
func (r *SkillRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// ToJSON 序列化为JSON
func (res *Result) ToJSON() ([]byte, error) {
	return json.Marshal(res)
}

// FromJSON 从JSON反序列化
func (res *Result) FromJSON(data []byte) error {
	return json.Unmarshal(data, res)
}
