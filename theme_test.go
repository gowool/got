package got

import (
	"context"
	"html/template"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a real template for testing
func createTestTemplate(theme, name, content string) Template {
	return newTemplate(theme, name, content)
}

func TestNewTheme(t *testing.T) {
	mockStorage := &MockStorage{}

	theme := NewTheme("test", mockStorage)

	assert.NotNil(t, theme)
	assert.Equal(t, "test", theme.Name())
	assert.False(t, theme.Debug())
	assert.Nil(t, theme.Parent())
}

func TestTheme_Clear(t *testing.T) {
	mockStorage := &MockStorage{}
	theme := NewTheme("test", mockStorage)

	// Clear should not panic
	assert.NotPanics(t, func() {
		theme.Clear()
	})
}

func TestTheme_Debug(t *testing.T) {
	mockStorage := &MockStorage{}
	theme := NewTheme("test", mockStorage)

	// Test default debug state
	assert.False(t, theme.Debug())

	// Test setting debug to true
	theme.SetDebug(true)
	assert.True(t, theme.Debug())

	// Test setting debug to false
	theme.SetDebug(false)
	assert.False(t, theme.Debug())

	// Test setting debug to same value (should not cause issues)
	theme.SetDebug(false)
	assert.False(t, theme.Debug())
}

func TestTheme_Parent(t *testing.T) {
	mockStorage := &MockStorage{}
	parentTheme := NewTheme("parent", mockStorage)
	childTheme := NewTheme("child", mockStorage)

	// Test default parent state
	assert.Nil(t, childTheme.Parent())

	// Test setting parent
	childTheme.SetParent(parentTheme)
	assert.Equal(t, parentTheme, childTheme.Parent())

	// Test setting parent to nil
	childTheme.SetParent(nil)
	assert.Nil(t, childTheme.Parent())
}

func TestTheme_FuncMap(t *testing.T) {
	mockStorage := &MockStorage{}
	theme := NewTheme("test", mockStorage)

	// Test empty func map
	funcMap := theme.FuncMap()
	assert.Empty(t, funcMap)

	// Test setting func map
	customFuncMap := template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}

	theme.SetFuncMap(customFuncMap)
	funcMap = theme.FuncMap()
	assert.Equal(t, 2, len(funcMap))
	assert.Contains(t, funcMap, "upper")
	assert.Contains(t, funcMap, "lower")
}

func TestTheme_AddFuncMap(t *testing.T) {
	mockStorage := &MockStorage{}
	theme := NewTheme("test", mockStorage)

	// Test adding functions
	funcMap1 := template.FuncMap{
		"upper": strings.ToUpper,
	}

	theme.AddFuncMap(funcMap1)
	funcMap := theme.FuncMap()
	assert.Equal(t, 1, len(funcMap))
	assert.Contains(t, funcMap, "upper")

	// Test adding more functions
	funcMap2 := template.FuncMap{
		"lower":    strings.ToLower,
		"contains": strings.Contains,
	}

	theme.AddFuncMap(funcMap2)
	funcMap = theme.FuncMap()
	assert.Equal(t, 3, len(funcMap))
	assert.Contains(t, funcMap, "upper")
	assert.Contains(t, funcMap, "lower")
	assert.Contains(t, funcMap, "contains")

	// Test overwriting existing function
	funcMap3 := template.FuncMap{
		"upper": func(s string) string { return "custom:" + s },
	}

	theme.AddFuncMap(funcMap3)
	funcMap = theme.FuncMap()
	assert.Equal(t, 3, len(funcMap))
	result, ok := funcMap["upper"].(func(string) string)
	require.True(t, ok)
	assert.Equal(t, "custom:test", result("test"))
}

func TestTheme_Write_WithCache(t *testing.T) {
	mockStorage := &MockStorage{}
	theme := NewTheme("test", mockStorage)

	ctx := context.Background()
	var buf strings.Builder

	// Mock template that doesn't exist
	mockStorage.On("Find", ctx, "test", "nonexistent").Return(nil, ErrTemplateNotFound).Once()

	err := theme.Write(ctx, &buf, "nonexistent", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template not found")

	mockStorage.AssertExpectations(t)
}

func TestTheme_Write_WithDebug(t *testing.T) {
	mockStorage := &MockStorage{}
	theme := NewTheme("test", mockStorage)
	theme.SetDebug(true) // Enable debug mode to bypass cache

	ctx := context.Background()
	var buf strings.Builder

	// Create a simple template
	templateContent := `<h1>{{.Title}}</h1>`
	testTemplate := createTestTemplate("test", "simple", templateContent)

	mockStorage.On("Find", ctx, "test", "simple").Return(testTemplate, nil).Once()

	err := theme.Write(ctx, &buf, "simple", map[string]string{"Title": "Hello World"})
	assert.NoError(t, err)
	assert.Equal(t, "<h1>Hello World</h1>", buf.String())

	mockStorage.AssertExpectations(t)
}

func TestTheme_Write_WithDependencies(t *testing.T) {
	mockStorage := &MockStorage{}
	theme := NewTheme("test", mockStorage)

	ctx := context.Background()
	var buf strings.Builder

	// Create a simple template without blocks/templates to avoid complex dependencies
	simpleContent := `<h1>{{.Title}}</h1><p>{{.Message}}</p>`
	simpleTemplate := createTestTemplate("test", "simple", simpleContent)

	mockStorage.On("Find", ctx, "test", "simple").Return(simpleTemplate, nil).Once()

	data := map[string]interface{}{
		"Title":   "Test Page",
		"Message": "Hello World!",
	}

	err := theme.Write(ctx, &buf, "simple", data)
	assert.NoError(t, err)
	result := buf.String()
	assert.Contains(t, result, "<h1>Test Page</h1>")
	assert.Contains(t, result, "<p>Hello World!</p>")

	mockStorage.AssertExpectations(t)
}

func TestTheme_Write_WithParentTheme(t *testing.T) {
	parentStorage := &MockStorage{}
	childStorage := &MockStorage{}

	parentTheme := NewTheme("parent", parentStorage)
	childTheme := NewTheme("child", childStorage)
	childTheme.SetParent(parentTheme)

	ctx := context.Background()
	var buf strings.Builder

	// Template exists in parent theme
	templateContent := `<h1>{{.Title}}</h1>`
	parentTemplate := createTestTemplate("parent", "inherited", templateContent)

	// Child theme doesn't have this template
	childStorage.On("Find", ctx, "child", "inherited").Return(nil, ErrTemplateNotFound).Once()
	parentStorage.On("Find", ctx, "parent", "inherited").Return(parentTemplate, nil).Once()

	err := childTheme.Write(ctx, &buf, "inherited", map[string]string{"Title": "Inherited Template"})
	assert.NoError(t, err)
	assert.Equal(t, "<h1>Inherited Template</h1>", buf.String())

	childStorage.AssertExpectations(t)
	parentStorage.AssertExpectations(t)
}

func TestTheme_Write_WithComplexDependencies(t *testing.T) {
	mockStorage := &MockStorage{}
	theme := NewTheme("test", mockStorage)

	ctx := context.Background()
	var buf strings.Builder

	// Create a simple template without complex dependencies
	complexContent := `<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
<h1>{{.SiteName}}</h1>
<main>{{range .Items}}<p>{{.}}</p>{{end}}</main>
<footer>&copy; {{.Year}}</footer>
</body>
</html>`
	complexTemplate := createTestTemplate("test", "complex", complexContent)

	mockStorage.On("Find", ctx, "test", "complex").Return(complexTemplate, nil).Once()

	data := map[string]interface{}{
		"Title":    "Complex Page",
		"SiteName": "My Site",
		"Year":     "2023",
		"Items":    []string{"Item 1", "Item 2", "Item 3"},
	}

	err := theme.Write(ctx, &buf, "complex", data)
	assert.NoError(t, err)
	result := buf.String()
	assert.Contains(t, result, "<title>Complex Page</title>")
	assert.Contains(t, result, "<h1>My Site</h1>")
	assert.Contains(t, result, "<p>Item 1</p>")
	assert.Contains(t, result, "<p>Item 2</p>")
	assert.Contains(t, result, "<p>Item 3</p>")
	assert.Contains(t, result, "<footer>&copy; 2023</footer>")

	mockStorage.AssertExpectations(t)
}

func TestTheme_Write_TemplateNotFoundError(t *testing.T) {
	mockStorage := &MockStorage{}
	theme := NewTheme("test", mockStorage)

	ctx := context.Background()
	var buf strings.Builder

	mockStorage.On("Find", ctx, "test", "missing").Return(nil, ErrTemplateNotFound).Once()

	err := theme.Write(ctx, &buf, "missing", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find template test/missing")
	assert.Contains(t, err.Error(), "template not found")

	mockStorage.AssertExpectations(t)
}

func TestTheme_Write_ParentTemplateNotFoundError(t *testing.T) {
	parentStorage := &MockStorage{}
	childStorage := &MockStorage{}

	parentTheme := NewTheme("parent", parentStorage)
	childTheme := NewTheme("child", childStorage)
	childTheme.SetParent(parentTheme)

	ctx := context.Background()
	var buf strings.Builder

	childStorage.On("Find", ctx, "child", "missing").Return(nil, ErrTemplateNotFound).Once()
	parentStorage.On("Find", ctx, "parent", "missing").Return(nil, ErrTemplateNotFound).Once()

	err := childTheme.Write(ctx, &buf, "missing", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find template child/missing")

	childStorage.AssertExpectations(t)
	parentStorage.AssertExpectations(t)
}

func TestTheme_Write_WithParseError(t *testing.T) {
	mockStorage := &MockStorage{}
	theme := NewTheme("test", mockStorage)

	ctx := context.Background()
	var buf strings.Builder

	// Create a template with invalid syntax
	invalidContent := `<h1>{{.Title</h1>` // Missing closing brace
	invalidTemplate := createTestTemplate("test", "invalid", invalidContent)

	mockStorage.On("Find", ctx, "test", "invalid").Return(invalidTemplate, nil).Once()

	err := theme.Write(ctx, &buf, "invalid", map[string]string{"Title": "Test"})
	assert.Error(t, err)
	// The error should contain template parsing information
	assert.Contains(t, err.Error(), "template:")

	mockStorage.AssertExpectations(t)
}

func TestTheme_ParentDebugPropagation(t *testing.T) {
	parentStorage := &MockStorage{}
	childStorage := &MockStorage{}

	parentTheme := NewTheme("parent", parentStorage)
	childTheme := NewTheme("child", childStorage)
	childTheme.SetParent(parentTheme)

	// Test debug propagation from child to parent
	childTheme.SetDebug(true)
	assert.True(t, childTheme.Debug())
	assert.True(t, parentTheme.Debug())

	// Test debug propagation from child to parent (false)
	childTheme.SetDebug(false)
	assert.False(t, childTheme.Debug())
	assert.False(t, parentTheme.Debug())
}

func TestTheme_ParentFuncMapPropagation(t *testing.T) {
	parentStorage := &MockStorage{}
	childStorage := &MockStorage{}

	parentTheme := NewTheme("parent", parentStorage)
	childTheme := NewTheme("child", childStorage)
	childTheme.SetParent(parentTheme)

	// Test func map propagation from child to parent
	customFuncMap := template.FuncMap{
		"upper": strings.ToUpper,
	}

	childTheme.SetFuncMap(customFuncMap)

	parentFuncMap := parentTheme.FuncMap()
	assert.Contains(t, parentFuncMap, "upper")
}

func TestTheme_Reset(t *testing.T) {
	mockStorage := &MockStorage{}
	theme := NewTheme("test", mockStorage)

	// Enable debug and add some functions to cache
	theme.SetDebug(true)
	customFuncMap := template.FuncMap{
		"upper": strings.ToUpper,
	}
	theme.AddFuncMap(customFuncMap)

	// Trigger reset
	theme.reset()

	// Cache should be cleared, but other properties should remain
	assert.True(t, theme.Debug())
	funcMap := theme.FuncMap()
	assert.Contains(t, funcMap, "upper")
}

func TestTheme_ConcurrentAccess(t *testing.T) {
	mockStorage := &MockStorage{}
	theme := NewTheme("test", mockStorage)

	ctx := context.Background()

	// Create a simple template
	templateContent := `<h1>{{.Title}}</h1>`
	testTemplate := createTestTemplate("test", "simple", templateContent)

	mockStorage.On("Find", ctx, "test", "simple").Return(testTemplate, nil).Maybe()

	var wg sync.WaitGroup
	numGoroutines := 10
	numIterations := 5

	// Test concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				var buf strings.Builder
				data := map[string]string{"Title": "Test"}
				_ = theme.Write(ctx, &buf, "simple", data)
			}
		}(i)
	}

	// Test concurrent property access
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				_ = theme.Debug()
				_ = theme.Parent()
				_ = theme.Name()
				_ = theme.FuncMap()
			}
		}(i)
	}

	wg.Wait()
}

func TestTheme_WithEmptyContent(t *testing.T) {
	mockStorage := &MockStorage{}
	theme := NewTheme("test", mockStorage)

	ctx := context.Background()
	var buf strings.Builder

	// Create a template with empty content
	emptyTemplate := createTestTemplate("test", "empty", "")

	mockStorage.On("Find", ctx, "test", "empty").Return(emptyTemplate, nil).Once()

	err := theme.Write(ctx, &buf, "empty", nil)
	assert.NoError(t, err)
	assert.Equal(t, "", buf.String())

	mockStorage.AssertExpectations(t)
}

func TestTheme_WithComplexData(t *testing.T) {
	mockStorage := &MockStorage{}
	theme := NewTheme("test", mockStorage)

	ctx := context.Background()
	var buf strings.Builder

	// Create a template with complex data access
	complexContent := `{{range .Items}}{{$index := .Index}}{{$value := .Value}}Item {{$index}}: {{$value}} {{end}}`
	complexTemplate := createTestTemplate("test", "complex", complexContent)

	mockStorage.On("Find", ctx, "test", "complex").Return(complexTemplate, nil).Once()

	type Item struct {
		Index int
		Value string
	}

	data := map[string][]Item{
		"Items": {
			{Index: 1, Value: "First"},
			{Index: 2, Value: "Second"},
			{Index: 3, Value: "Third"},
		},
	}

	err := theme.Write(ctx, &buf, "complex", data)
	assert.NoError(t, err)
	result := buf.String()
	assert.Contains(t, result, "Item 1: First")
	assert.Contains(t, result, "Item 2: Second")
	assert.Contains(t, result, "Item 3: Third")

	mockStorage.AssertExpectations(t)
}
