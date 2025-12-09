package internal

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoArithmetic_IntegerOperations(t *testing.T) {
	tests := []struct {
		name     string
		a        any
		b        any
		op       rune
		expected any
	}{
		// Addition tests
		{
			name:     "int addition",
			a:        5,
			b:        3,
			op:       '+',
			expected: int64(8),
		},
		{
			name:     "int8 addition",
			a:        int8(5),
			b:        int8(3),
			op:       '+',
			expected: int64(8),
		},
		{
			name:     "int16 addition",
			a:        int16(5),
			b:        int16(3),
			op:       '+',
			expected: int64(8),
		},
		{
			name:     "int32 addition",
			a:        int32(5),
			b:        int32(3),
			op:       '+',
			expected: int64(8),
		},
		{
			name:     "int64 addition",
			a:        int64(5),
			b:        int64(3),
			op:       '+',
			expected: int64(8),
		},
		{
			name:     "negative int addition",
			a:        -5,
			b:        3,
			op:       '+',
			expected: int64(-2),
		},
		{
			name:     "zero addition",
			a:        0,
			b:        5,
			op:       '+',
			expected: int64(5),
		},

		// Subtraction tests
		{
			name:     "int subtraction",
			a:        10,
			b:        3,
			op:       '-',
			expected: int64(7),
		},
		{
			name:     "negative subtraction",
			a:        5,
			b:        10,
			op:       '-',
			expected: int64(-5),
		},
		{
			name:     "zero subtraction",
			a:        0,
			b:        5,
			op:       '-',
			expected: int64(-5),
		},

		// Multiplication tests
		{
			name:     "int multiplication",
			a:        4,
			b:        3,
			op:       '*',
			expected: int64(12),
		},
		{
			name:     "negative multiplication",
			a:        -4,
			b:        3,
			op:       '*',
			expected: int64(-12),
		},
		{
			name:     "zero multiplication",
			a:        0,
			b:        5,
			op:       '*',
			expected: int64(0),
		},

		// Division tests
		{
			name:     "int division",
			a:        12,
			b:        3,
			op:       '/',
			expected: int64(4),
		},
		{
			name:     "negative division",
			a:        -12,
			b:        3,
			op:       '/',
			expected: int64(-4),
		},
		{
			name:     "fractional division truncates",
			a:        7,
			b:        2,
			op:       '/',
			expected: int64(3),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DoArithmetic(tt.a, tt.b, tt.op)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDoArithmetic_FloatOperations(t *testing.T) {
	tests := []struct {
		name     string
		a        any
		b        any
		op       rune
		expected any
	}{
		// Addition tests
		{
			name:     "float32 addition",
			a:        float32(5.5),
			b:        float32(3.2),
			op:       '+',
			expected: 8.7,
		},
		{
			name:     "float64 addition",
			a:        5.5,
			b:        3.2,
			op:       '+',
			expected: 8.7,
		},
		{
			name:     "negative float addition",
			a:        -5.5,
			b:        3.2,
			op:       '+',
			expected: -2.3,
		},

		// Subtraction tests
		{
			name:     "float subtraction",
			a:        10.5,
			b:        3.2,
			op:       '-',
			expected: 7.299999999999999,
		},
		{
			name:     "negative float subtraction",
			a:        5.5,
			b:        10.2,
			op:       '-',
			expected: -4.699999999999999,
		},

		// Multiplication tests
		{
			name:     "float multiplication",
			a:        4.5,
			b:        3.0,
			op:       '*',
			expected: 13.5,
		},
		{
			name:     "negative float multiplication",
			a:        -4.5,
			b:        3.0,
			op:       '*',
			expected: -13.5,
		},

		// Division tests
		{
			name:     "float division",
			a:        12.0,
			b:        3.0,
			op:       '/',
			expected: 4.0,
		},
		{
			name:     "fractional float division",
			a:        7.0,
			b:        2.0,
			op:       '/',
			expected: 3.5,
		},
		{
			name:     "negative float division",
			a:        -12.0,
			b:        3.0,
			op:       '/',
			expected: -4.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DoArithmetic(tt.a, tt.b, tt.op)
			require.NoError(t, err)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}

func TestDoArithmetic_UintOperations(t *testing.T) {
	tests := []struct {
		name     string
		a        any
		b        any
		op       rune
		expected any
	}{
		// Addition tests
		{
			name:     "uint addition",
			a:        uint(5),
			b:        uint(3),
			op:       '+',
			expected: uint64(8),
		},
		{
			name:     "uint8 addition",
			a:        uint8(5),
			b:        uint8(3),
			op:       '+',
			expected: uint64(8),
		},
		{
			name:     "uint16 addition",
			a:        uint16(5),
			b:        uint16(3),
			op:       '+',
			expected: uint64(8),
		},
		{
			name:     "uint32 addition",
			a:        uint32(5),
			b:        uint32(3),
			op:       '+',
			expected: uint64(8),
		},
		{
			name:     "uint64 addition",
			a:        uint64(5),
			b:        uint64(3),
			op:       '+',
			expected: uint64(8),
		},

		// Subtraction tests
		{
			name:     "uint subtraction",
			a:        uint(10),
			b:        uint(3),
			op:       '-',
			expected: uint64(7),
		},

		// Multiplication tests
		{
			name:     "uint multiplication",
			a:        uint(4),
			b:        uint(3),
			op:       '*',
			expected: uint64(12),
		},

		// Division tests
		{
			name:     "uint division",
			a:        uint(12),
			b:        uint(3),
			op:       '/',
			expected: uint64(4),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DoArithmetic(tt.a, tt.b, tt.op)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDoArithmetic_MixedTypeOperations(t *testing.T) {
	tests := []struct {
		name     string
		a        any
		b        any
		op       rune
		expected any
	}{
		// Int with Float
		{
			name:     "int + float",
			a:        5,
			b:        3.5,
			op:       '+',
			expected: 8.5,
		},
		{
			name:     "float + int",
			a:        5.5,
			b:        3,
			op:       '+',
			expected: 8.5,
		},
		{
			name:     "int - float",
			a:        10,
			b:        3.5,
			op:       '-',
			expected: 6.5,
		},
		{
			name:     "int * float",
			a:        4,
			b:        2.5,
			op:       '*',
			expected: 10.0,
		},
		{
			name:     "int / float",
			a:        10,
			b:        2.0,
			op:       '/',
			expected: 5.0,
		},

		// Int with Uint (positive values)
		{
			name:     "positive int + uint",
			a:        5,
			b:        uint(3),
			op:       '+',
			expected: uint64(8),
		},
		{
			name:     "positive int - uint",
			a:        10,
			b:        uint(3),
			op:       '-',
			expected: uint64(7),
		},
		{
			name:     "positive int * uint",
			a:        4,
			b:        uint(3),
			op:       '*',
			expected: uint64(12),
		},
		{
			name:     "positive int / uint",
			a:        12,
			b:        uint(3),
			op:       '/',
			expected: uint64(4),
		},

		// Float with Uint
		{
			name:     "float + uint",
			a:        5.5,
			b:        uint(3),
			op:       '+',
			expected: 8.5,
		},
		{
			name:     "float - uint",
			a:        10.5,
			b:        uint(3),
			op:       '-',
			expected: 7.5,
		},
		{
			name:     "float * uint",
			a:        4.5,
			b:        uint(2),
			op:       '*',
			expected: 9.0,
		},
		{
			name:     "float / uint",
			a:        9.0,
			b:        uint(3),
			op:       '/',
			expected: 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DoArithmetic(tt.a, tt.b, tt.op)
			require.NoError(t, err)

			// For float comparisons, use InDelta
			if _, ok := tt.expected.(float64); ok {
				assert.InDelta(t, tt.expected, result, 0.0001)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDoArithmetic_StringOperations(t *testing.T) {
	tests := []struct {
		name     string
		a        any
		b        any
		op       rune
		expected any
		wantErr  bool
	}{
		{
			name:     "string concatenation",
			a:        "hello",
			b:        " world",
			op:       '+',
			expected: "hello world",
			wantErr:  false,
		},
		{
			name:     "empty string concatenation",
			a:        "",
			b:        "test",
			op:       '+',
			expected: "test",
			wantErr:  false,
		},
		{
			name:    "string subtraction",
			a:       "hello",
			b:       " world",
			op:      '-',
			wantErr: true,
		},
		{
			name:    "string multiplication",
			a:       "hello",
			b:       " world",
			op:      '*',
			wantErr: true,
		},
		{
			name:    "string division",
			a:       "hello",
			b:       " world",
			op:      '/',
			wantErr: true,
		},
		{
			name:    "string + non-string",
			a:       "hello",
			b:       5,
			op:      '+',
			wantErr: true,
		},
		{
			name:    "non-string + string",
			a:       5,
			b:       "hello",
			op:      '+',
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DoArithmetic(tt.a, tt.b, tt.op)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDoArithmetic_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		a           any
		b           any
		op          rune
		expectedErr string
	}{
		// Division by zero
		{
			name:        "int division by zero",
			a:           5,
			b:           0,
			op:          '/',
			expectedErr: "can't divide the value by 0",
		},
		{
			name:        "float division by zero",
			a:           5.0,
			b:           0.0,
			op:          '/',
			expectedErr: "can't divide the value by 0",
		},
		{
			name:        "uint division by zero",
			a:           uint(5),
			b:           uint(0),
			op:          '/',
			expectedErr: "can't divide the value by 0",
		},

		// Invalid operations
		{
			name:        "unsupported operator",
			a:           5,
			b:           3,
			op:          '%',
			expectedErr: "there is no such an operation",
		},

		// Type compatibility errors
		{
			name:        "int + bool",
			a:           5,
			b:           true,
			op:          '+',
			expectedErr: "can't apply the operator to the values",
		},
		{
			name:        "float + slice",
			a:           5.5,
			b:           []int{1, 2, 3},
			op:          '+',
			expectedErr: "can't apply the operator to the values",
		},
		{
			name:        "uint + map",
			a:           uint(5),
			b:           map[string]int{"key": 1},
			op:          '+',
			expectedErr: "can't apply the operator to the values",
		},

		// Unsupported types
		{
			name:        "slice + slice",
			a:           []int{1, 2},
			b:           []int{3, 4},
			op:          '+',
			expectedErr: "can't apply the operator to the values",
		},
		{
			name:        "map + map",
			a:           map[string]int{"a": 1},
			b:           map[string]int{"b": 2},
			op:          '+',
			expectedErr: "can't apply the operator to the values",
		},
		{
			name:        "bool + bool",
			a:           true,
			b:           false,
			op:          '+',
			expectedErr: "can't apply the operator to the values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DoArithmetic(tt.a, tt.b, tt.op)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestDoArithmetic_EdgeCases(t *testing.T) {
	t.Run("negative int with uint", func(t *testing.T) {
		// Negative int + positive uint should use int arithmetic
		result, err := DoArithmetic(-5, uint(3), '+')
		require.NoError(t, err)
		assert.Equal(t, int64(-2), result)

		// Negative int - positive uint should use int arithmetic
		result, err = DoArithmetic(-5, uint(3), '-')
		require.NoError(t, err)
		assert.Equal(t, int64(-8), result)

		// Negative int * positive uint should use int arithmetic
		result, err = DoArithmetic(-5, uint(3), '*')
		require.NoError(t, err)
		assert.Equal(t, int64(-15), result)

		// Negative int / positive uint should use int arithmetic
		result, err = DoArithmetic(-6, uint(3), '/')
		require.NoError(t, err)
		assert.Equal(t, int64(-2), result)
	})

	t.Run("positive int with negative int cast to uint", func(t *testing.T) {
		// When second param is negative int, should use int arithmetic
		result, err := DoArithmetic(uint(5), -3, '+')
		require.NoError(t, err)
		assert.Equal(t, int64(2), result)
	})

	t.Run("max values", func(t *testing.T) {
		// Test with maximum int64 values
		maxInt := int64(math.MaxInt64)
		result, err := DoArithmetic(maxInt, int64(0), '+')
		require.NoError(t, err)
		assert.Equal(t, maxInt, result)

		// Test with maximum uint64 values
		maxUint := uint64(math.MaxUint64)
		result, err = DoArithmetic(maxUint, uint64(0), '+')
		require.NoError(t, err)
		assert.Equal(t, maxUint, result)
	})
}

func TestDoArithmetic_TypeSpecificBehavior(t *testing.T) {
	t.Run("integer type promotion", func(t *testing.T) {
		// All integer types should be promoted to int64
		tests := []struct {
			name string
			a    any
			b    any
		}{
			{"int8", int8(5), int8(3)},
			{"int16", int16(5), int16(3)},
			{"int32", int32(5), int32(3)},
			{"int64", int64(5), int64(3)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := DoArithmetic(tt.a, tt.b, '+')
				require.NoError(t, err)
				assert.IsType(t, int64(0), result)
				assert.Equal(t, int64(8), result)
			})
		}
	})

	t.Run("uint type promotion", func(t *testing.T) {
		// All uint types should be promoted to uint64
		tests := []struct {
			name string
			a    any
			b    any
		}{
			{"uint8", uint8(5), uint8(3)},
			{"uint16", uint16(5), uint16(3)},
			{"uint32", uint32(5), uint32(3)},
			{"uint64", uint64(5), uint64(3)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := DoArithmetic(tt.a, tt.b, '+')
				require.NoError(t, err)
				assert.IsType(t, uint64(0), result)
				assert.Equal(t, uint64(8), result)
			})
		}
	})

	t.Run("float type promotion", func(t *testing.T) {
		// Float32 and float64 should both result in float64
		result, err := DoArithmetic(float32(5.5), float64(3.2), '+')
		require.NoError(t, err)
		assert.IsType(t, float64(0), result)
		assert.InDelta(t, 8.7, result, 0.0001)
	})
}

// Benchmark tests
func BenchmarkDoArithmetic(b *testing.B) {
	benchmarks := []struct {
		name string
		a    any
		b    any
		op   rune
	}{
		{"int_add", 5, 3, '+'},
		{"int_mul", 5, 3, '*'},
		{"float_add", 5.5, 3.2, '+'},
		{"float_mul", 5.5, 3.2, '*'},
		{"uint_add", uint(5), uint(3), '+'},
		{"uint_mul", uint(5), uint(3), '*'},
		{"string_concat", "hello", "world", '+'},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = DoArithmetic(bm.a, bm.b, bm.op)
			}
		})
	}
}
