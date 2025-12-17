package got

import (
	"context"
	"errors"
	"fmt"
)

var _ Store = (*StoreChain)(nil)

// StoreChain is a store implementation that chains multiple stores together.
type StoreChain struct {
	stores []Store
}

func NewStoreChain(stores ...Store) *StoreChain {
	return &StoreChain{stores: stores}
}

func (s *StoreChain) Add(store Store) {
	s.stores = append(s.stores, store)
}

func (s *StoreChain) Find(ctx context.Context, theme, name string) (Template, error) {
	for _, store := range s.stores {
		tpl, err := store.Find(ctx, theme, name)
		if err == nil {
			return tpl, nil
		}
		if !errors.Is(err, ErrTemplateNotFound) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("store chain: template %s/%s not found: %w", theme, name, ErrTemplateNotFound)
}
