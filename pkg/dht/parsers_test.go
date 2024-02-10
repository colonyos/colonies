package dht

import "testing"

func TestGetRootKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		want    string
		wantErr bool
	}{
		{"Valid key", "/key1/key2/key3", "key1", false},
		{"Invalid key format", "key1/key2", "", true},
		{"No prefix", "/", "", true},
		{"Max sublevels", "/key1/key2/key3/key4/key5", "key1", false},
		{"Exceed sublevels", "/key1/key2/key3/key4/key5/key6", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRootKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePrefix(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parsePrefix(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}
