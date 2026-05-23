package common

import (
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var numberPrinter = message.NewPrinter(language.Mongolian)

func FormatAmount(f float32) string {
	r := numberPrinter.Sprintf("%.2f", f)
	r = strings.ReplaceAll(r, ",", " ")
	r = strings.ReplaceAll(r, ".", ",")
	return r
}

func CleanString(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.TrimSpace(s)
	return s
}
