package got

import (
	"context"
	"fmt"
	"sync"
)

var _ Store = (*StoreMemory)(nil)

// StoreMemory is a store implementation that stores templates in memory.
type StoreMemory struct {
	templates sync.Map
}

func NewStoreMemory() *StoreMemory {
	return &StoreMemory{}
}

func (s *StoreMemory) Add(theme, name, content string) {
	s.templates.Store(theme+name, newTemplate(theme, name, content))
}

func (s *StoreMemory) Find(_ context.Context, theme, name string) (Template, error) {
	if v, ok := s.templates.Load(theme + name); ok {
		return v.(Template), nil
	}

	return nil, fmt.Errorf("store memory: template %s/%s not found: %w", theme, name, ErrTemplateNotFound)
}
