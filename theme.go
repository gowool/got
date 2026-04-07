package got

import (
	"bytes"
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
	name     string
	store    Store
	children sync.Map
	cache    sync.Map
	funcMap  sync.Map
	debug    atomic.Bool
	parent   atomic.Pointer[Theme]
	pool     sync.Pool
}

func NewTheme(name string, store Store) *Theme {
	return &Theme{
		name:  name,
		store: store,
		pool:  sync.Pool{New: func() any { return new(bytes.Buffer) }},
	}
}

func (t *Theme) Clear() {
	t.children.Range(func(_, value any) bool {
		value.(*Theme).SetParent(nil)
		return true
	})
	t.children.Clear()
	t.reset()
}

func (t *Theme) Name() string {
	return t.name
}

func (t *Theme) Debug() bool {
	if parent := t.Parent(); parent != nil {
		return parent.Debug()
	}
	return t.debug.Load()
}

func (t *Theme) SetDebug(debug bool) {
	if parent := t.Parent(); parent != nil {
		parent.SetDebug(debug)
		return
	}

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
	if parent == nil {
		t.reset()
		return
	}

	parent.children.Store(t.name, t)
	parent.parentReset()
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
	t.parentReset()
}

func (t *Theme) parentReset() {
	if parent := t.Parent(); parent != nil {
		parent.parentReset()
		return
	}
	t.reset()
}

func (t *Theme) reset() {
	t.cache.Clear()

	t.children.Range(func(_, value any) bool {
		child := value.(*Theme)

		t.funcMap.Range(func(key, value any) bool {
			if _, ok := child.funcMap.Load(key); !ok {
				child.funcMap.Store(key, value)
			}
			return true
		})

		child.reset()

		return true
	})
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

func (t *Theme) Render(ctx context.Context, name string, data any) ([]byte, error) {
	buf := t.pool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		t.pool.Put(buf)
	}()

	if err := t.Write(ctx, buf, name, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
	item, err := t.store.Find(ctx, t.name, name)
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
