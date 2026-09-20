package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var nonDigit = regexp.MustCompile(`\D+`)

// NormalizePhone converts AR phone numbers to international digits without +.
func NormalizePhone(raw string) string {
	if raw == "" {
		return ""
	}
	digits := nonDigit.ReplaceAllString(raw, "")
	if digits == "" {
		return ""
	}
	if strings.HasPrefix(digits, "00") {
		digits = digits[2:]
	}
	if !strings.HasPrefix(digits, "54") {
		if strings.HasPrefix(digits, "0") {
			digits = "54" + digits[1:]
		} else if len(digits) <= 10 {
			digits = "54" + digits
		}
	}
	if strings.HasPrefix(digits, "54") && !strings.HasPrefix(digits, "549") && len(digits) >= 12 {
		digits = "549" + digits[2:]
	}
	return digits
}

func TelHref(raw string) *string {
	digits := nonDigit.ReplaceAllString(raw, "")
	if digits == "" {
		return nil
	}
	s := "tel:" + digits
	return &s
}

func FillTemplate(body string, vars map[string]string) string {
	re := regexp.MustCompile(`\{(\w+)\}`)
	return re.ReplaceAllStringFunc(body, func(m string) string {
		key := m[1 : len(m)-1]
		if v, ok := vars[key]; ok {
			return v
		}
		return ""
	})
}

func OrderWhatsAppVars(order map[string]any) map[string]string {
	str := func(key string) string {
		if v, ok := order[key]; ok && v != nil {
			return fmt.Sprint(v)
		}
		return ""
	}
	addr := strings.TrimSpace(strings.Join(
		filterNonEmpty(str("client_address"), str("locality")),
		", ",
	))
	return map[string]string{
		"cliente":   str("client_name"),
		"trabajo":   str("title"),
		"estado":    str("status_name"),
		"fecha":     str("scheduled_date"),
		"hora":      str("scheduled_time"),
		"tecnico":   str("technician_name"),
		"direccion": addr,
		"telefono":  str("client_phone"),
		"orden":     str("id"),
	}
}

func filterNonEmpty(parts ...string) []string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" && p != "<nil>" {
			out = append(out, p)
		}
	}
	return out
}

func BuildWhatsAppURL(phone, text string) string {
	n := NormalizePhone(phone)
	if n == "" {
		return ""
	}
	return "https://wa.me/" + n + "?text=" + url.QueryEscape(text)
}
