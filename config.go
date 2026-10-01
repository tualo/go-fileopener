package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type config struct {
	Folders map[string]string `json:"folders"`
	Scheme  string            `json:"scheme,omitempty"`
}

func (cfg config) urlScheme() string {
	if cfg.Scheme == "" {
		return defaultURLScheme
	}
	return cfg.Scheme
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("Konfigurationsordner konnte nicht ermittelt werden: %w", err)
	}
	return filepath.Join(dir, "tualo-fileopener", "config.json"), nil
}

func loadConfig(path string) (config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return config{Folders: make(map[string]string)}, nil
	}
	if err != nil {
		return config{}, fmt.Errorf("Konfiguration konnte nicht gelesen werden: %w", err)
	}

	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return config{}, fmt.Errorf("Konfiguration ist ungueltig: %w", err)
	}
	if cfg.Folders == nil {
		cfg.Folders = make(map[string]string)
	}
	if cfg.Scheme != "" {
		cfg.Scheme, err = normalizeScheme(cfg.Scheme)
		if err != nil {
			return config{}, fmt.Errorf("Schema in der Konfiguration ist ungueltig: %w", err)
		}
	}
	return cfg, nil
}

func saveConfig(path string, cfg config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("Konfigurationsordner konnte nicht erstellt werden: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("Konfiguration konnte nicht erzeugt werden: %w", err)
	}
	data = append(data, '\n')

	temporaryPath := path + ".tmp"
	if err := os.WriteFile(temporaryPath, data, 0o600); err != nil {
		return fmt.Errorf("Konfiguration konnte nicht geschrieben werden: %w", err)
	}
	if runtime.GOOS == "windows" {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			_ = os.Remove(temporaryPath)
			return fmt.Errorf("Alte Konfiguration konnte nicht ersetzt werden: %w", err)
		}
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("Konfiguration konnte nicht gespeichert werden: %w", err)
	}
	return nil
}

func validateAlias(alias string) error {
	if strings.TrimSpace(alias) != alias || alias == "" || strings.ContainsAny(alias, `/\\`) {
		return fmt.Errorf("Alias darf nicht leer sein und keine Schraegstriche enthalten")
	}
	return nil
}

func setFolder(cfg *config, alias, path string) (string, error) {
	if err := validateAlias(alias); err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("Ordner %q ist nicht erreichbar: %w", path, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%q ist kein Ordner", path)
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("Absoluter Pfad konnte nicht ermittelt werden: %w", err)
	}

	for existingAlias := range cfg.Folders {
		if strings.EqualFold(existingAlias, alias) && existingAlias != alias {
			delete(cfg.Folders, existingAlias)
		}
	}
	cfg.Folders[alias] = filepath.Clean(absolutePath)
	return cfg.Folders[alias], nil
}

func folderForAlias(cfg config, alias string) (string, error) {
	for configuredAlias, path := range cfg.Folders {
		if strings.EqualFold(configuredAlias, alias) {
			info, err := os.Stat(path)
			if err != nil {
				return "", fmt.Errorf("Ordner fuer Alias %q ist nicht erreichbar: %w", configuredAlias, err)
			}
			if !info.IsDir() {
				return "", fmt.Errorf("Pfad fuer Alias %q ist kein Ordner", configuredAlias)
			}
			return path, nil
		}
	}
	return "", fmt.Errorf("Alias %q ist nicht konfiguriert", alias)
}

func folderForTarget(cfg config, alias, relativePath string) (string, error) {
	root, err := folderForAlias(cfg, alias)
	if err != nil {
		return "", err
	}
	if relativePath == "" {
		return root, nil
	}

	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("Stammordner fuer Alias %q konnte nicht aufgeloest werden: %w", alias, err)
	}

	target := root
	for _, segment := range strings.Split(filepath.FromSlash(relativePath), string(filepath.Separator)) {
		if segment == "" || segment == "." || segment == ".." {
			return "", fmt.Errorf("Unterordner enthaelt einen ungueltigen Pfad")
		}
		next := filepath.Join(target, segment)
		if err := os.Mkdir(next, 0o755); err != nil && !os.IsExist(err) {
			return "", fmt.Errorf("Unterordner %q konnte nicht angelegt werden: %w", relativePath, err)
		}
		next, err = filepath.EvalSymlinks(next)
		if err != nil {
			return "", fmt.Errorf("Unterordner %q ist nicht erreichbar: %w", relativePath, err)
		}
		relativeTarget, err := filepath.Rel(root, next)
		if err != nil || relativeTarget == ".." || strings.HasPrefix(relativeTarget, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("Unterordner liegt ausserhalb des konfigurierten Stammordners")
		}
		info, err := os.Stat(next)
		if err != nil {
			return "", fmt.Errorf("Unterordner %q ist nicht erreichbar: %w", relativePath, err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("Unterpfad %q ist kein Ordner", relativePath)
		}
		target = next
	}
	return target, nil
}

func removeFolder(cfg *config, alias string) error {
	for configuredAlias := range cfg.Folders {
		if strings.EqualFold(configuredAlias, alias) {
			delete(cfg.Folders, configuredAlias)
			return nil
		}
	}
	return fmt.Errorf("Alias %q ist nicht konfiguriert", alias)
}

func sortedAliases(cfg config) []string {
	aliases := make([]string, 0, len(cfg.Folders))
	for alias := range cfg.Folders {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	return aliases
}
