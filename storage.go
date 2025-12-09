package got

import (
	"context"
	"errors"
)

var ErrTemplateNotFound = errors.New("template not found")

// Storage is an interface for loading templates from a storage.
type Storage interface {
	// Find returns a template by its theme and name.
	//
	// If the template is not found, it returns ErrTemplateNotFound.
	Find(ctx context.Context, theme, name string) (Template, error)
}
