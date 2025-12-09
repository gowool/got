package got

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockStorage is a mock implementation of the Storage interface
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Find(ctx context.Context, theme, name string) (Template, error) {
	args := m.Called(ctx, theme, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(Template), args.Error(1)
}

func TestNewStorageChain(t *testing.T) {
	tests := []struct {
		name     string
		storages []Storage
		wantLen  int
	}{
		{
			name:     "empty storage chain",
			storages: []Storage{},
			wantLen:  0,
		},
		{
			name:     "single storage",
			storages: []Storage{&MockStorage{}},
			wantLen:  1,
		},
		{
			name:     "multiple storages",
			storages: []Storage{&MockStorage{}, &MockStorage{}, &MockStorage{}},
			wantLen:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := NewStorageChain(tt.storages...)
			require.NotNil(t, chain, "NewStorageChain() should not return nil")
			assert.Len(t, chain.storages, tt.wantLen, "Expected storage chain length to match")
		})
	}
}

func TestStorageChain_Add(t *testing.T) {
	chain := NewStorageChain()

	// Test adding to empty chain
	mockStorage := &MockStorage{}
	chain.Add(mockStorage)
	assert.Len(t, chain.storages, 1, "Chain should have 1 storage after adding")

	// Test adding multiple storages
	mockStorage2 := &MockStorage{}
	chain.Add(mockStorage2)
	assert.Len(t, chain.storages, 2, "Chain should have 2 storages after adding second")

	// Verify storages are added in order
	assert.Same(t, mockStorage, chain.storages[0], "First storage should be the first one added")
	assert.Same(t, mockStorage2, chain.storages[1], "Second storage should be the second one added")
}

func TestStorageChain_Find_Success(t *testing.T) {
	// Create mock storages
	mockStorage1 := &MockStorage{}
	mockStorage2 := &MockStorage{}
	mockStorage3 := &MockStorage{}

	// Create a test template
	template := newTemplate("default", "home.html", "<div>Home</div>")

	// Setup expectations - first storage fails, second succeeds
	mockStorage1.On("Find", mock.Anything, "default", "home.html").Return(nil, ErrTemplateNotFound)
	mockStorage2.On("Find", mock.Anything, "default", "home.html").Return(template, nil)
	mockStorage3.On("Find", mock.Anything, "default", "home.html").Return(nil, ErrTemplateNotFound).Maybe()

	// Create chain with storages
	chain := NewStorageChain(mockStorage1, mockStorage2, mockStorage3)

	// Test Find
	ctx := context.Background()
	result, err := chain.Find(ctx, "default", "home.html")

	// Assertions
	assert.NoError(t, err, "Find should not return an error")
	assert.NotNil(t, result, "Find should return a template")
	assert.Equal(t, template, result, "Should return the template from the second storage")

	// Verify all mocks were called according to expectations
	mockStorage1.AssertExpectations(t)
	mockStorage2.AssertExpectations(t)
}

func TestStorageChain_Find_FirstStorageSuccess(t *testing.T) {
	// Create mock storages
	mockStorage1 := &MockStorage{}
	mockStorage2 := &MockStorage{}

	// Create a test template
	template := newTemplate("admin", "dashboard.html", "<div>Dashboard</div>")

	// Setup expectations - first storage succeeds
	mockStorage1.On("Find", mock.Anything, "admin", "dashboard.html").Return(template, nil)

	// Create chain with storages
	chain := NewStorageChain(mockStorage1, mockStorage2)

	// Test Find
	ctx := context.Background()
	result, err := chain.Find(ctx, "admin", "dashboard.html")

	// Assertions
	assert.NoError(t, err, "Find should not return an error")
	assert.NotNil(t, result, "Find should return a template")
	assert.Equal(t, template, result, "Should return the template from the first storage")

	// Verify only the first storage was called
	mockStorage1.AssertExpectations(t)
	mockStorage2.AssertNotCalled(t, "Find")
}

func TestStorageChain_Find_NotFound(t *testing.T) {
	// Create mock storages
	mockStorage1 := &MockStorage{}
	mockStorage2 := &MockStorage{}
	mockStorage3 := &MockStorage{}

	// Setup expectations - all storages fail
	mockStorage1.On("Find", mock.Anything, "missing", "template.html").Return(nil, ErrTemplateNotFound)
	mockStorage2.On("Find", mock.Anything, "missing", "template.html").Return(nil, ErrTemplateNotFound)
	mockStorage3.On("Find", mock.Anything, "missing", "template.html").Return(nil, ErrTemplateNotFound)

	// Create chain with storages
	chain := NewStorageChain(mockStorage1, mockStorage2, mockStorage3)

	// Test Find
	ctx := context.Background()
	result, err := chain.Find(ctx, "missing", "template.html")

	// Assertions
	assert.Error(t, err, "Find should return an error when template not found")
	assert.Nil(t, result, "Find should return nil when template not found")
	assert.ErrorIs(t, err, ErrTemplateNotFound, "Error should be ErrTemplateNotFound")
	assert.Contains(t, err.Error(), "storage chain: template missing/template.html not found", "Error message should contain template info")

	// Verify all storages were called
	mockStorage1.AssertExpectations(t)
	mockStorage2.AssertExpectations(t)
	mockStorage3.AssertExpectations(t)
}

func TestStorageChain_Find_NonTemplateNotFoundErrors(t *testing.T) {
	// Create mock storages
	mockStorage1 := &MockStorage{}
	mockStorage2 := &MockStorage{}

	// Create a custom error
	customErr := errors.New("database connection failed")

	// Setup expectations - first storage returns non-ErrTemplateNotFound error
	mockStorage1.On("Find", mock.Anything, "test", "error.html").Return(nil, customErr)

	// Create chain with storages
	chain := NewStorageChain(mockStorage1, mockStorage2)

	// Test Find
	ctx := context.Background()
	result, err := chain.Find(ctx, "test", "error.html")

	// Assertions
	assert.Error(t, err, "Find should return an error when storage returns non-ErrTemplateNotFound error")
	assert.Nil(t, result, "Find should return nil when storage returns error")
	assert.Equal(t, customErr, err, "Should return the original error from storage")

	// Verify only the first storage was called (error should stop the chain)
	mockStorage1.AssertExpectations(t)
	mockStorage2.AssertNotCalled(t, "Find")
}

func TestStorageChain_Find_EmptyChain(t *testing.T) {
	// Create empty chain
	chain := NewStorageChain()

	// Test Find
	ctx := context.Background()
	result, err := chain.Find(ctx, "any", "template.html")

	// Assertions
	assert.Error(t, err, "Find should return an error when chain is empty")
	assert.Nil(t, result, "Find should return nil when chain is empty")
	assert.ErrorIs(t, err, ErrTemplateNotFound, "Error should be ErrTemplateNotFound")
	assert.Contains(t, err.Error(), "storage chain: template any/template.html not found", "Error message should contain template info")
}

func TestStorageChain_Find_MixedSuccessAndErrors(t *testing.T) {
	// Create mock storages
	mockStorage1 := &MockStorage{}
	mockStorage2 := &MockStorage{}
	mockStorage3 := &MockStorage{}
	mockStorage4 := &MockStorage{}

	// Create a test template
	template := newTemplate("mixed", "test.html", "<div>Mixed</div>")

	// Setup expectations - ErrTemplateNotFound, then custom error, then success
	mockStorage1.On("Find", mock.Anything, "mixed", "test.html").Return(nil, ErrTemplateNotFound)
	mockStorage2.On("Find", mock.Anything, "mixed", "test.html").Return(nil, errors.New("temporary failure"))
	mockStorage3.On("Find", mock.Anything, "mixed", "test.html").Return(template, nil)

	// Create chain with storages
	chain := NewStorageChain(mockStorage1, mockStorage2, mockStorage3, mockStorage4)

	// Test Find
	ctx := context.Background()
	result, err := chain.Find(ctx, "mixed", "test.html")

	// Assertions
	assert.Error(t, err, "Find should return an error when encountering non-ErrTemplateNotFound error")
	assert.Nil(t, result, "Find should return nil when encountering non-ErrTemplateNotFound error")
	assert.Equal(t, errors.New("temporary failure"), err, "Should return the non-ErrTemplateNotFound error")

	// Verify only the first two storages were called (error stops the chain)
	mockStorage1.AssertExpectations(t)
	mockStorage2.AssertExpectations(t)
	mockStorage3.AssertNotCalled(t, "Find")
	mockStorage4.AssertNotCalled(t, "Find")
}

func TestStorageChain_Find_ContextPassing(t *testing.T) {
	// Create mock storage
	mockStorage := &MockStorage{}
	template := newTemplate("test", "ctx.html", "<div>Context</div>")

	// Create chain with storage
	chain := NewStorageChain(mockStorage)

	// Test with different contexts
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "background context",
			ctx:  context.Background(),
		},
		{
			name: "TODO context",
			ctx:  context.TODO(),
		},
		{
			name: "context with value",
			ctx:  context.WithValue(context.Background(), "key", "value"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup expectation
			mockStorage.On("Find", tt.ctx, "test", "ctx.html").Return(template, nil).Once()

			// Test Find
			result, err := chain.Find(tt.ctx, "test", "ctx.html")

			// Assertions
			assert.NoError(t, err, "Find should not return an error")
			assert.NotNil(t, result, "Find should return a template")
			assert.Equal(t, template, result, "Should return the expected template")

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestStorageChain_ConcurrentFind(t *testing.T) {
	// Test concurrent access to the same storage chain
	t.Run("concurrent access to same template", func(t *testing.T) {
		// Create a fresh chain for this test
		freshMock1 := &MockStorage{}
		freshChain := NewStorageChain(freshMock1)

		// Create template
		template := newTemplate("test", "template.html", "<div>Template</div>")

		// Setup expectation - the same template should be found by both goroutines
		freshMock1.On("Find", mock.Anything, "test", "template.html").Return(template, nil).Twice()

		done := make(chan bool, 2)
		var results [2]Template
		var errors [2]error

		// Goroutine 1
		go func() {
			defer func() { done <- true }()
			result, err := freshChain.Find(context.Background(), "test", "template.html")
			results[0] = result
			errors[0] = err
		}()

		// Goroutine 2
		go func() {
			defer func() { done <- true }()
			result, err := freshChain.Find(context.Background(), "test", "template.html")
			results[1] = result
			errors[1] = err
		}()

		// Wait for both goroutines
		for i := 0; i < 2; i++ {
			<-done
		}

		// Both should succeed with the same template
		for i := 0; i < 2; i++ {
			assert.NoError(t, errors[i], "Goroutine %d should not return error", i)
			assert.Equal(t, template, results[i], "Goroutine %d should get the same template", i)
		}

		freshMock1.AssertExpectations(t)
	})
}

func TestStorageChain_StorageInterface(t *testing.T) {
	// Verify that StorageChain implements the Storage interface
	var _ Storage = (*StorageChain)(nil)

	// Test with real storage implementations
	storage1 := NewStorageMemory()
	storage2 := NewStorageMemory()

	// Add templates to storages
	storage1.Add("theme1", "template1.html", "<div>Template 1 from Storage 1</div>")
	storage2.Add("theme2", "template2.html", "<div>Template 2 from Storage 2</div>")
	storage1.Add("theme1", "template3.html", "<div>Template 3 from Storage 1</div>")

	// Create chain
	chain := NewStorageChain(storage1, storage2)

	// Test finding template from first storage
	tmpl1, err := chain.Find(context.Background(), "theme1", "template1.html")
	assert.NoError(t, err)
	assert.NotNil(t, tmpl1)
	assert.Equal(t, "theme1", tmpl1.Theme())
	assert.Equal(t, "template1.html", tmpl1.Name())

	// Test finding template from second storage
	tmpl2, err := chain.Find(context.Background(), "theme2", "template2.html")
	assert.NoError(t, err)
	assert.NotNil(t, tmpl2)
	assert.Equal(t, "theme2", tmpl2.Theme())
	assert.Equal(t, "template2.html", tmpl2.Name())

	// Test finding non-existent template
	_, err = chain.Find(context.Background(), "missing", "template.html")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTemplateNotFound)
}
