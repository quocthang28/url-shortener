package shortener

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name string
		n    int
	}{
		{name: "zero length", n: 0},
		{name: "single char", n: 1},
		{name: "short code", n: 7},
		{name: "long code", n: 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Generate(tt.n)

			// length matches the request
			if len(got) != tt.n {
				t.Errorf("Generate(%d) length = %d, want %d", tt.n, len(got), tt.n)
			}

			// every character is in the base62 alphabet
			for i, c := range got {
				if !strings.ContainsRune(Alphabet, c) {
					t.Errorf("Generate(%d)[%d] = %q, not in Alphabet", tt.n, i, c)
				}
			}
		})
	}
}
