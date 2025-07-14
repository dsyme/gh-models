package pex

import (
	"testing"
)

func TestContainsJSONRule(t *testing.T) {
	handler := &pexCommandHandler{}
	
	tests := []struct {
		rule     string
		expected bool
	}{
		{"output should be JSON", true},
		{"return a JSON object", true},
		{"use JavaScript notation", true},
		{"respond with text", false},
		{"format as XML", false},
	}
	
	for _, test := range tests {
		result := handler.containsJSONRule(test.rule)
		if result != test.expected {
			t.Errorf("containsJSONRule(%q) = %v, expected %v", test.rule, result, test.expected)
		}
	}
}

func TestContainsLengthRule(t *testing.T) {
	handler := &pexCommandHandler{}
	
	tests := []struct {
		rule     string
		expected bool
	}{
		{"keep it brief", true},
		{"response should be short", true},
		{"limit to 100 characters", true},
		{"be concise", true},
		{"output should be JSON", false},
		{"format as XML", false},
	}
	
	for _, test := range tests {
		result := handler.containsLengthRule(test.rule)
		if result != test.expected {
			t.Errorf("containsLengthRule(%q) = %v, expected %v", test.rule, result, test.expected)
		}
	}
}