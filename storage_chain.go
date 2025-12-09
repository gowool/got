package got

import (
	"context"
	"errors"
	"fmt"
)

var _ Storage = (*StorageChain)(nil)

// StorageChain is a storage implementation that chains multiple storages together.
type StorageChain struct {
	storages []Storage
}

func NewStorageChain(storages ...Storage) *StorageChain {
	return &StorageChain{storages: storages}
}

func (s *StorageChain) Add(storage Storage) {
	s.storages = append(s.storages, storage)
}

func (s *StorageChain) Find(ctx context.Context, theme, name string) (Template, error) {
	for _, storage := range s.storages {
		tpl, err := storage.Find(ctx, theme, name)
		if err == nil {
			return tpl, nil
		}
		if !errors.Is(err, ErrTemplateNotFound) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("storage chain: template %s/%s not found: %w", theme, name, ErrTemplateNotFound)
}
