package got

import (
	"regexp"
	"strings"
)

var commentRe = regexp.MustCompile(`^\s*<!--(.*?)-->`)

type Template interface {
	Theme() string
	Path() string
	Name() string
	Content() string
}

type tmpl struct {
	theme   string
	path    string
	name    string
	content string
}

func newTemplate(theme, name, content string) *tmpl {
	p := name
	if comment := commentRe.FindStringSubmatch(content); len(comment) > 0 {
		content = commentRe.ReplaceAllString(content, "")
		p = strings.TrimSpace(comment[1])
	}

	return &tmpl{
		theme:   theme,
		name:    name,
		path:    p,
		content: content,
	}
}

func (t *tmpl) Theme() string {
	return t.theme
}

func (t *tmpl) Path() string {
	return t.path
}

func (t *tmpl) Name() string {
	return t.name
}

func (t *tmpl) Content() string {
	return t.content
}
