// Package plugins provides the plugin system for Tortoise
package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// PluginType defines the type of plugin
type PluginType int

const (
	TypeChannel PluginType = iota
	TypeSkill
	TypeTool
	TypeIntegration
)

// Plugin is the interface all plugins must implement
type Plugin interface {
	// Metadata returns plugin metadata
	Metadata() PluginMetadata
	
	// Initialize initializes the plugin with configuration
	Initialize(ctx context.Context, config json.RawMessage) error
	
	// Start starts the plugin
	Start(ctx context.Context) error
	
	// Stop stops the plugin
	Stop(ctx context.Context) error
	
	// Execute executes a plugin-specific operation
	Execute(ctx context.Context, method string, args json.RawMessage) (json.RawMessage, error)
}

// PluginMetadata contains plugin metadata
type PluginMetadata struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	License     string            `json:"license"`
	Type        PluginType        `json:"type"`
	Tags        []string          `json:"tags"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

// PluginState represents the current state of a plugin
type PluginState int

const (
	StateLoaded PluginState = iota
	StateInitializing
	StateRunning
	StateStopping
	StateStopped
	StateError
)

// PluginInstance represents a loaded plugin instance
type PluginInstance struct {
	Metadata PluginMetadata
	State    PluginState
	Plugin   Plugin
	Error    error
	mu       sync.RWMutex
}

// NewPluginInstance creates a new plugin instance
func NewPluginInstance(plugin Plugin) *PluginInstance {
	return &PluginInstance{
		Metadata: plugin.Metadata(),
		State:   StateLoaded,
		Plugin:   plugin,
	}
}

// SetState sets the plugin state
func (p *PluginInstance) SetState(state PluginState) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.State = state
}

// SetError sets the plugin error
func (p *PluginInstance) SetError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Error = err
	p.State = StateError
}

// State returns the current plugin state
func (p *PluginInstance) State() PluginState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.State
}

// Execute calls the plugin's execute method
func (p *PluginInstance) Execute(ctx context.Context, method string, args json.RawMessage) (json.RawMessage, error) {
	return p.Plugin.Execute(ctx, method, args)
}

// PluginManager manages plugin lifecycle
type PluginManager struct {
	plugins map[string]*PluginInstance
	mu     sync.RWMutex
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]*PluginInstance),
	}
}

// Register registers a plugin
func (m *PluginManager) Register(plugin Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	metadata := plugin.Metadata()
	if _, exists := m.plugins[metadata.ID]; exists {
		return fmt.Errorf("plugin already registered: %s", metadata.ID)
	}

	instance := NewPluginInstance(plugin)
	m.plugins[metadata.ID] = instance
	return nil
}

// Unregister unregisters a plugin
func (m *PluginManager) Unregister(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.plugins[id]
	if !exists {
		return fmt.Errorf("plugin not found: %s", id)
	}

	if instance.State() == StateRunning {
		return fmt.Errorf("cannot unregister running plugin: %s", id)
	}

	delete(m.plugins, id)
	return nil
}

// Get returns a plugin by ID
func (m *PluginManager) Get(id string) *PluginInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.plugins[id]
}

// List returns all registered plugins
func (m *PluginManager) List() []*PluginInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*PluginInstance, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		result = append(result, plugin)
	}
	return result
}

// Load loads and initializes a plugin
func (m *PluginManager) Load(ctx context.Context, id string, config json.RawMessage) error {
	instance := m.Get(id)
	if instance == nil {
		return fmt.Errorf("plugin not found: %s", id)
	}

	instance.SetState(StateInitializing)
	if err := instance.Plugin.Initialize(ctx, config); err != nil {
		instance.SetError(err)
		return fmt.Errorf("failed to initialize plugin %s: %w", id, err)
	}

	return nil
}

// Start starts a plugin
func (m *PluginManager) Start(ctx context.Context, id string) error {
	instance := m.Get(id)
	if instance == nil {
		return fmt.Errorf("plugin not found: %s", id)
	}

	if instance.State() == StateRunning {
		return nil
	}

	instance.SetState(StateRunning)
	if err := instance.Plugin.Start(ctx); err != nil {
		instance.SetError(err)
		return fmt.Errorf("failed to start plugin %s: %w", id, err)
	}

	return nil
}

// Stop stops a plugin
func (m *PluginManager) Stop(ctx context.Context, id string) error {
	instance := m.Get(id)
	if instance == nil {
		return fmt.Errorf("plugin not found: %s", id)
	}

	if instance.State() == StateStopped {
		return nil
	}

	instance.SetState(StateStopping)
	if err := instance.Plugin.Stop(ctx); err != nil {
		instance.SetError(err)
		return fmt.Errorf("failed to stop plugin %s: %w", id, err)
	}

	instance.SetState(StateStopped)
	return nil
}

// Execute executes a plugin method
func (m *PluginManager) Execute(ctx context.Context, id, method string, args json.RawMessage) (json.RawMessage, error) {
	instance := m.Get(id)
	if instance == nil {
		return nil, fmt.Errorf("plugin not found: %s", id)
	}

	if instance.State() != StateRunning {
		return nil, fmt.Errorf("plugin not running: %s", id)
	}

	return instance.Execute(ctx, method, args)
}
