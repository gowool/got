package got

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"html"
	"html/template"
	"maps"
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/davecgh/go-spew/spew"
	"github.com/segmentio/go-camelcase"
	"github.com/segmentio/go-snakecase"
	"github.com/spf13/cast"
	"gopkg.in/yaml.v3"

	"github.com/gowool/got/internal"
)

var Funcs = template.FuncMap{
	"ternary": func(condition bool, trueValue, falseValue any) any {
		if condition {
			return trueValue
		}
		return falseValue
	},
	"empty": func(given any) bool {
		g := reflect.ValueOf(given)
		return !g.IsValid() || g.IsNil() || g.IsZero()
	},
	"escape": html.EscapeString,
	"deref": func(s any) any {
		v := reflect.ValueOf(s)
		if v.Kind() == reflect.Pointer {
			return v.Elem().Interface()
		}
		return s
	},
	"dump": spew.Sdump,

	// arithmetic functions
	"mul": func(inputs ...any) any {
		return doArithmetic(inputs, '*')
	},
	"div": func(inputs ...any) any {
		return doArithmetic(inputs, '/')
	},
	"add": func(inputs ...any) any {
		return doArithmetic(inputs, '+')
	},
	"sub": func(inputs ...any) any {
		return doArithmetic(inputs, '-')
	},

	// type conversion functions
	"to_js":             func(str string) template.JS { return template.JS(str) },
	"to_css":            func(str string) template.CSS { return template.CSS(str) },
	"to_html":           func(str string) template.HTML { return template.HTML(str) },
	"to_html_attr":      func(str string) template.HTMLAttr { return template.HTMLAttr(str) },
	"to_int":            cast.ToInt,
	"to_uint":           cast.ToUint,
	"to_int64":          cast.ToInt64,
	"to_uint64":         cast.ToUint64,
	"to_int32":          cast.ToInt32,
	"to_uint32":         cast.ToUint32,
	"to_int16":          cast.ToInt16,
	"to_uint16":         cast.ToUint16,
	"to_int8":           cast.ToInt8,
	"to_uint8":          cast.ToUint8,
	"to_float64":        cast.ToFloat64,
	"to_float32":        cast.ToFloat32,
	"to_bool":           cast.ToBool,
	"to_string":         cast.ToString,
	"to_time":           cast.ToTime,
	"to_duration":       cast.ToDuration,
	"to_slice":          cast.ToSlice,
	"to_string_slice":   cast.ToStringSlice,
	"to_float64_slice":  cast.ToFloat64Slice,
	"to_int64_slice":    cast.ToInt64Slice,
	"to_int_slice":      cast.ToIntSlice,
	"to_uint_slice":     cast.ToUintSlice,
	"to_bool_slice":     cast.ToBoolSlice,
	"to_duration_slice": cast.ToDurationSlice,

	// string functions
	"str_build": func(str ...string) string {
		var b strings.Builder
		for _, s := range str {
			b.WriteString(s)
		}
		return b.String()
	},
	"str_camelcase":   camelcase.Camelcase,
	"str_snakecase":   snakecase.Snakecase,
	"str_trim_space":  strings.TrimSpace,
	"str_trim_left":   strings.TrimLeft,
	"str_trim_right":  strings.TrimRight,
	"str_trim_prefix": strings.TrimPrefix,
	"str_trim_suffix": strings.TrimSuffix,
	"str_has_prefix":  strings.HasPrefix,
	"str_has_suffix":  strings.HasSuffix,
	"str_upper":       strings.ToUpper,
	"str_lower":       strings.ToLower,
	"str_title":       strings.ToTitle,
	"str_contains":    strings.Contains,
	"str_replace":     strings.ReplaceAll,
	"str_equal":       strings.EqualFold,
	"str_index":       strings.Index,
	"str_join":        strings.Join,
	"str_split":       strings.Split,
	"str_split_n":     strings.SplitN,
	"str_fields":      strings.Fields,
	"str_repeat":      strings.Repeat,
	"str_len":         func(s string) int { return utf8.RuneCountInString(s) },

	// encoding functions
	"json": func(v any) string {
		return encode(v, json.Marshal)
	},
	"xml": func(v any) string {
		return encode(v, xml.Marshal)
	},
	"yaml": func(v any) string {
		return encode(v, yaml.Marshal)
	},
	"json_pretty": func(v any) string {
		return pretty(v, json.MarshalIndent)
	},
	"xml_pretty": func(v any) string {
		return pretty(v, xml.MarshalIndent)
	},
	"yaml_pretty": func(v any) string {
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		if err := enc.Encode(v); err != nil {
			return ""
		}
		return template.JSEscapeString(internal.String(buf.Bytes()))
	},

	// slice functions
	"seq":      internal.Seq,
	"list":     func(v ...any) []any { return v },
	"first":    func(v []any) any { return v[0] },
	"last":     func(v []any) any { return v[len(v)-1] },
	"append":   func(v []any, e ...any) []any { return append(v, e...) },
	"prepend":  func(v []any, e ...any) []any { return append(e, v...) },
	"reverse":  func(v []any) []any { slices.Reverse(v); return v },
	"repeat":   func(v []any, count int) []any { return slices.Repeat(v, count) },
	"contains": func(v []any, i any) bool { return slices.Contains(v, i) },
	"index_of": func(v []any, i any) int { return slices.Index(v, i) },
	"concat":   func(sl ...[]any) []any { return slices.Concat(sl...) },

	// map functions
	"dict": func(v ...any) map[any]any {
		if len(v)%2 != 0 {
			v = append(v, "")
		}
		dict := make(map[any]any, len(v)/2)
		for i := 0; i < len(v); i += 2 {
			dict[v[i]] = v[i+1]
		}
		return dict
	},
	"keys":   func(m map[any]any) []any { return slices.Collect(maps.Keys(m)) },
	"values": func(m map[any]any) []any { return slices.Collect(maps.Values(m)) },
	"has":    func(m map[any]any, k any) bool { _, ok := m[k]; return ok },
	"get":    func(m map[any]any, k any) any { return m[k] },
	"set":    func(m map[any]any, k, v any) map[any]any { m[k] = v; return m },
	"unset":  func(m map[any]any, k any) map[any]any { delete(m, k); return m },

	// time functions
	"now":  time.Now,
	"date": FormatDate,
	"date_local": func(fmt string, date any) string {
		return FormatDate(fmt, date, "Local")
	},
	"date_utc": func(fmt string, date any) string {
		return FormatDate(fmt, date, "UTC")
	},
}

func FormatDate(fmt string, date any, location string) string {
	var t time.Time
	switch date := date.(type) {
	case time.Time:
		t = date
	case *time.Time:
		t = *date
	case int64:
		t = time.Unix(date, 0)
	case int:
		t = time.Unix(int64(date), 0)
	case int32:
		t = time.Unix(int64(date), 0)
	default:
		t = time.Now()
	}

	loc, err := time.LoadLocation(location)
	if err != nil {
		loc, _ = time.LoadLocation("UTC")
	}

	return t.In(loc).Format(fmt)
}

func encode(v any, fn func(v any) ([]byte, error)) string {
	raw, err := fn(v)
	if err != nil {
		return ""
	}
	return template.JSEscapeString(internal.String(raw))
}

func pretty(v any, fn func(v any, prefix, indent string) ([]byte, error)) string {
	raw, err := fn(v, "", "  ")
	if err != nil {
		return ""
	}
	return template.JSEscapeString(internal.String(raw))
}

func doArithmetic(inputs []any, operation rune) (value any) {
	if len(inputs) < 2 {
		if len(inputs) == 1 {
			return inputs[0]
		}
		return
	}

	var err error
	value = inputs[0]
	for i := 1; i < len(inputs); i++ {
		value, err = internal.DoArithmetic(value, inputs[i], operation)
		if err != nil {
			return
		}
	}
	return
}
