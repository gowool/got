package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeq_SingleArgument(t *testing.T) {
	tests := []struct {
		name     string
		arg      int
		expected []int
	}{
		{
			name:     "positive single argument",
			arg:      3,
			expected: []int{1, 2, 3},
		},
		{
			name:     "single argument 1",
			arg:      1,
			expected: []int{1},
		},
		{
			name:     "negative single argument",
			arg:      -3,
			expected: []int{-1, -2, -3},
		},
		{
			name:     "negative single argument -1",
			arg:      -1,
			expected: []int{-1},
		},
		{
			name:     "zero argument returns nil",
			arg:      0,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Seq(tt.arg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSeq_TwoArguments(t *testing.T) {
	tests := []struct {
		name     string
		first    int
		last     int
		expected []int
	}{
		{
			name:     "ascending sequence",
			first:    1,
			last:     3,
			expected: []int{1, 2, 3},
		},
		{
			name:     "ascending sequence same values",
			first:    5,
			last:     5,
			expected: []int{5},
		},
		{
			name:     "descending sequence",
			first:    3,
			last:     1,
			expected: []int{3, 2, 1},
		},
		{
			name:     "descending sequence same values",
			first:    5,
			last:     5,
			expected: []int{5},
		},
		{
			name:     "positive to negative",
			first:    1,
			last:     -2,
			expected: []int{1, 0, -1, -2},
		},
		{
			name:     "negative to positive",
			first:    -2,
			last:     1,
			expected: []int{-2, -1, 0, 1},
		},
		{
			name:     "zero to positive",
			first:    0,
			last:     3,
			expected: []int{0, 1, 2, 3},
		},
		{
			name:     "positive to zero",
			first:    3,
			last:     0,
			expected: []int{3, 2, 1, 0},
		},
		{
			name:     "zero to negative",
			first:    0,
			last:     -3,
			expected: []int{0, -1, -2, -3},
		},
		{
			name:     "negative to zero",
			first:    -3,
			last:     0,
			expected: []int{-3, -2, -1, 0},
		},
		{
			name:     "large ascending sequence",
			first:    1,
			last:     10,
			expected: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Seq(tt.first, tt.last)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSeq_ThreeArguments(t *testing.T) {
	tests := []struct {
		name     string
		first    int
		inc      int
		last     int
		expected []int
	}{
		{
			name:     "positive increment",
			first:    1,
			inc:      2,
			last:     5,
			expected: []int{1, 3, 5},
		},
		{
			name:     "positive increment exact",
			first:    1,
			inc:      3,
			last:     7,
			expected: []int{1, 4, 7},
		},
		{
			name:     "positive increment over",
			first:    1,
			inc:      3,
			last:     6,
			expected: []int{1, 4},
		},
		{
			name:     "negative increment",
			first:    5,
			inc:      -2,
			last:     1,
			expected: []int{5, 3, 1},
		},
		{
			name:     "negative increment exact",
			first:    7,
			inc:      -3,
			last:     1,
			expected: []int{7, 4, 1},
		},
		{
			name:     "negative increment over",
			first:    7,
			inc:      -3,
			last:     2,
			expected: []int{7, 4},
		},
		{
			name:     "increment of 1",
			first:    1,
			inc:      1,
			last:     3,
			expected: []int{1, 2, 3},
		},
		{
			name:     "increment of -1",
			first:    3,
			inc:      -1,
			last:     1,
			expected: []int{3, 2, 1},
		},
		{
			name:     "zero start with positive inc",
			first:    0,
			inc:      2,
			last:     6,
			expected: []int{0, 2, 4, 6},
		},
		{
			name:     "zero start with negative inc",
			first:    0,
			inc:      -2,
			last:     -6,
			expected: []int{0, -2, -4, -6},
		},
		{
			name:     "negative to positive with inc",
			first:    -5,
			inc:      3,
			last:     7,
			expected: []int{-5, -2, 1, 4, 7},
		},
		{
			name:     "positive to negative with inc",
			first:    5,
			inc:      -3,
			last:     -7,
			expected: []int{5, 2, -1, -4, -7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Seq(tt.first, tt.inc, tt.last)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSeq_ErrorCases(t *testing.T) {
	tests := []struct {
		name string
		args []int
	}{
		{
			name: "no arguments",
			args: []int{},
		},
		{
			name: "too many arguments",
			args: []int{1, 2, 3, 4},
		},
		{
			name: "three arguments with zero increment",
			args: []int{1, 0, 5},
		},
		{
			name: "three arguments with positive first/last but negative increment",
			args: []int{1, -1, 5},
		},
		{
			name: "three arguments with negative first/last but positive increment",
			args: []int{5, 1, 1},
		},
		{
			name: "size exceeds limit - large negative last",
			args: []int{1, -200000},
		},
		{
			name: "size exceeds limit - large positive sequence",
			args: []int{1, 1, 3000},
		},
		{
			name: "size exceeds limit - large negative sequence",
			args: []int{-1, -1, -3000},
		},
		{
			name: "size zero or negative - first > last with positive inc",
			args: []int{5, 1, 1},
		},
		{
			name: "size zero or negative - first < last with negative inc",
			args: []int{1, -1, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Seq(tt.args...)
			assert.Nil(t, result)
		})
	}
}

func TestSeq_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		args     []int
		expected []int
	}{
		{
			name:     "maximum allowed size positive",
			args:     []int{1, 1, 2000},
			expected: makeRangeWithInc(1, 1, 2000),
		},
		{
			name:     "maximum allowed size negative",
			args:     []int{-1, -1, -2000},
			expected: makeRangeWithInc(-1, -1, -2000),
		},
		{
			name:     "boundary - just under limit",
			args:     []int{1, 1, 1999},
			expected: makeRangeWithInc(1, 1, 1999),
		},
		{
			name:     "boundary - just over limit returns nil",
			args:     []int{1, 1, 2001},
			expected: nil,
		},
		{
			name:     "boundary - very large negative last",
			args:     []int{1, -100001},
			expected: nil,
		},
		{
			name:     "boundary - just above large negative limit",
			args:     []int{1, -99999},
			expected: makeRange(1, -99999),
		},
		{
			name:     "single element with three args",
			args:     []int{5, 1, 5},
			expected: []int{5},
		},
		{
			name:     "single element with three args negative inc",
			args:     []int{5, -1, 5},
			expected: []int{5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Seq(tt.args...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create expected ranges for testing
func makeRange(first, last int) []int {
	inc := 1
	if last < first {
		inc = -1
	}
	size := ((last - first) / inc) + 1
	if size <= 0 || size > 2000 {
		return nil
	}

	seq := make([]int, size)
	val := first
	for i := 0; ; i++ {
		seq[i] = val
		val += inc
		if (inc < 0 && val < last) || (inc > 0 && val > last) {
			break
		}
	}
	return seq
}

// Helper function to create expected ranges for testing with increment
func makeRangeWithInc(first, inc, last int) []int {
	size := ((last - first) / inc) + 1
	if size <= 0 || size > 2000 {
		return nil
	}

	seq := make([]int, size)
	val := first
	for i := 0; ; i++ {
		seq[i] = val
		val += inc
		if (inc < 0 && val < last) || (inc > 0 && val > last) {
			break
		}
	}
	return seq
}

// Benchmark tests
func BenchmarkSeq_SingleArg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Seq(10)
	}
}

func BenchmarkSeq_TwoArgs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Seq(1, 100)
	}
}

func BenchmarkSeq_ThreeArgs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Seq(1, 2, 100)
	}
}

func BenchmarkSeq_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Seq(1, 1, 1000)
	}
}
