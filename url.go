package main

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

const defaultURLScheme = "tualo-fs"

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

func aliasFromURL(rawURL, expectedScheme string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("Link ist ungueltig: %w", err)
	}
	if !strings.EqualFold(parsed.Scheme, expectedScheme) {
		return "", fmt.Errorf("Schema muss %q sein", expectedScheme)
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("Link darf keine Zugangsdaten, Parameter oder Fragmente enthalten")
	}
	if parsed.Port() != "" {
		return "", fmt.Errorf("Link darf keinen Port enthalten")
	}

	alias := parsed.Host
	if alias == "" && parsed.Opaque != "" {
		alias = parsed.Opaque
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return "", fmt.Errorf("Link darf keinen zusaetzlichen Pfad enthalten")
	}
	if alias == "" || strings.ContainsAny(alias, `/\\`) {
		return "", fmt.Errorf("Link enthaelt keinen gueltigen Ordner-Alias")
	}

	decoded, err := url.PathUnescape(alias)
	if err != nil || decoded == "" || strings.ContainsAny(decoded, `/\\`) {
		return "", fmt.Errorf("Link enthaelt keinen gueltigen Ordner-Alias")
	}
	return decoded, nil
}
