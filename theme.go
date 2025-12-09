package got

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"regexp"
	"sync"
	"sync/atomic"
)

var (
	defineRe   = regexp.MustCompile(`define\s+"([^"]+)"`)
	templateRe = regexp.MustCompile(`(template|block)\s+"([^"]+)"`)
)

type Theme struct {
	name    string
	storage Storage
	cache   sync.Map
	funcMap sync.Map
	debug   atomic.Bool
	parent  atomic.Pointer[Theme]
}

func NewTheme(name string, storage Storage) *Theme {
	return &Theme{
		name:    name,
		storage: storage,
	}
}

func (t *Theme) Clear() {
	t.reset()
}

func (t *Theme) Name() string {
	return t.name
}

func (t *Theme) Debug() bool {
	return t.debug.Load()
}

func (t *Theme) SetDebug(debug bool) {
	if t.debug.Load() == debug {
		return
	}

	t.debug.Store(debug)
	t.reset()
}

func (t *Theme) Parent() *Theme {
	return t.parent.Load()
}

func (t *Theme) SetParent(parent *Theme) {
	t.parent.Store(parent)
	t.reset()
}

func (t *Theme) FuncMap() template.FuncMap {
	funcMap := make(template.FuncMap)
	t.funcMap.Range(func(key, value any) bool {
		funcMap[key.(string)] = value
		return true
	})
	return funcMap
}

func (t *Theme) SetFuncMap(funcMap template.FuncMap) {
	t.funcMap.Clear()
	t.AddFuncMap(funcMap)
}

func (t *Theme) AddFuncMap(funcMap template.FuncMap) {
	for k, v := range funcMap {
		t.funcMap.Store(k, v)
	}
	t.reset()
}

func (t *Theme) reset() {
	t.cache.Clear()

	if parent := t.parent.Load(); parent != nil {
		parent.SetFuncMap(t.FuncMap())
		parent.SetDebug(t.debug.Load())
	}
}

func (t *Theme) Write(ctx context.Context, w io.Writer, name string, data any) error {
	debug := t.debug.Load()

	if !debug {
		if tpl, ok := t.cache.Load(name); ok {
			return tpl.(*template.Template).Execute(w, data)
		}
	}

	tpl, err := t.buildTemplate(ctx, name)
	if err != nil {
		return err
	}

	if !debug {
		t.cache.Store(name, tpl)
	}

	return tpl.Execute(w, data)
}

func (t *Theme) buildTemplate(ctx context.Context, name string) (*template.Template, error) {
	data := make(map[string]Template)
	if err := t.findByName(ctx, data, name); err != nil {
		return nil, err
	}

	page, ok := data[name]
	if !ok {
		return nil, fmt.Errorf("theme: template %s/%s not found: %w", t.name, name, ErrTemplateNotFound)
	}

	for page.Path() != page.Name() {
		page = data[page.Path()]
	}

	funcs := t.FuncMap()

	tpl, err := template.New(page.Name()).Funcs(funcs).Parse(page.Content())
	if err != nil {
		return nil, err
	}

	for _, item := range data {
		if item == page {
			continue
		}

		content := item.Content()

		matches := defineRe.FindAllStringSubmatch(content, -1)

		if len(matches) == 0 {
			if _, err = tpl.New(item.Name()).Funcs(funcs).Parse(content); err != nil {
				return nil, err
			}
			continue
		}

		for _, m := range matches {
			if len(m) > 1 {
				if _, err = tpl.New(m[1]).Funcs(funcs).Parse(content); err != nil {
					return nil, err
				}
			}
		}
	}

	return tpl, nil
}

func (t *Theme) findByName(ctx context.Context, data map[string]Template, name string) error {
	if _, ok := data[name]; ok {
		return nil
	}

	dep, err := t.find(ctx, name)
	if err != nil {
		return err
	}

	data[name] = dep

	if err = t.findByTemplate(ctx, data, dep); err != nil {
		return err
	}

	return nil
}

func (t *Theme) findByTemplate(ctx context.Context, data map[string]Template, item Template) error {
	if item.Path() != item.Name() {
		if err := t.findByName(ctx, data, item.Path()); err != nil {
			return err
		}
	}

	matches := templateRe.FindAllStringSubmatch(item.Content(), -1)
	for _, match := range matches {
		if len(match) > 2 {
			if err := t.findByName(ctx, data, match[2]); err != nil {
				if !errors.Is(err, ErrTemplateNotFound) {
					return err
				}
			}
		}
	}

	return nil
}

func (t *Theme) find(ctx context.Context, name string) (Template, error) {
	item, err := t.storage.Find(ctx, t.name, name)
	if err == nil {
		return item, nil
	}

	if errors.Is(err, ErrTemplateNotFound) {
		if parent := t.parent.Load(); parent != nil {
			item, err1 := parent.find(ctx, name)
			if err1 == nil {
				return item, nil
			}
			err = errors.Join(err, err1)
		}
	}

	return nil, fmt.Errorf("theme: failed to find template %s/%s: %w", t.name, name, err)
}
