package main

import (
	"fmt"
	"net/url"
	"strings"
)

const urlScheme = "tualo-fs"

func aliasFromURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("Link ist ungueltig: %w", err)
	}
	if !strings.EqualFold(parsed.Scheme, urlScheme) {
		return "", fmt.Errorf("Schema muss %q sein", urlScheme)
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
