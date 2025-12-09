package got

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/gowool/got/internal"
)

var ErrTemplateNotFound = errors.New("template not found")

type Storage interface {
	Find(ctx context.Context, theme, name string) (Template, error)
}

type StorageFS struct {
	fs fs.FS
}

func NewStorageFS(fsys fs.FS) *StorageFS {
	return &StorageFS{
		fs: fsys,
	}
}

func (s *StorageFS) Find(_ context.Context, theme, name string) (Template, error) {
	fsys, err := fs.Sub(s.fs, theme)
	if err != nil {
		return nil, err
	}

	raw, err := fs.ReadFile(fsys, name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = errors.Join(err, ErrTemplateNotFound)
		}
		return nil, fmt.Errorf("storage: failed to read template %s/%s: %w", theme, name, err)
	}

	return newTemplate(theme, name, internal.String(raw)), nil
}
