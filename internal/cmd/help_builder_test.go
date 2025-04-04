package cmd

import (
	"strings"
	"testing"
)

func TestCommandHelpBuilder_Build(t *testing.T) {
	tests := []struct {
		name         string
		builder      CommandHelpBuilder
		disableColor bool
		want         string
	}{
		{
			name:    "empty builder",
			builder: CommandHelpBuilder{},
			want:    "\n",
		},
		{
			name: "long description only",
			builder: CommandHelpBuilder{
				LongDesc: "This is a long description.",
			},
			want: "This is a long description.\n",
		},
		{
			name: "long description only - no color",
			builder: CommandHelpBuilder{
				LongDesc: "This is a long description.",
			},
			disableColor: true,
			want:         "This is a long description.\n",
		},
		{
			name: "required flags",
			builder: CommandHelpBuilder{
				RequiredFlags: map[string]string{
					"flag1": "Description for flag1",
					"flag2": "Description for flag2",
				},
			},
			want: "Required fields:\n  flag1: Description for flag1\n  flag2: Description for flag2\n",
		},
		{
			name: "optional flags",
			builder: CommandHelpBuilder{
				OptionalFlags: map[string]string{
					"flag1": "Description for flag1",
					"flag2": "Description for flag2",
				},
			},
			want: "Optional fields:\n  flag1: Description for flag1\n  flag2: Description for flag2\n",
		},
		{
			name: "important notes",
			builder: CommandHelpBuilder{
				ImportantNotes: []string{
					"Note 1",
					"Note 2",
				},
			},
			want: "Important notes:\n  - Note 1\n  - Note 2\n",
		},
		{
			name: "examples",
			builder: CommandHelpBuilder{
				Examples: []string{
					"example1",
					"example2",
				},
			},
			want: "Example usage:\n\n  example1\n  example2\n",
		},
		{
			name: "json examples",
			builder: CommandHelpBuilder{
				JSONExamples: []string{
					`{"key": "value"}`,
					`{"key2": "value2"}`,
				},
			},
			want: "JSON format example:\n{\"key\": \"value\"}\n{\"key2\": \"value2\"}\n",
		},
		{
			name: "all sections",
			builder: CommandHelpBuilder{
				LongDesc: "This is a long description.",
				RequiredFlags: map[string]string{
					"flag1": "Description for flag1",
				},
				OptionalFlags: map[string]string{
					"flag2": "Description for flag2",
				},
				ImportantNotes: []string{
					"Note 1",
				},
				Examples: []string{
					"example1",
				},
				JSONExamples: []string{
					`{"key": "value"}`,
				},
			},
			want: "This is a long description.\n\n" +
				"Required fields:\n" +
				"  flag1: Description for flag1\n\n" +
				"Optional fields:\n" +
				"  flag2: Description for flag2\n\n" +
				"Important notes:\n" +
				"  - Note 1\n\n" +
				"Example usage:\n\n" +
				"  example1\n\n" +
				"JSON format example:\n" +
				"{\"key\": \"value\"}\n",
		},
		{
			name: "all sections - no color",
			builder: CommandHelpBuilder{
				LongDesc: "This is a long description.",
				RequiredFlags: map[string]string{
					"flag1": "Description for flag1",
				},
				OptionalFlags: map[string]string{
					"flag2": "Description for flag2",
				},
				ImportantNotes: []string{
					"Note 1",
				},
				Examples: []string{
					"example1",
				},
				JSONExamples: []string{
					`{"key": "value"}`,
				},
			},
			disableColor: true,
			want: "This is a long description.\n\n" +
				"Required fields:\n" +
				"  flag1: Description for flag1\n\n" +
				"Optional fields:\n" +
				"  flag2: Description for flag2\n\n" +
				"Important notes:\n" +
				"  - Note 1\n\n" +
				"Example usage:\n\n" +
				"  example1\n\n" +
				"JSON format example:\n" +
				"{\"key\": \"value\"}\n",
		},
		{
			name: "no trailing newline",
			builder: CommandHelpBuilder{
				LongDesc: "This is a long description.",
			},
			want: "This is a long description.\n",
		},
		{
			name: "no trailing newline - no color",
			builder: CommandHelpBuilder{
				LongDesc: "This is a long description.",
			},
			disableColor: true,
			want:         "This is a long description.\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.builder.DisableColor = tt.disableColor
			got := tt.builder.Build()

			// Strip ANSI colors for comparison
			gotStripped := stripANSIColors(got)
			wantStripped := stripANSIColors(tt.want)

			if gotStripped != wantStripped {
				t.Errorf("CommandHelpBuilder.Build() = %q, want %q", gotStripped, wantStripped)
			}

			// Check for newline only if the output is not empty
			if got != "" && !strings.HasSuffix(got, "\n") {
				t.Errorf("CommandHelpBuilder.Build() result should end with a newline, but it doesn't")
			}
		})
	}
}

func TestCommandHelpBuilder_DisableColor(t *testing.T) {
	// Basic test that verifies color codes aren't present when DisableColor is true
	builder := CommandHelpBuilder{
		LongDesc:     "This is a long description.",
		DisableColor: true,
	}

	result := builder.Build()

	// No ANSI color codes should be present
	if strings.Contains(result, "\x1b[") {
		t.Errorf("Expected no color codes when DisableColor is true, but found some in: %q", result)
	}

	// Try with color enabled
	builder.DisableColor = false
	resultWithColor := builder.Build()

	// ANSI color codes should be present when color is enabled
	if !strings.Contains(resultWithColor, "\x1b[") {
		t.Errorf("Expected color codes when DisableColor is false, but found none in: %q", resultWithColor)
	}
}

func TestCommandHelpBuilder_NoTrailingWhitespace(t *testing.T) {
	// Test that there's no trailing whitespace before the final newline
	builder := CommandHelpBuilder{
		LongDesc: "This is a description.",
	}

	result := builder.Build()
	resultTrimmed := strings.TrimRight(result, "\n")

	if strings.HasSuffix(resultTrimmed, " ") || strings.HasSuffix(resultTrimmed, "\t") {
		t.Errorf("CommandHelpBuilder.Build() has trailing whitespace before the final newline: %q", result)
	}
}

func TestCommandHelpBuilder_EmptyOutput(t *testing.T) {
	// Test the case of a completely empty builder
	builder := CommandHelpBuilder{}

	result := builder.Build()

	// An empty builder should either return an empty string or just a newline
	if result != "" && result != "\n" {
		t.Errorf("Empty builder should return an empty string or just a newline, got: %q", result)
	}
}
