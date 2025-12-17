package got

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockStore is a mock implementation of the Store interface
type MockStore struct {
	mock.Mock
}

func (m *MockStore) Find(ctx context.Context, theme, name string) (Template, error) {
	args := m.Called(ctx, theme, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(Template), args.Error(1)
}

func TestNewStoreChain(t *testing.T) {
	tests := []struct {
		name    string
		stores  []Store
		wantLen int
	}{
		{
			name:    "empty store chain",
			stores:  []Store{},
			wantLen: 0,
		},
		{
			name:    "single store",
			stores:  []Store{&MockStore{}},
			wantLen: 1,
		},
		{
			name:    "multiple stores",
			stores:  []Store{&MockStore{}, &MockStore{}, &MockStore{}},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := NewStoreChain(tt.stores...)
			require.NotNil(t, chain, "NewStoreChain() should not return nil")
			assert.Len(t, chain.stores, tt.wantLen, "Expected store chain length to match")
		})
	}
}

func TestStoreChain_Add(t *testing.T) {
	chain := NewStoreChain()

	// Test adding to empty chain
	mockStore := &MockStore{}
	chain.Add(mockStore)
	assert.Len(t, chain.stores, 1, "Chain should have 1 store after adding")

	// Test adding multiple stores
	mockStore2 := &MockStore{}
	chain.Add(mockStore2)
	assert.Len(t, chain.stores, 2, "Chain should have 2 stores after adding second")

	// Verify stores are added in order
	assert.Same(t, mockStore, chain.stores[0], "First store should be the first one added")
	assert.Same(t, mockStore2, chain.stores[1], "Second store should be the second one added")
}

func TestStoreChain_Find_Success(t *testing.T) {
	// Create mock stores
	mockStore1 := &MockStore{}
	mockStore2 := &MockStore{}
	mockStore3 := &MockStore{}

	// Create a test template
	template := newTemplate("default", "home.html", "<div>Home</div>")

	// Setup expectations - first store fails, second succeeds
	mockStore1.On("Find", mock.Anything, "default", "home.html").Return(nil, ErrTemplateNotFound)
	mockStore2.On("Find", mock.Anything, "default", "home.html").Return(template, nil)
	mockStore3.On("Find", mock.Anything, "default", "home.html").Return(nil, ErrTemplateNotFound).Maybe()

	// Create chain with stores
	chain := NewStoreChain(mockStore1, mockStore2, mockStore3)

	// Test Find
	ctx := context.Background()
	result, err := chain.Find(ctx, "default", "home.html")

	// Assertions
	assert.NoError(t, err, "Find should not return an error")
	assert.NotNil(t, result, "Find should return a template")
	assert.Equal(t, template, result, "Should return the template from the second store")

	// Verify all mocks were called according to expectations
	mockStore1.AssertExpectations(t)
	mockStore2.AssertExpectations(t)
}

func TestStoreChain_Find_FirstStoreSuccess(t *testing.T) {
	// Create mock stores
	mockStore1 := &MockStore{}
	mockStore2 := &MockStore{}

	// Create a test template
	template := newTemplate("admin", "dashboard.html", "<div>Dashboard</div>")

	// Setup expectations - first store succeeds
	mockStore1.On("Find", mock.Anything, "admin", "dashboard.html").Return(template, nil)

	// Create chain with stores
	chain := NewStoreChain(mockStore1, mockStore2)

	// Test Find
	ctx := context.Background()
	result, err := chain.Find(ctx, "admin", "dashboard.html")

	// Assertions
	assert.NoError(t, err, "Find should not return an error")
	assert.NotNil(t, result, "Find should return a template")
	assert.Equal(t, template, result, "Should return the template from the first store")

	// Verify only the first store was called
	mockStore1.AssertExpectations(t)
	mockStore2.AssertNotCalled(t, "Find")
}

func TestStoreChain_Find_NotFound(t *testing.T) {
	// Create mock stores
	mockStore1 := &MockStore{}
	mockStore2 := &MockStore{}
	mockStore3 := &MockStore{}

	// Setup expectations - all stores fail
	mockStore1.On("Find", mock.Anything, "missing", "template.html").Return(nil, ErrTemplateNotFound)
	mockStore2.On("Find", mock.Anything, "missing", "template.html").Return(nil, ErrTemplateNotFound)
	mockStore3.On("Find", mock.Anything, "missing", "template.html").Return(nil, ErrTemplateNotFound)

	// Create chain with stores
	chain := NewStoreChain(mockStore1, mockStore2, mockStore3)

	// Test Find
	ctx := context.Background()
	result, err := chain.Find(ctx, "missing", "template.html")

	// Assertions
	assert.Error(t, err, "Find should return an error when template not found")
	assert.Nil(t, result, "Find should return nil when template not found")
	assert.ErrorIs(t, err, ErrTemplateNotFound, "Error should be ErrTemplateNotFound")
	assert.Contains(t, err.Error(), "store chain: template missing/template.html not found", "Error message should contain template info")

	// Verify all stores were called
	mockStore1.AssertExpectations(t)
	mockStore2.AssertExpectations(t)
	mockStore3.AssertExpectations(t)
}

func TestStoreChain_Find_NonTemplateNotFoundErrors(t *testing.T) {
	// Create mock stores
	mockStore1 := &MockStore{}
	mockStore2 := &MockStore{}

	// Create a custom error
	customErr := errors.New("database connection failed")

	// Setup expectations - first store returns non-ErrTemplateNotFound error
	mockStore1.On("Find", mock.Anything, "test", "error.html").Return(nil, customErr)

	// Create chain with stores
	chain := NewStoreChain(mockStore1, mockStore2)

	// Test Find
	ctx := context.Background()
	result, err := chain.Find(ctx, "test", "error.html")

	// Assertions
	assert.Error(t, err, "Find should return an error when store returns non-ErrTemplateNotFound error")
	assert.Nil(t, result, "Find should return nil when store returns error")
	assert.Equal(t, customErr, err, "Should return the original error from store")

	// Verify only the first store was called (error should stop the chain)
	mockStore1.AssertExpectations(t)
	mockStore2.AssertNotCalled(t, "Find")
}

func TestStoreChain_Find_EmptyChain(t *testing.T) {
	// Create empty chain
	chain := NewStoreChain()

	// Test Find
	ctx := context.Background()
	result, err := chain.Find(ctx, "any", "template.html")

	// Assertions
	assert.Error(t, err, "Find should return an error when chain is empty")
	assert.Nil(t, result, "Find should return nil when chain is empty")
	assert.ErrorIs(t, err, ErrTemplateNotFound, "Error should be ErrTemplateNotFound")
	assert.Contains(t, err.Error(), "store chain: template any/template.html not found", "Error message should contain template info")
}

func TestStoreChain_Find_MixedSuccessAndErrors(t *testing.T) {
	// Create mock stores
	mockStore1 := &MockStore{}
	mockStore2 := &MockStore{}
	mockStore3 := &MockStore{}
	mockStore4 := &MockStore{}

	// Create a test template
	template := newTemplate("mixed", "test.html", "<div>Mixed</div>")

	// Setup expectations - ErrTemplateNotFound, then custom error, then success
	mockStore1.On("Find", mock.Anything, "mixed", "test.html").Return(nil, ErrTemplateNotFound)
	mockStore2.On("Find", mock.Anything, "mixed", "test.html").Return(nil, errors.New("temporary failure"))
	mockStore3.On("Find", mock.Anything, "mixed", "test.html").Return(template, nil)

	// Create chain with stores
	chain := NewStoreChain(mockStore1, mockStore2, mockStore3, mockStore4)

	// Test Find
	ctx := context.Background()
	result, err := chain.Find(ctx, "mixed", "test.html")

	// Assertions
	assert.Error(t, err, "Find should return an error when encountering non-ErrTemplateNotFound error")
	assert.Nil(t, result, "Find should return nil when encountering non-ErrTemplateNotFound error")
	assert.Equal(t, errors.New("temporary failure"), err, "Should return the non-ErrTemplateNotFound error")

	// Verify only the first two stores were called (error stops the chain)
	mockStore1.AssertExpectations(t)
	mockStore2.AssertExpectations(t)
	mockStore3.AssertNotCalled(t, "Find")
	mockStore4.AssertNotCalled(t, "Find")
}

func TestStoreChain_Find_ContextPassing(t *testing.T) {
	// Create mock store
	mockStore := &MockStore{}
	template := newTemplate("test", "ctx.html", "<div>Context</div>")

	// Create chain with store
	chain := NewStoreChain(mockStore)

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
			ctx:  context.WithValue(context.Background(), "key", "value"), //nolint:staticcheck
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup expectation
			mockStore.On("Find", tt.ctx, "test", "ctx.html").Return(template, nil).Once()

			// Test Find
			result, err := chain.Find(tt.ctx, "test", "ctx.html")

			// Assertions
			assert.NoError(t, err, "Find should not return an error")
			assert.NotNil(t, result, "Find should return a template")
			assert.Equal(t, template, result, "Should return the expected template")

			mockStore.AssertExpectations(t)
		})
	}
}

func TestStoreChain_ConcurrentFind(t *testing.T) {
	// Test concurrent access to the same store chain
	t.Run("concurrent access to same template", func(t *testing.T) {
		// Create a fresh chain for this test
		freshMock1 := &MockStore{}
		freshChain := NewStoreChain(freshMock1)

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

func TestStoreChain_StoreInterface(t *testing.T) {
	// Verify that StoreChain implements the Store interface
	var _ Store = (*StoreChain)(nil)

	// Test with real store implementations
	store1 := NewStoreMemory()
	store2 := NewStoreMemory()

	// Add templates to stores
	store1.Add("theme1", "template1.html", "<div>Template 1 from Store 1</div>")
	store2.Add("theme2", "template2.html", "<div>Template 2 from Store 2</div>")
	store1.Add("theme1", "template3.html", "<div>Template 3 from Store 1</div>")

	// Create chain
	chain := NewStoreChain(store1, store2)

	// Test finding template from first store
	tmpl1, err := chain.Find(context.Background(), "theme1", "template1.html")
	assert.NoError(t, err)
	assert.NotNil(t, tmpl1)
	assert.Equal(t, "theme1", tmpl1.Theme())
	assert.Equal(t, "template1.html", tmpl1.Name())

	// Test finding template from second store
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
