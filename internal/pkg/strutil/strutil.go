package strutil

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Capitalize(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if s == "" {
		return ""
	}

	return cases.Title(language.Und).String(strings.ToLower(s))
}

func CapitalizePtr(p *string) *string {
	if p == nil {
		return nil
	}
	res := Capitalize(*p)
	return &res
}

func NormalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func Plural(n int, one, few, many string) string {
	v := n % 100
	if v >= 11 && v <= 19 {
		return many
	}

	switch v % 10 {
	case 1:
		return one
	case 2, 3, 4:
		return few
	default:
		return many
	}
}
