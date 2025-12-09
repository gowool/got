package got

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemplate(t *testing.T) {
	tests := []struct {
		name     string
		theme    string
		tmplName string
		content  string
		expected *tmpl
	}{
		{
			name:     "template with HTML comment path",
			theme:    "default",
			tmplName: "index.html",
			content:  "<!-- layouts/base -->\n<html><body>{{.Content}}</body></html>",
			expected: &tmpl{
				theme:   "default",
				name:    "index.html",
				path:    "layouts/base",
				content: "\n<html><body>{{.Content}}</body></html>",
			},
		},
		{
			name:     "template with comment having whitespace",
			theme:    "admin",
			tmplName: "dashboard.html",
			content:  "<!--    layouts/admin    -->\n<div>{{.Dashboard}}</div>",
			expected: &tmpl{
				theme:   "admin",
				name:    "dashboard.html",
				path:    "layouts/admin",
				content: "\n<div>{{.Dashboard}}</div>",
			},
		},
		{
			name:     "template without HTML comment",
			theme:    "simple",
			tmplName: "plain.html",
			content:  "<div>Simple template without comment</div>",
			expected: &tmpl{
				theme:   "simple",
				name:    "plain.html",
				path:    "plain.html",
				content: "<div>Simple template without comment</div>",
			},
		},
		{
			name:     "template with empty comment",
			theme:    "test",
			tmplName: "empty.html",
			content:  "<!-- -->\n<div>Empty comment</div>",
			expected: &tmpl{
				theme:   "test",
				name:    "empty.html",
				path:    "", // Empty comment results in empty path after TrimSpace
				content: "\n<div>Empty comment</div>",
			},
		},
		{
			name:     "template with multi-line comment (regex doesn't match multiline)",
			theme:    "multi",
			tmplName: "multi.html",
			content:  "<!-- \nlayouts/multi\n-->\n<div>Multi-line</div>",
			expected: &tmpl{
				theme:   "multi",
				name:    "multi.html",
				path:    "multi.html",                                       // Regex doesn't match multiline, so falls back to name
				content: "<!-- \nlayouts/multi\n-->\n<div>Multi-line</div>", // Content unchanged
			},
		},
		{
			name:     "template with leading whitespace before comment",
			theme:    "spaced",
			tmplName: "spaced.html",
			content:  "   <!-- layouts/spaced -->\n<div>Spaced template</div>",
			expected: &tmpl{
				theme:   "spaced",
				name:    "spaced.html",
				path:    "layouts/spaced",
				content: "\n<div>Spaced template</div>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := newTemplate(tt.theme, tt.tmplName, tt.content)

			assert.Equal(t, tt.expected.theme, result.theme, "Theme should match expected")
			assert.Equal(t, tt.expected.name, result.name, "Name should match expected")
			assert.Equal(t, tt.expected.path, result.path, "Path should match expected")
			assert.Equal(t, tt.expected.content, result.content, "Content should match expected")
		})
	}
}

func TestTemplateInterfaceMethods(t *testing.T) {
	theme := "test-theme"
	name := "test.html"
	content := "<!-- layouts/test -->\n<div>{{.Content}}</div>"

	tmpl := newTemplate(theme, name, content)

	t.Run("Theme method", func(t *testing.T) {
		result := tmpl.Theme()
		assert.Equal(t, theme, result, "Theme() should return the theme name")
	})

	t.Run("Name method", func(t *testing.T) {
		result := tmpl.Name()
		assert.Equal(t, name, result, "Name() should return the template name")
	})

	t.Run("Path method", func(t *testing.T) {
		result := tmpl.Path()
		assert.Equal(t, "layouts/test", result, "Path() should return the extracted path from comment")
	})

	t.Run("Content method", func(t *testing.T) {
		result := tmpl.Content()
		expectedContent := "\n<div>{{.Content}}</div>"
		assert.Equal(t, expectedContent, result, "Content() should return content without HTML comment")
	})
}

func TestTemplateWithoutComment(t *testing.T) {
	theme := "no-comment"
	name := "plain.html"
	content := "<div>Template without comment</div>"

	tmpl := newTemplate(theme, name, content)

	assert.Equal(t, theme, tmpl.Theme(), "Theme should be set correctly")
	assert.Equal(t, name, tmpl.Name(), "Name should be set correctly")
	assert.Equal(t, name, tmpl.Path(), "Path should default to name when no comment exists")
	assert.Equal(t, content, tmpl.Content(), "Content should remain unchanged")
}

func TestTemplateCommentExtraction(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectedPath string
		expectedBody string
	}{
		{
			name:         "simple comment",
			content:      "<!-- layouts/base -->content",
			expectedPath: "layouts/base",
			expectedBody: "content",
		},
		{
			name:         "comment with spaces",
			content:      "<!--  layouts/spaced  -->content",
			expectedPath: "layouts/spaced",
			expectedBody: "content",
		},
		{
			name:         "comment with tabs",
			content:      "<!--\tlayouts/tabbed\t-->content",
			expectedPath: "layouts/tabbed",
			expectedBody: "content",
		},
		{
			name:         "comment on multiple lines (no match)",
			content:      "<!--\nlayouts\nmulti\n-->content",
			expectedPath: "test.html",                        // Falls back to template name
			expectedBody: "<!--\nlayouts\nmulti\n-->content", // Content unchanged
		},
		{
			name:         "comment with extra whitespace around path",
			content:      "<!-- \t layouts/whitespace \t -->content",
			expectedPath: "layouts/whitespace",
			expectedBody: "content",
		},
		{
			name:         "empty comment",
			content:      "<!-- -->content",
			expectedPath: "", // Empty after TrimSpace
			expectedBody: "content",
		},
		{
			name:         "comment only whitespace (multiline no match)",
			content:      "<!--    \t\n   -->content",
			expectedPath: "test.html",                 // Falls back to template name (no regex match)
			expectedBody: "<!--    \t\n   -->content", // Content unchanged
		},
		{
			name:         "no comment",
			content:      "plain content",
			expectedPath: "test.html", // Falls back to template name
			expectedBody: "plain content",
		},
		{
			name:         "multiple comments (should only extract first)",
			content:      "<!-- first -->content<!-- second -->",
			expectedPath: "first",
			expectedBody: "content<!-- second -->",
		},
		{
			name:         "comment in middle of content (no match)",
			content:      "start<!-- middle -->end",
			expectedPath: "test.html",               // Falls back to template name
			expectedBody: "start<!-- middle -->end", // Content unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := newTemplate("test", "test.html", tt.content)
			assert.Equal(t, tt.expectedPath, tmpl.Path(), "Path extraction should match expected")
			assert.Equal(t, tt.expectedBody, tmpl.Content(), "Content after comment removal should match expected")
		})
	}
}

func TestTemplateImplementsInterface(t *testing.T) {
	// This test verifies that the tmpl struct implements the Template interface
	var _ Template = (*tmpl)(nil)

	tmpl := newTemplate("test", "test.html", "<!-- path -->content")

	// All interface methods should be available and not panic
	require.NotPanics(t, func() {
		tmpl.Theme()
		tmpl.Path()
		tmpl.Name()
		tmpl.Content()
	})
}

func TestTemplateWithComplexContent(t *testing.T) {
	theme := "complex"
	name := "complex.html"
	content := `<!-- layouts/complex -->
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    {{template "header" .}}
    <main>{{.Content}}</main>
    {{template "footer" .}}
</body>
</html>`

	tmpl := newTemplate(theme, name, content)

	expectedContent := `
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    {{template "header" .}}
    <main>{{.Content}}</main>
    {{template "footer" .}}
</body>
</html>`

	assert.Equal(t, theme, tmpl.Theme())
	assert.Equal(t, name, tmpl.Name())
	assert.Equal(t, "layouts/complex", tmpl.Path())
	assert.Equal(t, expectedContent, tmpl.Content())
}

func TestTemplateEdgeCases(t *testing.T) {
	t.Run("empty template name", func(t *testing.T) {
		content := "<!-- path -->content"
		tmpl := newTemplate("theme", "", content)

		assert.Equal(t, "", tmpl.Name())
		assert.Equal(t, "path", tmpl.Path())
	})

	t.Run("empty content", func(t *testing.T) {
		tmpl := newTemplate("theme", "empty.html", "")

		assert.Equal(t, "theme", tmpl.Theme())
		assert.Equal(t, "empty.html", tmpl.Name())
		assert.Equal(t, "empty.html", tmpl.Path()) // Should default to name
		assert.Equal(t, "", tmpl.Content())
	})

	t.Run("template name with path", func(t *testing.T) {
		content := "<!-- layouts/base -->content"
		tmpl := newTemplate("theme", "templates/index.html", content)

		assert.Equal(t, "templates/index.html", tmpl.Name())
		assert.Equal(t, "layouts/base", tmpl.Path()) // Comment path takes precedence
	})
}
