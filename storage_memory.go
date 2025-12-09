package got

import (
	"context"
	"fmt"
	"sync"
)

var _ Storage = (*StorageMemory)(nil)

// StorageMemory is a storage implementation that stores templates in memory.
type StorageMemory struct {
	templates sync.Map
}

func NewStorageMemory() *StorageMemory {
	return &StorageMemory{}
}

func (s *StorageMemory) Add(theme, name, content string) {
	s.templates.Store(theme+name, newTemplate(theme, name, content))
}

func (s *StorageMemory) Find(_ context.Context, theme, name string) (Template, error) {
	if v, ok := s.templates.Load(theme + name); ok {
		return v.(Template), nil
	}

	return nil, fmt.Errorf("storage memory: template %s/%s not found: %w", theme, name, ErrTemplateNotFound)
}
