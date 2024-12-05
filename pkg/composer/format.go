package composer

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/dharmab/skyeye/pkg/bearings"
)

// pronounceBearing composes a text representation of a bearing.
func pronounceBearing(bearing bearings.Bearing) (s string) {
	θ := int(bearing.RoundedDegrees())
	s = pronounceInt(θ)
	if θ < 10 {
		s = "0 " + s
	}
	if θ < 100 {
		s = "0 " + s
	}
	return
}

// pronounceInt composes a text representation of a sequence of digits.
func pronounceInt(d int) string {
	if d < 0 {
		return "minus " + pronounceInt(-d)
	}

	if d >= 10 {
		return pronounceInt(d/10) + " " + pronounceInt(d%10)
	}

	return strconv.Itoa(d)
}

var defaultDecimalSeparator = "point"

// pronounceDecimal composes a text representation of the given float as a sequence of digits.
func pronounceDecimal(f float64, precision int, separator string) string {
	if separator == "" {
		separator = defaultDecimalSeparator
	}
	integerPart := int(f)

	fractionalPartFloat := f - float64(integerPart)
	_, fractionalPartStr, _ := strings.Cut(fmt.Sprintf("%.*f", precision, fractionalPartFloat), ".")
	if fractionalPartStr == "" {
		fractionalPartStr = strings.Repeat("0", precision)
	}
	if fractionalPartStr == "" {
		return pronounceInt(integerPart)
	}
	fractionalPart, err := strconv.Atoi(fractionalPartStr)
	if err != nil {
		panic("unexpected fractional part: " + fractionalPartStr)
	}

	return fmt.Sprintf("%s %s %s", pronounceInt(integerPart), separator, pronounceInt(fractionalPart))
}

// pronounceNumbers composes a text representation of the digits in the given string as a sequence of digits.
// Non-digit characters are ignored.
func pronounceNumbers(s string) string {
	var builder strings.Builder
	for _, char := range s {
		if unicode.IsDigit(char) {
			i, err := strconv.Atoi(string(char))
			if err != nil {
				continue
			}
			_, _ = builder.WriteString(pronounceInt(i) + " ")
		}
	}
	return builder.String()
}
