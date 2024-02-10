package dht

import "testing"

func TestIsValidKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"/key1/key2/key3", true},                      // Valid: exactly 3 sublevels
		{"/key1/key2/key3/key4/key5", true},            // Valid: exactly 5 sublevels
		{"/key1/", false},                              // Invalid: ends with a slash
		{"/key1/key2/key3/key4/key5/key6", false},      // Invalid: more than 5 sublevels
		{"key1/key2", false},                           // Invalid: no leading slash
		{"/key1/key2//key3", false},                    // Invalid: double slash
		{"/key1/key2/key3/", false},                    // Invalid: trailing slash
		{"/key1/key2/key3/key4/key5/key6/key7", false}, // Invalid: more than 5 sublevels
		{"/", false},                                   // Invalid: no sublevels
		{"/key1key2/key3", true},                       // Valid: alphanumeric keys
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, err := isValidKey(tt.key)
			if err != nil {
				t.Fatalf("isValidKey(%q) returned an unexpected error: %v", tt.key, err)
			}
			if got != tt.want {
				t.Errorf("isValidKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}
