package internal

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "empty byte slice",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "nil byte slice",
			input:    nil,
			expected: "",
		},
		{
			name:     "simple string",
			input:    []byte("hello"),
			expected: "hello",
		},
		{
			name:     "string with spaces",
			input:    []byte("hello world"),
			expected: "hello world",
		},
		{
			name:     "string with special characters",
			input:    []byte("hello\tworld\n"),
			expected: "hello\tworld\n",
		},
		{
			name:     "string with unicode",
			input:    []byte("héllo 🌍"),
			expected: "héllo 🌍",
		},
		{
			name:     "single byte",
			input:    []byte("a"),
			expected: "a",
		},
		{
			name:     "long string",
			input:    []byte("this is a very long string that contains many words and characters to test the performance and correctness of the unsafe.String conversion"),
			expected: "this is a very long string that contains many words and characters to test the performance and correctness of the unsafe.String conversion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := String(tt.input)
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestStringMemorySafety(t *testing.T) {
	// Test that the String function works correctly with unsafe operations
	t.Run("unsafe behavior verification", func(t *testing.T) {
		// Create a byte slice
		original := []byte("test data")

		// Convert to string using our function
		str := String(original)

		// Verify the string is correct
		if str != "test data" {
			t.Errorf("String() = %q, want %q", str, "test data")
		}

		// Note: With unsafe.String, the created string shares the same memory
		// as the original byte slice, so modifying the slice will affect the string
		// This is the expected behavior for unsafe operations
		original[0] = 'T'
		if str != "Test data" {
			t.Errorf("String should reflect changes to underlying memory, got %q", str)
		}
	})

	t.Run("slice modification after conversion", func(t *testing.T) {
		// Test that we can handle slices that are modified after conversion
		data := make([]byte, 10)
		copy(data, "hello")

		// Convert while slice has partial data
		str := String(data[:5])
		if str != "hello" {
			t.Errorf("String() = %q, want %q", str, "hello")
		}
	})
}

func TestStringPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Basic performance test to ensure the function is efficient
	t.Run("performance test", func(t *testing.T) {
		// Create a large byte slice
		data := make([]byte, 1024*1024) // 1MB
		for i := range data {
			data[i] = byte(i % 256)
		}

		// Run the conversion multiple times
		iterations := 1000
		for i := 0; i < iterations; i++ {
			_ = String(data)
		}

		t.Logf("Completed %d conversions of 1MB data", iterations)
	})
}

func TestStringWithUnsafePackage(t *testing.T) {
	// Test that our implementation matches the behavior of using unsafe package directly
	t.Run("matches unsafe.String behavior", func(t *testing.T) {
		testData := []byte("comparison test")

		// Our implementation
		result1 := String(testData)

		// Direct unsafe package usage
		result2 := unsafe.String(unsafe.SliceData(testData), len(testData))

		if result1 != result2 {
			t.Errorf("String() = %q, unsafe.String() = %q", result1, result2)
		}
	})
}

func TestStringEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantErr  bool
		expected string
	}{
		{
			name:     "zero-length slice",
			input:    make([]byte, 0),
			expected: "",
		},
		{
			name:     "slice with capacity but no length",
			input:    make([]byte, 0, 10),
			expected: "",
		},
		{
			name:     "slice with null bytes",
			input:    []byte("hello\x00world"),
			expected: "hello\x00world",
		},
		{
			name:     "all null bytes",
			input:    []byte{0, 0, 0, 0},
			expected: "\x00\x00\x00\x00",
		},
		{
			name:     "high unicode characters",
			input:    []byte("🚀🌟💫✨"),
			expected: "🚀🌟💫✨",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := String(tt.input)
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func BenchmarkString(b *testing.B) {
	// Benchmark different sizes of byte slices
	sizes := []int{0, 1, 10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = String(data)
			}
		})
	}
}

func TestStringZeroValue(t *testing.T) {
	// Test that the function handles zero values correctly
	var nilSlice []byte
	result := String(nilSlice)

	if result != "" {
		t.Errorf("String(nil) = %q, want empty string", result)
	}
}
