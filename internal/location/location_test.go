package location_test

import (
	"testing"

	"github.com/roma-glushko/cargo/internal/location"
	"github.com/stretchr/testify/require"
)

func TestNewUNLocode_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected location.UNLocode
	}{
		{"SESTO", "SESTO"},
		{"USNYC", "USNYC"},
		{"CNHKG", "CNHKG"},
		{"AA2BB", "AA2BB"},
	}

	for _, tt := range tests {
		code, err := location.NewUNLocode(tt.input)
		require.NoError(t, err)
		require.Equal(t, tt.expected, code)
	}
}

func TestNewUNLocode_Uppercase(t *testing.T) {
	code, err := location.NewUNLocode("sesto")
	require.NoError(t, err)
	require.Equal(t, location.UNLocode("SESTO"), code)
}

func TestNewUNLocode_Invalid(t *testing.T) {
	tests := []string{
		"",
		"AB",
		"ABCDEF",
		"12345",
		"AB1CD",
		"A-BCD",
	}

	for _, input := range tests {
		_, err := location.NewUNLocode(input)
		require.Error(t, err, "expected error for input %q", input)
	}
}
