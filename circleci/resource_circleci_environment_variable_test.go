package circleci

import (
	"testing"
)

func TestCensorValue(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"1", "xxxx"},
		{"22", "xxxx2"},
		{"333", "xxxx3"},
		{"4444", "xxxx44"},
		{"55555", "xxxx55"},
		{"666666", "xxxx666"},
		{"7777777", "xxxx777"},
		{"88888888", "xxxx8888"},
	}

	for _, tt := range testCases {
		actual := censorValue(tt.input)
		if actual != tt.expected {
			t.Errorf("%s but expected %s", actual, tt.expected)
		}
	}
}
