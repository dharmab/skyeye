package composer

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/dharmab/skyeye/pkg/bearings"
)

// PronounceBearing composes a text representation ofbearing.
func PronounceBearing(bearing bearings.Bearing) (s string) {
	θ := int(bearing.RoundedDegrees())
	s = PronounceInt(θ)
	if θ < 10 {
		s = "zero " + s
	}
	if θ < 100 {
		s = "zero " + s
	}
	return
}

// PronounceInt composes a text representation of a a sequence of digits, using aviation pronunciation.
func PronounceInt(d int) string {
	if d < 0 {
		return "minus " + PronounceInt(-d)
	}

	if d >= 10 {
		return PronounceInt(d/10) + " " + PronounceInt(d%10)
	}

	switch d {
	case 0:
		return "zero"
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	case 4:
		return "four"
	case 5:
		return "five"
	case 6:
		return "six"
	case 7:
		return "seven"
	case 8:
		return "eight"
	case 9:
		return "nine"
	}

	panic(fmt.Sprintf("unexpected digit: %d", d))
}

var defaultDecimalSeparator = "point"

// PronounceFractional composes a text representation of the given float as a sequence of digits.
func PronounceDecimal(f float64, precision int, separator string) string {
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
		return PronounceInt(integerPart)
	} else {
		fractionalPart, err := strconv.Atoi(fractionalPartStr)
		if err != nil {
			panic("unexpected fractional part: " + fractionalPartStr)
		}

		return fmt.Sprintf("%s %s %s", PronounceInt(integerPart), separator, PronounceInt(fractionalPart))
	}
}

// PronounceNumbers composes a text representation of the digits in the given string as a sequence of digits.
// Non-digit characters are ignored.
func PronounceNumbers(s string) string {
	var builder strings.Builder
	for _, char := range s {
		if unicode.IsDigit(char) {
			i, err := strconv.Atoi(string(char))
			if err != nil {
				continue
			}
			builder.WriteString(PronounceInt(i) + " ")
		}
	}
	return builder.String()
}
