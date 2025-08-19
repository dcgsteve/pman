package commands

import (
	"testing"
)

func TestExpandCombinedFlags(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "single flag -r",
			input:    []string{"-r", "path"},
			expected: []string{"-r", "path"},
		},
		{
			name:     "single flag -f",
			input:    []string{"-f", "path"},
			expected: []string{"-f", "path"},
		},
		{
			name:     "combined flags -rf",
			input:    []string{"-rf", "path"},
			expected: []string{"-r", "-f", "path"},
		},
		{
			name:     "combined flags -fr",
			input:    []string{"-fr", "path"},
			expected: []string{"-f", "-r", "path"},
		},
		{
			name:     "multiple combined flags -rfg",
			input:    []string{"-rfg", "group1", "path"},
			expected: []string{"-r", "-f", "-g", "group1", "path"},
		},
		{
			name:     "mixed flags",
			input:    []string{"-rf", "-g", "group1", "path"},
			expected: []string{"-r", "-f", "-g", "group1", "path"},
		},
		{
			name:     "long flag --group",
			input:    []string{"--group", "group1", "path"},
			expected: []string{"--group", "group1", "path"},
		},
		{
			name:     "no flags",
			input:    []string{"path"},
			expected: []string{"path"},
		},
		{
			name:     "flags after path",
			input:    []string{"path", "-rf"},
			expected: []string{"-r", "-f", "path"}, // Flags should be reordered to come first
		},
		{
			name:     "group flag with value",
			input:    []string{"-g", "mygroup", "path"},
			expected: []string{"-g", "mygroup", "path"},
		},
		{
			name:     "combined flags with group",
			input:    []string{"-rfg", "mygroup", "path"},
			expected: []string{"-r", "-f", "-g", "mygroup", "path"},
		},
		{
			name:     "path then group flag",
			input:    []string{"path", "-g", "mygroup"},
			expected: []string{"-g", "mygroup", "path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandCombinedFlags(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expandCombinedFlags() = %v, want %v", result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("expandCombinedFlags() = %v, want %v", result, tt.expected)
					return
				}
			}
		})
	}
}