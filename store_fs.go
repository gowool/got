package got

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/gowool/got/internal"
)

var _ Store = (*StoreFS)(nil)

// StoreFS is a store implementation that loads templates from a filesystem.
type StoreFS struct {
	fs fs.FS
}

func NewStoreFS(fsys fs.FS) *StoreFS {
	return &StoreFS{
		fs: fsys,
	}
}

func (s *StoreFS) Find(_ context.Context, theme, name string) (Template, error) {
	fsys, err := fs.Sub(s.fs, theme)
	if err != nil {
		return nil, err
	}

	raw, err := fs.ReadFile(fsys, name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = errors.Join(err, ErrTemplateNotFound)
		}
		return nil, fmt.Errorf("store fs: failed to read template %s/%s: %w", theme, name, err)
	}

	return newTemplate(theme, name, internal.String(raw)), nil
}
