package composer

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func PronounceBearing(d int) (s string) {
	for d < 0 {
		d += 360
	}
	d = d % 360
	s = PronounceInt(d)
	if d < 100 {
		s = "zero " + s
	}
	return
}

// PronounceInt pronounces the given integer as a sequence of digits.
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
		return "tree"
	case 4:
		return "fohwer"
	case 5:
		return "fife"
	case 6:
		return "six"
	case 7:
		return "seven"
	case 8:
		return "ait"
	case 9:
		return "niner"
	}

	panic(fmt.Sprintf("unexpected digit: %d", d))
}

var defaultDecimalSeparator = "point"

// PronounceFractional pronounces the given float as a sequence of digits
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
			panic(fmt.Sprintf("unexpected fractional part: %s", fractionalPartStr))
		}

		return fmt.Sprintf("%s %s %s", PronounceInt(integerPart), separator, PronounceInt(fractionalPart))
	}
}

// PronounceNumbers pronounces the digits in the given string as a sequence of digits.
// Non-digit characters are ignored.
func PronounceNumbers(s string) string {
	var builder strings.Builder
	for _, char := range s {
		if unicode.IsDigit(char) {
			i, err := strconv.Atoi(string(char))
			if err != nil {
				continue
			}
			builder.WriteString(fmt.Sprintf("%s ", PronounceInt(i)))
		}
	}
	return builder.String()
}
