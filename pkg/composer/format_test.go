package composer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPronounceInt(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		arg    int
		expect string
	}{
		{arg: -1, expect: "minus 1"},
		{arg: 0, expect: "0"},
		{arg: 1, expect: "1"},
		{arg: 2, expect: "2"},
		{arg: 3, expect: "3"},
		{arg: 4, expect: "4"},
		{arg: 5, expect: "5"},
		{arg: 6, expect: "6"},
		{arg: 7, expect: "7"},
		{arg: 8, expect: "8"},
		{arg: 9, expect: "9"},
		{arg: 10, expect: "1 0"},
		{arg: 11, expect: "1 1"},
		{arg: 12, expect: "1 2"},
		{arg: 13, expect: "1 3"},
		{arg: 14, expect: "1 4"},
		{arg: 15, expect: "1 5"},
		{arg: 16, expect: "1 6"},
		{arg: 17, expect: "1 7"},
		{arg: 18, expect: "1 8"},
		{arg: 19, expect: "1 9"},
		{arg: 20, expect: "2 0"},
		{arg: 308, expect: "3 0 8"},
		{arg: 1688, expect: "1 6 8 8"},
		{arg: 2992, expect: "2 9 9 2"},
	}
	for _, test := range testCases {
		t.Run(strconv.Itoa(test.arg), func(t *testing.T) {
			t.Parallel()
			actual := PronounceInt(test.arg)
			require.Equal(t, test.expect, actual, fmt.Sprintf("got %v, expected %v", actual, test.expect))
		})
	}
}

func TestPronounceDecimal(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		arg       float64
		precision int
		separator string
		expect    string
	}{
		{arg: 136.0, precision: 0, separator: "", expect: "1 3 6"},
		{arg: 136.0, precision: 1, separator: "", expect: "1 3 6 point 0"},
		{arg: 136.0, precision: 1, separator: "decimal", expect: "1 3 6 decimal 0"},
		{arg: 249.500, precision: 1, separator: "", expect: "2 4 9 point 5"},
		{arg: 249.500, precision: 2, separator: "", expect: "2 4 9 point 5 0"},
	}
	for _, test := range testCases {
		t.Run(fmt.Sprintf("%v %v %v", test.arg, test.precision, test.separator), func(t *testing.T) {
			t.Parallel()
			actual := PronounceDecimal(test.arg, test.precision, test.separator)
			require.Equal(t, test.expect, actual, fmt.Sprintf("got %v, expected %v", actual, test.expect))
		})
	}
}
