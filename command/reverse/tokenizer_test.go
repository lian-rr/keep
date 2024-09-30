package reverse

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicTokenizer_tokenizeStr(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		exclusions set
		output     set
	}{
		{
			name:  "with no exclusions",
			input: " This is a test for tokens ",
			output: set{
				"This":   {},
				"test":   {},
				"for":    {},
				"tokens": {},
			},
		},
		{
			name:  "with exclusions",
			input: " This is a test for tokens ",
			exclusions: set{
				"test": {},
			},
			output: set{
				"This":   {},
				"for":    {},
				"tokens": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok := basicTokenizer{
				minLenght: 3,
			}

			got := tok.tokenizeStr(tt.input, tt.exclusions)
			assert.Equal(t, tt.output, got, "output not the expected")
		})
	}
}

func TestBasicTokenizer_stemsFilter(t *testing.T) {
	expected := set{
		"dock":   {},
		"docker": {},
		"fish":   {},
	}

	input := map[string]struct{}{
		"docking": {},
		"docker":  {},
		"fishing": {},
		"fish":    {},
	}

	tokenz := basicTokenizer{}
	got, err := tokenz.stemsFilter(input)
	require.NoError(t, err, "unexpected error")

	assert.Equal(t, expected, got, "stems unexpected")
}
