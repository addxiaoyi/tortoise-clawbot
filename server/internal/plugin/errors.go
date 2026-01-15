package plugin

import "errors"

var (
	ErrPluginNotFound  = errors.New("plugin not found")
	ErrToolNotFound    = errors.New("tool not found")
	ErrPluginDisabled  = errors.New("plugin disabled")
)
