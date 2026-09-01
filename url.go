package main

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

const defaultURLScheme = "tualo-fs"

type linkTarget struct {
	Alias        string
	RelativePath string
}

func normalizeScheme(scheme string) (string, error) {
	if scheme == "" {
		return "", fmt.Errorf("Schema darf nicht leer sein")
	}
	for index, character := range scheme {
		if index == 0 && !unicode.IsLetter(character) {
			return "", fmt.Errorf("Schema muss mit einem Buchstaben beginnen")
		}
		if !unicode.IsLetter(character) && !unicode.IsDigit(character) && !strings.ContainsRune("+.-", character) {
			return "", fmt.Errorf("Schema %q enthaelt ungueltige Zeichen", scheme)
		}
		if character > unicode.MaxASCII {
			return "", fmt.Errorf("Schema darf nur ASCII-Zeichen enthalten")
		}
	}
	return strings.ToLower(scheme), nil
}

func targetFromURL(rawURL, expectedScheme string) (linkTarget, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return linkTarget{}, fmt.Errorf("Link ist ungueltig: %w", err)
	}
	if !strings.EqualFold(parsed.Scheme, expectedScheme) {
		return linkTarget{}, fmt.Errorf("Schema muss %q sein", expectedScheme)
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return linkTarget{}, fmt.Errorf("Link darf keine Zugangsdaten, Parameter oder Fragmente enthalten")
	}
	if parsed.Port() != "" {
		return linkTarget{}, fmt.Errorf("Link darf keinen Port enthalten")
	}

	alias := parsed.Host
	if alias == "" && parsed.Opaque != "" {
		alias = parsed.Opaque
	}
	if alias == "" || strings.ContainsAny(alias, `/\\`) {
		return linkTarget{}, fmt.Errorf("Link enthaelt keinen gueltigen Ordner-Alias")
	}

	decoded, err := url.PathUnescape(alias)
	if err != nil || decoded == "" || strings.ContainsAny(decoded, `/\\`) {
		return linkTarget{}, fmt.Errorf("Link enthaelt keinen gueltigen Ordner-Alias")
	}

	relativePath := strings.Trim(parsed.Path, "/")
	if relativePath != "" {
		segments := strings.Split(relativePath, "/")
		for _, segment := range segments {
			if segment == "" || segment == "." || segment == ".." || strings.Contains(segment, `\`) {
				return linkTarget{}, fmt.Errorf("Link enthaelt einen ungueltigen Unterpfad")
			}
		}
	}
	return linkTarget{Alias: decoded, RelativePath: relativePath}, nil
}
