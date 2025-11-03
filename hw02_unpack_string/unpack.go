package hw02unpackstring

import (
	"errors"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(in string) (string, error) {
	var builder strings.Builder
	runes := []rune(in)
	for i := 0; i < len(runes); i++ {
		if unicode.IsDigit(runes[i]) {
			return builder.String(), ErrInvalidString
		}
		if runes[i] == '\\' {
			if i+1 < len(runes) {
				if unicode.IsDigit(runes[i+1]) || runes[i+1] == '\\' {
					i++
				}
			} else {
				return builder.String(), ErrInvalidString
			}
		}
		if i+1 < len(runes) && unicode.IsDigit(runes[i+1]) {
			for j := 0; j < int(runes[i+1]-'0'); j++ {
				builder.WriteRune(runes[i])
			}
			i++
		} else {
			builder.WriteRune(runes[i])
		}
	}
	return builder.String(), nil
}
