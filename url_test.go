package main

import "testing"

func TestAliasFromURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		want    string
		wantErr bool
	}{
		{name: "host", rawURL: "tualo-fs://OrdnerA", want: "OrdnerA"},
		{name: "escaped alias", rawURL: "tualo-fs:Mein%20Ordner", want: "Mein Ordner"},
		{name: "opaque", rawURL: "tualo-fs:OrdnerA", want: "OrdnerA"},
		{name: "wrong scheme", rawURL: "https://OrdnerA", wantErr: true},
		{name: "port", rawURL: "tualo-fs://OrdnerA:1234", wantErr: true},
		{name: "path traversal", rawURL: "tualo-fs://OrdnerA/../geheim", wantErr: true},
		{name: "query", rawURL: "tualo-fs://OrdnerA?path=geheim", wantErr: true},
		{name: "missing alias", rawURL: "tualo-fs://", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := aliasFromURL(test.rawURL)
			if (err != nil) != test.wantErr {
				t.Fatalf("aliasFromURL(%q) error = %v, wantErr %v", test.rawURL, err, test.wantErr)
			}
			if got != test.want {
				t.Errorf("aliasFromURL(%q) = %q, want %q", test.rawURL, got, test.want)
			}
		})
	}
}
