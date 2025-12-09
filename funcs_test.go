package got

import (
	"html/template"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFuncs_Ternary(t *testing.T) {
	tests := []struct {
		name       string
		condition  bool
		trueValue  any
		falseValue any
		expected   any
	}{
		{"condition true", true, "yes", "no", "yes"},
		{"condition false", false, "yes", "no", "no"},
		{"condition true with numbers", true, 1, 0, 1},
		{"condition false with numbers", false, 1, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := Funcs["ternary"].(func(bool, any, any) any)
			result := fn(tt.condition, tt.trueValue, tt.falseValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFuncs_Empty(t *testing.T) {
	tests := []struct {
		name     string
		given    any
		expected bool
	}{
		{"nil value", nil, true},
		{"zero int", 0, true},
		{"zero string", "", true},
		{"zero bool", false, true},
		{"empty slice", []int{}, true},
		{"empty map", map[string]int{}, true},
		{"non-zero int", 42, false},
		{"non-empty string", "hello", false},
		{"true bool", true, false},
		{"non-empty slice", []int{1, 2, 3}, false},
		{"non-empty map", map[string]int{"a": 1}, false},
		{"nil pointer", (*int)(nil), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := Funcs["empty"].(func(any) bool)
			result := fn(tt.given)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFuncs_Escape(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain text", "hello", "hello"},
		{"HTML tags", "<div>content</div>", "&lt;div&gt;content&lt;/div&gt;"},
		{"ampersand", "me & you", "me &amp; you"},
		{"quotes", "say \"hello\"", "say &#34;hello&#34;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := Funcs["escape"].(func(string) string)
			result := fn(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFuncs_Deref(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{"pointer to int", func() *int { i := 42; return &i }(), 42},
		{"pointer to string", func() *string { s := "hello"; return &s }(), "hello"},
		{"nil pointer", (*int)(nil), (*int)(nil)},
		{"non-pointer value", "string", "string"},
		{"number", 42, 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := Funcs["deref"].(func(any) any)
			result := fn(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFuncs_Dump(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		contains string
	}{
		{"simple string", "hello", "hello"},
		{"number", 42, "42"},
		{"slice", []int{1, 2, 3}, "len=3"},
		{"map", map[string]int{"a": 1}, "len=1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := Funcs["dump"].(func(...any) string)
			result := fn(tt.input)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestFuncs_Arithmetic(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		inputs   []any
		expected any
	}{
		{"add two ints", "add", []any{2, 3}, int64(5)},
		{"add multiple ints", "add", []any{1, 2, 3, 4}, int64(10)},
		{"add float and int", "add", []any{2.5, 3}, float64(5.5)},
		{"sub two ints", "sub", []any{5, 3}, int64(2)},
		{"sub multiple ints", "sub", []any{10, 3, 2}, int64(5)},
		{"mul two ints", "mul", []any{3, 4}, int64(12)},
		{"mul multiple ints", "mul", []any{2, 3, 4}, int64(24)},
		{"div two ints", "div", []any{8, 2}, int64(4)},
		{"div float", "div", []any{7.0, 2.0}, float64(3.5)},
		{"single value", "add", []any{42}, 42},
		{"no values", "add", []any{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := Funcs[tt.funcName]
			require.NotNil(t, fn)
			result := fn.(func(...any) any)(tt.inputs...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFuncs_Arithmetic_StringConcat(t *testing.T) {
	fn := Funcs["add"].(func(...any) any)
	result := fn("hello", " ", "world")
	assert.Equal(t, "hello world", result)
}

func TestFuncs_Arithmetic_DivisionByZero(t *testing.T) {
	fn := Funcs["div"].(func(...any) any)
	result := fn(10, 0)
	assert.Nil(t, result)
}

func TestFuncs_TypeConversions(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		input    any
		expected any
	}{
		{"to_int", "to_int", "42", 42},
		{"to_uint", "to_uint", "42", uint(42)},
		{"to_int64", "to_int64", "42", int64(42)},
		{"to_uint64", "to_uint64", "42", uint64(42)},
		{"to_float64", "to_float64", "3.14", float64(3.14)},
		{"to_float32", "to_float32", "3.14", float32(3.14)},
		{"to_bool", "to_bool", "true", true},
		{"to_string", "to_string", 42, "42"},
		{"to_js", "to_js", "<script>", template.JS("<script>")},
		{"to_css", "to_css", "color:red;", template.CSS("color:red;")},
		{"to_html", "to_html", "<div>", template.HTML("<div>")},
		{"to_html_attr", "to_html_attr", "class=\"test\"", template.HTMLAttr("class=\"test\"")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := Funcs[tt.funcName]
			require.NotNil(t, fn)

			switch tt.funcName {
			case "to_int":
				result := fn.(func(any) int)(tt.input)
				assert.Equal(t, tt.expected, result)
			case "to_uint":
				result := fn.(func(any) uint)(tt.input)
				assert.Equal(t, tt.expected, result)
			case "to_int64":
				result := fn.(func(any) int64)(tt.input)
				assert.Equal(t, tt.expected, result)
			case "to_uint64":
				result := fn.(func(any) uint64)(tt.input)
				assert.Equal(t, tt.expected, result)
			case "to_float64":
				result := fn.(func(any) float64)(tt.input)
				assert.Equal(t, tt.expected, result)
			case "to_float32":
				result := fn.(func(any) float32)(tt.input)
				assert.Equal(t, tt.expected, result)
			case "to_bool":
				result := fn.(func(any) bool)(tt.input)
				assert.Equal(t, tt.expected, result)
			case "to_string":
				result := fn.(func(any) string)(tt.input)
				assert.Equal(t, tt.expected, result)
			case "to_js":
				result := fn.(func(string) template.JS)(tt.input.(string))
				assert.Equal(t, tt.expected, result)
			case "to_css":
				result := fn.(func(string) template.CSS)(tt.input.(string))
				assert.Equal(t, tt.expected, result)
			case "to_html":
				result := fn.(func(string) template.HTML)(tt.input.(string))
				assert.Equal(t, tt.expected, result)
			case "to_html_attr":
				result := fn.(func(string) template.HTMLAttr)(tt.input.(string))
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFuncs_StringOperations(t *testing.T) {
	t.Run("str_build", func(t *testing.T) {
		fn := Funcs["str_build"].(func(...string) string)
		result := fn("hello", " ", "world")
		assert.Equal(t, "hello world", result)
	})

	t.Run("str_camelcase", func(t *testing.T) {
		fn := Funcs["str_camelcase"].(func(string) string)
		result := fn("hello_world")
		assert.Equal(t, "helloWorld", result)
	})

	t.Run("str_snakecase", func(t *testing.T) {
		fn := Funcs["str_snakecase"].(func(string) string)
		result := fn("HelloWorld")
		assert.Equal(t, "hello_world", result)
	})

	t.Run("str_trim_space", func(t *testing.T) {
		fn := Funcs["str_trim_space"].(func(string) string)
		result := fn("  hello  ")
		assert.Equal(t, "hello", result)
	})

	t.Run("str_trim_left", func(t *testing.T) {
		fn := Funcs["str_trim_left"].(func(string, string) string)
		result := fn("  hello", " ")
		assert.Equal(t, "hello", result)
	})

	t.Run("str_trim_right", func(t *testing.T) {
		fn := Funcs["str_trim_right"].(func(string, string) string)
		result := fn("hello  ", " ")
		assert.Equal(t, "hello", result)
	})

	t.Run("str_trim_prefix", func(t *testing.T) {
		fn := Funcs["str_trim_prefix"].(func(string, string) string)
		result := fn("hello_world", "hello_")
		assert.Equal(t, "world", result)
	})

	t.Run("str_trim_suffix", func(t *testing.T) {
		fn := Funcs["str_trim_suffix"].(func(string, string) string)
		result := fn("hello_world", "_world")
		assert.Equal(t, "hello", result)
	})

	t.Run("str_has_prefix", func(t *testing.T) {
		fn := Funcs["str_has_prefix"].(func(string, string) bool)
		assert.True(t, fn("hello_world", "hello"))
		assert.False(t, fn("hello_world", "world"))
	})

	t.Run("str_has_suffix", func(t *testing.T) {
		fn := Funcs["str_has_suffix"].(func(string, string) bool)
		assert.True(t, fn("hello_world", "_world"))
		assert.False(t, fn("hello_world", "_hello"))
	})

	t.Run("str_upper", func(t *testing.T) {
		fn := Funcs["str_upper"].(func(string) string)
		result := fn("hello")
		assert.Equal(t, "HELLO", result)
	})

	t.Run("str_lower", func(t *testing.T) {
		fn := Funcs["str_lower"].(func(string) string)
		result := fn("HELLO")
		assert.Equal(t, "hello", result)
	})

	t.Run("str_title", func(t *testing.T) {
		fn := Funcs["str_title"].(func(string) string)
		result := fn("hello world")
		assert.Equal(t, "HELLO WORLD", result)
	})

	t.Run("str_contains", func(t *testing.T) {
		fn := Funcs["str_contains"].(func(string, string) bool)
		assert.True(t, fn("hello world", "world"))
		assert.False(t, fn("hello world", "moon"))
	})

	t.Run("str_replace", func(t *testing.T) {
		fn := Funcs["str_replace"].(func(string, string, string) string)
		result := fn("hello world world", "world", "universe")
		assert.Equal(t, "hello universe universe", result)
	})

	t.Run("str_equal", func(t *testing.T) {
		fn := Funcs["str_equal"].(func(string, string) bool)
		assert.True(t, fn("HELLO", "hello"))
		assert.False(t, fn("hello", "world"))
	})

	t.Run("str_index", func(t *testing.T) {
		fn := Funcs["str_index"].(func(string, string) int)
		assert.Equal(t, 6, fn("hello world", "world"))
		assert.Equal(t, -1, fn("hello", "world"))
	})

	t.Run("str_join", func(t *testing.T) {
		fn := Funcs["str_join"].(func([]string, string) string)
		result := fn([]string{"a", "b", "c"}, ",")
		assert.Equal(t, "a,b,c", result)
	})

	t.Run("str_split", func(t *testing.T) {
		fn := Funcs["str_split"].(func(string, string) []string)
		result := fn("a,b,c", ",")
		assert.Equal(t, []string{"a", "b", "c"}, result)
	})

	t.Run("str_split_n", func(t *testing.T) {
		fn := Funcs["str_split_n"].(func(string, string, int) []string)
		result := fn("a,b,c,d", ",", 2)
		assert.Equal(t, []string{"a", "b,c,d"}, result)
	})

	t.Run("str_fields", func(t *testing.T) {
		fn := Funcs["str_fields"].(func(string) []string)
		result := fn("  hello   world  ")
		assert.Equal(t, []string{"hello", "world"}, result)
	})

	t.Run("str_repeat", func(t *testing.T) {
		fn := Funcs["str_repeat"].(func(string, int) string)
		result := fn("ab", 3)
		assert.Equal(t, "ababab", result)
	})

	t.Run("str_len", func(t *testing.T) {
		fn := Funcs["str_len"].(func(string) int)
		assert.Equal(t, 5, fn("hello"))
		assert.Equal(t, 8, fn("hello 世界"))
	})
}

func TestFuncs_SliceOperations(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		inputs   []any
		expected any
	}{
		{"seq single positive", "seq", []any{3}, []int{1, 2, 3}},
		{"seq single negative", "seq", []any{-3}, []int{-1, -2, -3}},
		{"seq single zero", "seq", []any{0}, []int(nil)},
		{"seq range", "seq", []any{1, 4}, []int{1, 2, 3, 4}},
		{"seq range negative", "seq", []any{4, 1}, []int{4, 3, 2, 1}},
		{"seq with step", "seq", []any{1, 2, 5}, []int{1, 3, 5}},
		{"seq invalid args", "seq", []any{1, 2, 3, 4}, []int(nil)},
		{"list", "list", []any{1, "hello", true}, []any{1, "hello", true}},
		{"first", "first", []any{[]any{1, 2, 3}}, 1},
		{"last", "last", []any{[]any{1, 2, 3}}, 3},
		{"append", "append", []any{[]any{1, 2}, 3, 4}, []any{1, 2, 3, 4}},
		{"prepend", "prepend", []any{[]any{2, 3}, 1}, []any{1, 2, 3}},
		{"reverse", "reverse", []any{[]any{1, 2, 3}}, []any{3, 2, 1}},
		{"repeat", "repeat", []any{[]any{1, 2}, 2}, []any{1, 2, 1, 2}},
		{"contains true", "contains", []any{[]any{1, 2, 3}, 2}, true},
		{"contains false", "contains", []any{[]any{1, 2, 3}, 4}, false},
		{"index_of found", "index_of", []any{[]any{1, 2, 3}, 2}, 1},
		{"index_of not found", "index_of", []any{[]any{1, 2, 3}, 4}, -1},
		{"concat", "concat", []any{[]any{1, 2}, []any{3, 4}}, []any{1, 2, 3, 4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := Funcs[tt.funcName]
			require.NotNil(t, fn)

			switch v := fn.(type) {
			case func(...int) []int:
				if len(tt.inputs) == 1 {
					result := v(tt.inputs[0].(int))
					assert.Equal(t, tt.expected, result)
				} else if len(tt.inputs) == 2 {
					result := v(tt.inputs[0].(int), tt.inputs[1].(int))
					assert.Equal(t, tt.expected, result)
				} else if len(tt.inputs) == 3 {
					result := v(tt.inputs[0].(int), tt.inputs[1].(int), tt.inputs[2].(int))
					assert.Equal(t, tt.expected, result)
				}
			case func(...any) []any:
				result := v(tt.inputs...)
				assert.Equal(t, tt.expected, result)
			case func([]any) any:
				result := v(tt.inputs[0].([]any))
				assert.Equal(t, tt.expected, result)
			case func([]any, ...any) []any:
				slice := tt.inputs[0].([]any)
				extras := tt.inputs[1:]
				result := v(slice, extras...)
				assert.Equal(t, tt.expected, result)
			case func([]any, int) []any:
				result := v(tt.inputs[0].([]any), tt.inputs[1].(int))
				assert.Equal(t, tt.expected, result)
			case func([]any, any) bool:
				result := v(tt.inputs[0].([]any), tt.inputs[1])
				assert.Equal(t, tt.expected, result)
			case func([]any, any) int:
				result := v(tt.inputs[0].([]any), tt.inputs[1])
				assert.Equal(t, tt.expected, result)
			case func(...[]any) []any:
				slices := make([][]any, len(tt.inputs))
				for i, input := range tt.inputs {
					slices[i] = input.([]any)
				}
				result := v(slices...)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFuncs_MapOperations(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		inputs   []any
		expected any
	}{
		{"dict even pairs", "dict", []any{"a", 1, "b", 2}, map[any]any{"a": 1, "b": 2}},
		{"dict odd pairs", "dict", []any{"a", 1, "b"}, map[any]any{"a": 1, "b": ""}},
		{"dict empty", "dict", []any{}, map[any]any{}},
		{"keys", "keys", []any{map[any]any{"a": 1, "b": 2}}, []any{"a", "b"}},
		{"values", "values", []any{map[any]any{"a": 1, "b": 2}}, []any{1, 2}},
		{"has true", "has", []any{map[any]any{"a": 1}, "a"}, true},
		{"has false", "has", []any{map[any]any{"a": 1}, "b"}, false},
		{"get existing", "get", []any{map[any]any{"a": 1}, "a"}, 1},
		{"get missing", "get", []any{map[any]any{"a": 1}, "b"}, nil},
		{"set existing", "set", []any{map[any]any{"a": 1}, "a", 2}, map[any]any{"a": 2}},
		{"set new", "set", []any{map[any]any{"a": 1}, "b", 2}, map[any]any{"a": 1, "b": 2}},
		{"unset existing", "unset", []any{map[any]any{"a": 1, "b": 2}, "a"}, map[any]any{"b": 2}},
		{"unset missing", "unset", []any{map[any]any{"a": 1}, "b"}, map[any]any{"a": 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := Funcs[tt.funcName]
			require.NotNil(t, fn)

			switch v := fn.(type) {
			case func(...any) map[any]any:
				result := v(tt.inputs...)
				assert.Equal(t, tt.expected, result)
			case func(map[any]any) []any:
				result := v(tt.inputs[0].(map[any]any))
				expected := tt.expected.([]any)
				assert.Len(t, result, len(expected))
				switch tt.funcName {
				case "values":
					// For values function, check that all expected values are present (order not guaranteed)
					for _, val := range expected {
						assert.Contains(t, result, val)
					}
				case "keys":
					// For keys function, check that all expected keys are present (order not guaranteed)
					for _, key := range expected {
						assert.Contains(t, result, key)
					}
				default:
					assert.Equal(t, tt.expected, result)
				}
			case func(map[any]any, any) bool:
				result := v(tt.inputs[0].(map[any]any), tt.inputs[1])
				assert.Equal(t, tt.expected, result)
			case func(map[any]any, any) any:
				result := v(tt.inputs[0].(map[any]any), tt.inputs[1])
				assert.Equal(t, tt.expected, result)
			case func(map[any]any, any, any) map[any]any:
				result := v(tt.inputs[0].(map[any]any), tt.inputs[1], tt.inputs[2])
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFuncs_Encoding(t *testing.T) {
	testData := map[string]interface{}{
		"name":   "John",
		"age":    30,
		"active": true,
	}

	tests := []struct {
		name     string
		funcName string
		input    any
		contains string
	}{
		{"json", "json", testData, "\\\"name\\\":\\\"John\\\""},
		{"yaml", "yaml", testData, "name: John"},
		{"json_pretty", "json_pretty", testData, "\\u000A  \\\"name\\\": \\\"John\\\""},
		{"yaml_pretty", "yaml_pretty", testData, "name: John"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := Funcs[tt.funcName]
			require.NotNil(t, fn)
			result := fn.(func(any) string)(tt.input)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestFuncs_Encoding_Error(t *testing.T) {
	// Test with invalid data that should cause encoding errors
	invalidData := func() {} // functions can't be encoded

	tests := []struct {
		name     string
		funcName string
		input    any
		expected string
	}{
		{"json error", "json", invalidData, ""},
		{"yaml error", "yaml", invalidData, ""},
		{"json_pretty error", "json_pretty", invalidData, ""},
		{"yaml_pretty error", "yaml_pretty", invalidData, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := Funcs[tt.funcName]
			require.NotNil(t, fn)

			// Test that it doesn't panic and returns expected result
			assert.NotPanics(t, func() {
				result := fn.(func(any) string)(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		})
	}
}

func TestFuncs_Time(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		funcName string
		inputs   []any
		validate func(string)
	}{
		{"now", "now", nil, func(result string) {
			assert.NotEmpty(t, result)
		}},
		{"date with time.Time", "date", []any{"2006-01-02", now}, func(result string) {
			assert.Equal(t, now.Format("2006-01-02"), result)
		}},
		{"date with *time.Time", "date", []any{"2006-01-02", &now}, func(result string) {
			assert.Equal(t, now.Format("2006-01-02"), result)
		}},
		{"date with int64", "date", []any{"2006", int64(1640995200)}, func(result string) {
			// Just check it's a valid year format
			assert.Len(t, result, 4)
		}},
		{"date with int", "date", []any{"2006", 1640995200}, func(result string) {
			assert.Len(t, result, 4)
		}},
		{"date with int32", "date", []any{"2006", int32(1640995200)}, func(result string) {
			assert.Len(t, result, 4)
		}},
		{"date default", "date", []any{"2006"}, func(result string) {
			assert.Len(t, result, 4)
		}},
		{"date_local", "date_local", []any{"2006-01-02", now}, func(result string) {
			assert.NotEmpty(t, result)
		}},
		{"date_utc", "date_utc", []any{"2006-01-02", now}, func(result string) {
			assert.Equal(t, now.UTC().Format("2006-01-02"), result)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := Funcs[tt.funcName]
			require.NotNil(t, fn)

			switch v := fn.(type) {
			case func() time.Time:
				result := v()
				tt.validate(result.Format("2006-01-02"))
			case func(string, any) string:
				if len(tt.inputs) == 2 {
					result := v(tt.inputs[0].(string), tt.inputs[1])
					tt.validate(result)
				}
			case func(string, any, string) string:
				if len(tt.inputs) == 3 {
					result := v(tt.inputs[0].(string), tt.inputs[1], tt.inputs[2].(string))
					tt.validate(result)
				}
			case func(string) string:
				if len(tt.inputs) == 1 {
					result := v(tt.inputs[0].(string))
					tt.validate(result)
				}
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)

	tests := []struct {
		name     string
		fmt      string
		date     any
		location string
		expected string
	}{
		{"time.Time UTC", "2006-01-02", testTime, "UTC", "2023-12-25"},
		{"time.Time Local", "2006-01-02", testTime, "Local", testTime.Local().Format("2006-01-02")},
		{"*time.Time", "2006", &testTime, "UTC", "2023"},
		{"int64 timestamp", "2006", int64(1609459200), "UTC", "2021"},
		{"int timestamp", "2006", 1609459200, "UTC", "2021"},
		{"int32 timestamp", "2006", int32(1609459200), "UTC", "2021"},
		{"invalid type", "2006", "invalid", "UTC", time.Now().Format("2006")},
		{"invalid location", "2006", testTime, "InvalidLocation", testTime.UTC().Format("2006")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDate(tt.fmt, tt.date, tt.location)
			assert.Equal(t, tt.expected, result)
		})
	}
}
