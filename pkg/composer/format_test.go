package composer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPronounceInt(t *testing.T) {
	testCases := []struct {
		arg    int
		expect string
	}{
		{arg: -1, expect: "minus one"},
		{arg: 0, expect: "zero"},
		{arg: 1, expect: "one"},
		{arg: 2, expect: "two"},
		{arg: 3, expect: "tree"},
		{arg: 4, expect: "fohwer"},
		{arg: 5, expect: "fife"},
		{arg: 6, expect: "six"},
		{arg: 7, expect: "seven"},
		{arg: 8, expect: "ait"},
		{arg: 9, expect: "niner"},
		{arg: 10, expect: "one zero"},
		{arg: 11, expect: "one one"},
		{arg: 12, expect: "one two"},
		{arg: 13, expect: "one tree"},
		{arg: 14, expect: "one fohwer"},
		{arg: 15, expect: "one fife"},
		{arg: 16, expect: "one six"},
		{arg: 17, expect: "one seven"},
		{arg: 18, expect: "one ait"},
		{arg: 19, expect: "one niner"},
		{arg: 20, expect: "two zero"},
		{arg: 308, expect: "tree zero ait"},
		{arg: 1688, expect: "one six ait ait"},
		{arg: 2992, expect: "two niner niner two"},
	}
	for _, test := range testCases {
		t.Run(strconv.Itoa(test.arg), func(t *testing.T) {
			actual := PronounceInt(test.arg)
			require.Equal(t, test.expect, actual, fmt.Sprintf("got %v, expected %v", actual, test.expect))
		})
	}
}

func TestPronounceDecimal(t *testing.T) {
	testCases := []struct {
		arg       float64
		precision int
		separator string
		expect    string
	}{
		{arg: 136.0, precision: 0, separator: "", expect: "one tree six"},
		{arg: 136.0, precision: 1, separator: "", expect: "one tree six point zero"},
		{arg: 136.0, precision: 1, separator: "decimal", expect: "one tree six decimal zero"},
		{arg: 249.500, precision: 1, separator: "", expect: "two fohwer niner point fife"},
		{arg: 249.500, precision: 2, separator: "", expect: "two fohwer niner point fife zero"},
	}
	for _, test := range testCases {
		t.Run(fmt.Sprintf("%v %v %v", test.arg, test.precision, test.separator), func(t *testing.T) {
			actual := PronounceDecimal(test.arg, test.precision, test.separator)
			require.Equal(t, test.expect, actual, fmt.Sprintf("got %v, expected %v", actual, test.expect))
		})
	}
}
