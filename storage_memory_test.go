package got

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStorageMemory(t *testing.T) {
	storage := NewStorageMemory()
	require.NotNil(t, storage, "NewStorageMemory() returned nil")

	// Verify it's properly initialized by checking if we can add/find templates
	storage.Add("test", "example.html", "<div>test</div>")

	template, err := storage.Find(context.Background(), "test", "example.html")
	assert.NoError(t, err, "Expected to find template after initialization")
	assert.NotNil(t, template, "Expected template to be found")
}

func TestStorageMemory_Add(t *testing.T) {
	storage := NewStorageMemory()

	tests := []struct {
		name     string
		theme    string
		template string
		content  string
	}{
		{
			name:     "basic template",
			theme:    "default",
			template: "home.html",
			content:  "<div>Home</div>",
		},
		{
			name:     "template with HTML comment",
			theme:    "admin",
			template: "dashboard.html",
			content:  "<!-- layouts/admin -->\n<div>Dashboard</div>",
		},
		{
			name:     "template with complex content",
			theme:    "blog",
			template: "post.html",
			content:  "{{define \"content\"}}<article>Post content</article>{{end}}",
		},
		{
			name:     "empty template",
			theme:    "empty",
			template: "blank.html",
			content:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.Add(tt.theme, tt.template, tt.content)

			// Verify the template was added by finding it
			template, err := storage.Find(context.Background(), tt.theme, tt.template)
			require.NoError(t, err, "Add() failed, template not found")
			require.NotNil(t, template, "Expected template to be found")

			assert.Equal(t, tt.theme, template.Theme(), "Expected theme to match")
			assert.Equal(t, tt.template, template.Name(), "Expected name to match")
		})
	}
}

func TestStorageMemory_Find(t *testing.T) {
	storage := NewStorageMemory()

	// Add some test templates
	storage.Add("default", "home.html", "<div>Home</div>")
	storage.Add("admin", "dashboard.html", "<!-- layouts/admin -->\n<div>Dashboard</div>")
	storage.Add("blog", "post.html", "<!-- layouts/post -->\n{{define \"content\"}}<article>Post</article>{{end}}")

	tests := []struct {
		name     string
		theme    string
		template string
		wantErr  bool
		errIs    error
	}{
		{
			name:     "find existing template",
			theme:    "default",
			template: "home.html",
			wantErr:  false,
		},
		{
			name:     "find template with HTML comment",
			theme:    "admin",
			template: "dashboard.html",
			wantErr:  false,
		},
		{
			name:     "find template with define blocks",
			theme:    "blog",
			template: "post.html",
			wantErr:  false,
		},
		{
			name:     "non-existent template",
			theme:    "default",
			template: "missing.html",
			wantErr:  true,
			errIs:    ErrTemplateNotFound,
		},
		{
			name:     "non-existent theme",
			theme:    "missing",
			template: "home.html",
			wantErr:  true,
			errIs:    ErrTemplateNotFound,
		},
		{
			name:     "both theme and template missing",
			theme:    "missing",
			template: "missing.html",
			wantErr:  true,
			errIs:    ErrTemplateNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := storage.Find(context.Background(), tt.theme, tt.template)

			if tt.wantErr {
				assert.Error(t, err, "Expected error but got none")
				if tt.errIs != nil {
					assert.ErrorIs(t, err, tt.errIs, "Expected specific error type")
				}
				return
			}

			assert.NoError(t, err, "Unexpected error")
			assert.NotNil(t, template, "Expected template but got nil")
			assert.Equal(t, tt.theme, template.Theme(), "Expected theme to match")
			assert.Equal(t, tt.template, template.Name(), "Expected name to match")
		})
	}
}

func TestStorageMemory_Find_WithContext(t *testing.T) {
	storage := NewStorageMemory()
	storage.Add("test", "example.html", "<div>Test</div>")

	// Test with context
	ctx := context.Background()
	template, err := storage.Find(ctx, "test", "example.html")

	assert.NoError(t, err, "Unexpected error with context")
	assert.NotNil(t, template, "Expected template but got nil")

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	template, err = storage.Find(ctx, "test", "example.html")

	// Note: Current implementation doesn't check context cancellation,
	// but this test ensures it works with cancelled contexts
	assert.NoError(t, err, "Unexpected error with cancelled context")
	assert.NotNil(t, template, "Expected template even with cancelled context")
}

func TestStorageMemory_ConcurrentAccess(t *testing.T) {
	storage := NewStorageMemory()

	// Add initial template
	storage.Add("test", "base.html", "<div>Base</div>")

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Test concurrent Add operations
	t.Run("concurrent adds", func(t *testing.T) {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 10; j++ { // Reduce operations to avoid too many templates
					// Use unique keys to avoid overwrites
					templateName := fmt.Sprintf("template_%d_%d.html", id, j)
					storage.Add("concurrent", templateName, fmt.Sprintf("<div>Content from goroutine %d, iteration %d</div>", id, j))
				}
			}(i)
		}

		wg.Wait()

		// Verify some templates were added
		template, err := storage.Find(context.Background(), "concurrent", "template_0_0.html")
		assert.NoError(t, err, "Failed to find concurrently added template")
		assert.NotNil(t, template, "Expected template but got nil")
	})

	// Test concurrent Find operations
	t.Run("concurrent finds", func(t *testing.T) {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					_, err := storage.Find(context.Background(), "test", "base.html")
					assert.NoError(t, err, "Concurrent find failed: %v", err)
				}
			}()
		}

		wg.Wait()
	})

	// Test mixed concurrent operations
	t.Run("mixed concurrent operations", func(t *testing.T) {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					// Add operation
					templateName := fmt.Sprintf("mixed_%d_%d.html", id, j)
					storage.Add("mixed", templateName, "<div>Mixed</div>")

					// Find operation
					_, err := storage.Find(context.Background(), "test", "base.html")
					assert.NoError(t, err, "Mixed operation find failed: %v", err)

					// Find the template we just added
					_, err = storage.Find(context.Background(), "mixed", templateName)
					assert.NoError(t, err, "Mixed operation find added template failed: %v", err)
				}
			}(i)
		}

		wg.Wait()
	})
}

func TestStorageMemory_KeyGeneration(t *testing.T) {
	storage := NewStorageMemory()

	// Test that theme+name is used as the key
	storage.Add("theme1", "name1", "<div>Content 1</div>")
	storage.Add("theme1", "name2", "<div>Content 2</div>")
	storage.Add("theme2", "name1", "<div>Content 3</div>")

	// Verify all three templates exist and are different
	tmpl1, err1 := storage.Find(context.Background(), "theme1", "name1")
	tmpl2, err2 := storage.Find(context.Background(), "theme1", "name2")
	tmpl3, err3 := storage.Find(context.Background(), "theme2", "name1")

	require.NoError(t, err1, "Error finding template1")
	require.NoError(t, err2, "Error finding template2")
	require.NoError(t, err3, "Error finding template3")

	require.NotNil(t, tmpl1, "Template1 should not be nil")
	require.NotNil(t, tmpl2, "Template2 should not be nil")
	require.NotNil(t, tmpl3, "Template3 should not be nil")

	assert.NotEqual(t, tmpl1.Content(), tmpl2.Content(), "Different templates should have different content")
	assert.NotEqual(t, tmpl1.Content(), tmpl3.Content(), "Templates with same name but different themes should have different content")
}

func TestStorageMemory_OverwriteTemplate(t *testing.T) {
	storage := NewStorageMemory()

	// Add initial template
	storage.Add("test", "example.html", "<div>Original</div>")

	// Verify original content
	tmpl, err := storage.Find(context.Background(), "test", "example.html")
	require.NoError(t, err, "Error finding original template")
	require.NotNil(t, tmpl, "Expected template to be found")

	assert.Equal(t, "<div>Original</div>", tmpl.Content(), "Expected original content")

	// Overwrite with new content
	storage.Add("test", "example.html", "<div>Updated</div>")

	// Verify updated content
	tmpl, err = storage.Find(context.Background(), "test", "example.html")
	require.NoError(t, err, "Error finding updated template")
	require.NotNil(t, tmpl, "Expected template to be found after update")

	assert.Equal(t, "<div>Updated</div>", tmpl.Content(), "Expected updated content")
}

func TestStorageMemory_ComplexTemplateContent(t *testing.T) {
	storage := NewStorageMemory()

	complexContent := `<!-- layouts/base -->
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    {{define "content"}}{{end}}
</body>
</html>`

	storage.Add("complex", "base.html", complexContent)

	tmpl, err := storage.Find(context.Background(), "complex", "base.html")
	require.NoError(t, err, "Error finding complex template")
	require.NotNil(t, tmpl, "Expected template to be found")

	// Verify path extraction from HTML comment
	expectedPath := "layouts/base"
	assert.Equal(t, expectedPath, tmpl.Path(), "Expected path to match")

	// Verify content has comment removed
	expectedContent := `
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    {{define "content"}}{{end}}
</body>
</html>`

	assert.Equal(t, expectedContent, tmpl.Content(), "Content mismatch")
}

func TestStorageMemory_Performance(t *testing.T) {
	storage := NewStorageMemory()

	// Add performance test
	numTemplates := 1000
	start := time.Now()

	for i := 0; i < numTemplates; i++ {
		templateName := fmt.Sprintf("template_%d.html", i)
		content := fmt.Sprintf("<div>Template content %d</div>", i)
		storage.Add("perf", templateName, content)
	}

	addDuration := time.Since(start)

	// Find performance test
	start = time.Now()

	for i := 0; i < numTemplates; i++ {
		templateName := fmt.Sprintf("template_%d.html", i)
		_, err := storage.Find(context.Background(), "perf", templateName)
		assert.NoError(t, err, "Error finding template %d", i)
	}

	findDuration := time.Since(start)

	t.Logf("Added %d templates in %v (%.2f templates/sec)",
		numTemplates, addDuration, float64(numTemplates)/addDuration.Seconds())
	t.Logf("Found %d templates in %v (%.2f templates/sec)",
		numTemplates, findDuration, float64(numTemplates)/findDuration.Seconds())

	// Simple performance sanity checks
	assert.Less(t, addDuration, time.Second, "Add operation took too long: %v", addDuration)
	assert.Less(t, findDuration, time.Second, "Find operation took too long: %v", findDuration)
}
