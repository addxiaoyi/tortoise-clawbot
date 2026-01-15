package memory

import "errors"

var (
	ErrMemoryNotFound = errors.New("memory not found")
	ErrStorageFull    = errors.New("memory storage full")
)
