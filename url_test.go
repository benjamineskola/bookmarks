package main

import (
	"net/url"
	"testing"

	"github.com/benjamineskola/bookmarks/config"
	"github.com/stretchr/testify/assert"
)

func TestURLAddWWW(t *testing.T) {
	t.Parallel()

	config.Config = config.MakeConfig()
	config.Config.URLNormalisations.AddWWW = []string{"theguardian.com"}

	testCases := []struct {
		Input    string
		Expected string
	}{
		{Input: "https://theguardian.com", Expected: "https://www.theguardian.com"},
		{Input: "https://www.theguardian.com", Expected: "https://www.theguardian.com"},
		{Input: "https://nottheguardian.com", Expected: "https://nottheguardian.com"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Input, func(t *testing.T) {
			t.Parallel()

			expected, _ := url.Parse(tc.Expected)
			input, _ := url.Parse(tc.Input)
			actual := normaliseURL(*input)
			assert.Equal(t, *expected, actual)
		})
	}
}
