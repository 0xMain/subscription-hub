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
