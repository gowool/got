package got

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStorageFS(t *testing.T) {
	tests := []struct {
		name string
		fsys fstest.MapFS
	}{
		{
			name: "valid filesystem",
			fsys: fstest.MapFS{},
		},
		{
			name: "filesystem with templates",
			fsys: fstest.MapFS{
				"default/home.html": &fstest.MapFile{
					Data: []byte("<div>Home</div>"),
				},
			},
		},
		{
			name: "empty filesystem",
			fsys: fstest.MapFS{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewStorageFS(tt.fsys)
			require.NotNil(t, storage, "NewStorageFS() should not return nil")

			// Verify it implements Storage interface
			var _ Storage = storage
		})
	}
}

func TestStorageFS_Find_Success(t *testing.T) {
	// Create a mock filesystem with various template structures
	fsys := fstest.MapFS{
		"default/home.html": &fstest.MapFile{
			Data: []byte("<div>Home page</div>"),
		},
		"default/about.html": &fstest.MapFile{
			Data: []byte("<!-- layouts/base -->\n<div>About page</div>"),
		},
		"admin/dashboard.html": &fstest.MapFile{
			Data: []byte("<!-- layouts/admin -->\n<div>Dashboard</div>"),
		},
		"blog/post.html": &fstest.MapFile{
			Data: []byte(`<!-- layouts/blog -->
{{define "content"}}<article>Post content</article>{{end}}`),
		},
		"nested/deep/path/template.html": &fstest.MapFile{
			Data: []byte("<div>Deep nested template</div>"),
		},
		"empty/template.html": &fstest.MapFile{
			Data: []byte(""),
		},
	}

	storage := NewStorageFS(fsys)

	tests := []struct {
		name            string
		theme           string
		template        string
		expectedTheme   string
		expectedName    string
		expectedPath    string
		expectedContent string
	}{
		{
			name:            "simple template without comment",
			theme:           "default",
			template:        "home.html",
			expectedTheme:   "default",
			expectedName:    "home.html",
			expectedPath:    "home.html",
			expectedContent: "<div>Home page</div>",
		},
		{
			name:            "template with HTML comment path",
			theme:           "default",
			template:        "about.html",
			expectedTheme:   "default",
			expectedName:    "about.html",
			expectedPath:    "layouts/base",
			expectedContent: "\n<div>About page</div>",
		},
		{
			name:            "admin template with comment",
			theme:           "admin",
			template:        "dashboard.html",
			expectedTheme:   "admin",
			expectedName:    "dashboard.html",
			expectedPath:    "layouts/admin",
			expectedContent: "\n<div>Dashboard</div>",
		},
		{
			name:          "blog template with define blocks",
			theme:         "blog",
			template:      "post.html",
			expectedTheme: "blog",
			expectedName:  "post.html",
			expectedPath:  "layouts/blog",
			expectedContent: `
{{define "content"}}<article>Post content</article>{{end}}`,
		},
		{
			name:            "deeply nested template",
			theme:           "nested",
			template:        "deep/path/template.html",
			expectedTheme:   "nested",
			expectedName:    "deep/path/template.html",
			expectedPath:    "deep/path/template.html",
			expectedContent: "<div>Deep nested template</div>",
		},
		{
			name:            "empty template",
			theme:           "empty",
			template:        "template.html",
			expectedTheme:   "empty",
			expectedName:    "template.html",
			expectedPath:    "template.html",
			expectedContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := storage.Find(context.Background(), tt.theme, tt.template)

			assert.NoError(t, err, "Find() should not return an error")
			require.NotNil(t, template, "Find() should return a template")

			assert.Equal(t, tt.expectedTheme, template.Theme(), "Template theme should match")
			assert.Equal(t, tt.expectedName, template.Name(), "Template name should match")
			assert.Equal(t, tt.expectedPath, template.Path(), "Template path should match")
			assert.Equal(t, tt.expectedContent, template.Content(), "Template content should match")
		})
	}
}

func TestStorageFS_Find_WithContext(t *testing.T) {
	fsys := fstest.MapFS{
		"test/example.html": &fstest.MapFile{
			Data: []byte("<div>Test content</div>"),
		},
	}

	storage := NewStorageFS(fsys)

	// Test with context
	ctx := context.Background()
	template, err := storage.Find(ctx, "test", "example.html")

	assert.NoError(t, err, "Find() with context should not return an error")
	assert.NotNil(t, template, "Find() with context should return a template")

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	template, err = storage.Find(ctx, "test", "example.html")

	// Note: Current implementation doesn't check context cancellation,
	// but this test ensures it works with cancelled contexts
	assert.NoError(t, err, "Find() with cancelled context should not return an error")
	assert.NotNil(t, template, "Find() with cancelled context should return a template")
}

func TestStorageFS_Find_ErrorCases(t *testing.T) {
	fsys := fstest.MapFS{
		"default/home.html": &fstest.MapFile{
			Data: []byte("<div>Home</div>"),
		},
		"admin/dashboard.html": &fstest.MapFile{
			Data: []byte("<div>Dashboard</div>"),
		},
	}

	storage := NewStorageFS(fsys)

	tests := []struct {
		name     string
		theme    string
		template string
		wantErr  bool
		errIs    error
		errMsg   string
	}{
		{
			name:     "non-existent template in existing theme",
			theme:    "default",
			template: "missing.html",
			wantErr:  true,
			errIs:    ErrTemplateNotFound,
			errMsg:   "storage fs: failed to read template default/missing.html",
		},
		{
			name:     "non-existent theme",
			theme:    "missing",
			template: "home.html",
			wantErr:  true,
			errIs:    ErrTemplateNotFound,
			errMsg:   "storage fs: failed to read template missing/home.html",
		},
		{
			name:     "both theme and template missing",
			theme:    "missing",
			template: "missing.html",
			wantErr:  true,
			errIs:    ErrTemplateNotFound,
			errMsg:   "storage fs: failed to read template missing/missing.html",
		},
		{
			name:     "empty theme name",
			theme:    "",
			template: "home.html",
			wantErr:  true,
			errMsg:   "invalid argument",
		},
		{
			name:     "empty template name",
			theme:    "default",
			template: "",
			wantErr:  true,
			errMsg:   "invalid argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := storage.Find(context.Background(), tt.theme, tt.template)

			if tt.wantErr {
				assert.Error(t, err, "Find() should return an error")
				assert.Nil(t, template, "Find() should return nil template on error")

				if tt.errIs != nil {
					assert.ErrorIs(t, err, tt.errIs, "Error should be of expected type")
				}

				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
				}
				return
			}

			assert.NoError(t, err, "Find() should not return an error")
			assert.NotNil(t, template, "Find() should return a template")
		})
	}
}

func TestStorageFS_Find_SubFSError(t *testing.T) {
	// Create a mock filesystem that will cause Sub() to fail
	fsys := &failingFS{}

	storage := NewStorageFS(fsys)

	template, err := storage.Find(context.Background(), "test", "template.html")

	assert.Error(t, err, "Find() should return an error when Sub() fails")
	assert.Nil(t, template, "Find() should return nil template when Sub() fails")
	assert.NotErrorIs(t, err, ErrTemplateNotFound, "Error should not be ErrTemplateNotFound for Sub() failure")
}

func TestStorageFS_Find_ReadFileError(t *testing.T) {
	// Create a filesystem that succeeds for Sub() but fails for ReadFile
	fsys := &partialFailingFS{
		subSuccess: true,
	}

	storage := NewStorageFS(fsys)

	template, err := storage.Find(context.Background(), "test", "template.html")

	assert.Error(t, err, "Find() should return an error when ReadFile() fails")
	assert.Nil(t, template, "Find() should return nil template when ReadFile() fails")
	assert.NotErrorIs(t, err, ErrTemplateNotFound, "Error should not be ErrTemplateNotFound for ReadFile() failure")
}

func TestStorageFS_ImplementsInterface(t *testing.T) {
	// Verify that StorageFS implements the Storage interface
	var _ Storage = (*StorageFS)(nil)

	fsys := fstest.MapFS{
		"test/template.html": &fstest.MapFile{
			Data: []byte("<div>Test</div>"),
		},
	}

	storage := NewStorageFS(fsys)

	// All interface methods should be available and not panic
	require.NotPanics(t, func() {
		_, err := storage.Find(context.Background(), "test", "template.html")
		require.NoError(t, err)
	})
}

func TestStorageFS_ConcurrentAccess(t *testing.T) {
	fsys := fstest.MapFS{
		"default/home.html": &fstest.MapFile{
			Data: []byte("<div>Home</div>"),
		},
		"admin/dashboard.html": &fstest.MapFile{
			Data: []byte("<!-- layouts/admin -->\n<div>Dashboard</div>"),
		},
		"blog/post.html": &fstest.MapFile{
			Data: []byte(`{{define "content"}}<article>Post</article>{{end}}`),
		},
	}

	storage := NewStorageFS(fsys)

	// Test concurrent reads
	t.Run("concurrent finds", func(t *testing.T) {
		const numGoroutines = 10
		const numOperations = 50

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				for j := 0; j < numOperations; j++ {
					themes := []string{"default", "admin", "blog"}
					templates := []string{"home.html", "dashboard.html", "post.html"}

					for k, theme := range themes {
						template, err := storage.Find(context.Background(), theme, templates[k])
						assert.NoError(t, err, "Concurrent find failed: %v", err)
						assert.NotNil(t, template, "Template should not be nil")
					}
				}
			}(i)
		}
	})
}

// Helper types for testing error conditions

type failingFS struct{}

func (f *failingFS) Open(name string) (fs.File, error) {
	return nil, errors.New("simulated filesystem error")
}

type partialFailingFS struct {
	subSuccess bool
}

func (f *partialFailingFS) Open(name string) (fs.File, error) {
	if name == "test" && f.subSuccess {
		// Return a mock file system for the theme that will fail on Read
		return &mockFile{}, nil
	}
	return nil, errors.New("filesystem error")
}

type mockFile struct{}

func (m *mockFile) Read([]byte) (int, error) {
	return 0, errors.New("read file error")
}

func (m *mockFile) Stat() (fs.FileInfo, error) {
	return nil, errors.New("stat error")
}

func (m *mockFile) Close() error {
	return nil
}
