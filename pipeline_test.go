package pipeline

import "testing"

func TestParseEnvironment(t *testing.T) {
	var tests = []struct {
		input string
		expected bool
	}{
		{"", false},
		{"FAKE", false},
		{"DEBUG", true},
		{"INTEGRATION", true},
		{"PRODUCTION", true},
		{"debug", true},
		{"integration", true},
		{"production", true},
	}

	for _, test := range tests {
		if parseEnvironment(test.input) != test.expected {
			t.Errorf("%s is not a valid environment", test.input)
		}
	}
}

func TestParseLogOutput(t *testing.T) {
	var tests = []struct {
		input string
		expected bool
	}{
		{"stdout", true},
		{"syslog", true},
		{"STDOUT", true},
		{"SYSLOG", true},
		{"", false},
		{"FAKE", false},
	}
	for _, test := range tests {
		if parseLogOutput(test.input) != test.expected {
			t.Errorf("%s is not a valid environment", test.input)
		}
	}
}
