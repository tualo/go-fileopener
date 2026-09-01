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
			got, err := aliasFromURL(test.rawURL, defaultURLScheme)
			if (err != nil) != test.wantErr {
				t.Fatalf("aliasFromURL(%q) error = %v, wantErr %v", test.rawURL, err, test.wantErr)
			}
			if got != test.want {
				t.Errorf("aliasFromURL(%q) = %q, want %q", test.rawURL, got, test.want)
			}
		})
	}
}

func TestAliasFromURLWithCustomScheme(t *testing.T) {
	got, err := aliasFromURL("local-folders://OrdnerA", "local-folders")
	if err != nil {
		t.Fatalf("aliasFromURL() error = %v", err)
	}
	if got != "OrdnerA" {
		t.Errorf("aliasFromURL() = %q, want %q", got, "OrdnerA")
	}
}

func TestNormalizeScheme(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{input: "Tualo-FS", want: "tualo-fs"},
		{input: "local+folder.v2", want: "local+folder.v2"},
		{input: "2folders", wantErr: true},
		{input: "local_folder", wantErr: true},
		{input: "", wantErr: true},
	}
	for _, test := range tests {
		got, err := normalizeScheme(test.input)
		if (err != nil) != test.wantErr {
			t.Errorf("normalizeScheme(%q) error = %v, wantErr %v", test.input, err, test.wantErr)
		}
		if got != test.want {
			t.Errorf("normalizeScheme(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}
