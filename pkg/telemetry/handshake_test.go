package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordHash(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		password string
		expected string
	}{
		{
			password: "",
			expected: "0",
		},
		{
			password: "local",
			expected: "16d6497244e5930b",
		},
		{
			password: "zerosugar",
			expected: "b8cd9b02e5dcacbd",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.password, func(t *testing.T) {
			t.Parallel()
			actual := hashPassword(tc.password)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
