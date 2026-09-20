package utils

import (
	"fmt"
	"strconv"
	"time"
)

func LocalDate(value string) string {
	if value != "" {
		if t, err := time.Parse("2006-01-02", value); err == nil {
			return t.Format("2006-01-02")
		}
		if t, err := time.Parse(time.RFC3339, value); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return time.Now().Format("2006-01-02")
}

func Atoi64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func NormalizeSQLValue(v any) any {
	switch t := v.(type) {
	case []byte:
		return string(t)
	case time.Time:
		return t.UTC().Format("2006-01-02 15:04:05")
	default:
		return v
	}
}

func RowsToMaps(rows interface {
	Next() bool
	Scan(dest ...any) error
	Columns() ([]string, error)
	Err() error
}) ([]map[string]any, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var out []map[string]any
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		m := make(map[string]any, len(cols))
		for i, col := range cols {
			m[col] = NormalizeSQLValue(vals[i])
		}
		out = append(out, m)
	}
	if out == nil {
		out = []map[string]any{}
	}
	return out, rows.Err()
}

func AsInt64(v any) (int64, bool) {
	switch t := v.(type) {
	case int64:
		return t, true
	case int:
		return int64(t), true
	case float64:
		return int64(t), true
	case string:
		n, err := strconv.ParseInt(t, 10, 64)
		return n, err == nil
	case []byte:
		n, err := strconv.ParseInt(string(t), 10, 64)
		return n, err == nil
	case *int64:
		if t == nil {
			return 0, false
		}
		return *t, true
	default:
		return 0, false
	}
}

func MustInt64(v any) int64 {
	n, _ := AsInt64(v)
	return n
}

func OptInt64(v any) *int64 {
	if v == nil {
		return nil
	}
	if t, ok := v.(*int64); ok {
		return t
	}
	n, ok := AsInt64(v)
	if !ok || n == 0 {
		return nil
	}
	return &n
}

func OptStr(v any) *string {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case *string:
		return t
	case string:
		if t == "" {
			return nil
		}
		return &t
	default:
		s := fmt.Sprint(v)
		if s == "" || s == "<nil>" {
			return nil
		}
		return &s
	}
}

func StrOr(v any, fallback string) string {
	if v == nil {
		return fallback
	}
	switch t := v.(type) {
	case string:
		if t == "" {
			return fallback
		}
		return t
	case *string:
		if t == nil || *t == "" {
			return fallback
		}
		return *t
	default:
		s := fmt.Sprint(v)
		if s == "" || s == "<nil>" {
			return fallback
		}
		return s
	}
}

func FloatOr(v any, fallback float64) float64 {
	if v == nil {
		return fallback
	}
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case string:
		f, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return fallback
		}
		return f
	default:
		return fallback
	}
}

func CoalesceTitle(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case *string:
		if t == nil {
			return ""
		}
		return *t
	default:
		return fmt.Sprint(v)
	}
}
