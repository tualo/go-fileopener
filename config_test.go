package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigRoundTripAndCaseInsensitiveLookup(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "nested", "config.json")
	folder := t.TempDir()
	cfg := config{Folders: make(map[string]string)}

	storedPath, err := setFolder(&cfg, "OrdnerA", folder)
	if err != nil {
		t.Fatalf("setFolder() error = %v", err)
	}
	if err := saveConfig(configFile, cfg); err != nil {
		t.Fatalf("saveConfig() error = %v", err)
	}
	loaded, err := loadConfig(configFile)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	got, err := folderForAlias(loaded, "ordnera")
	if err != nil {
		t.Fatalf("folderForAlias() error = %v", err)
	}
	if got != storedPath {
		t.Errorf("folderForAlias() = %q, want %q", got, storedPath)
	}
}

func TestFolderForAliasRejectsMissingDirectory(t *testing.T) {
	cfg := config{Folders: map[string]string{"Alt": filepath.Join(t.TempDir(), "missing")}}
	if _, err := folderForAlias(cfg, "Alt"); err == nil {
		t.Fatal("folderForAlias() error = nil, want missing-directory error")
	}
}

func TestSetFolderReplacesAliasWithDifferentCase(t *testing.T) {
	cfg := config{Folders: map[string]string{"OrdnerA": t.TempDir()}}
	newFolder := t.TempDir()
	if _, err := setFolder(&cfg, "ordnera", newFolder); err != nil {
		t.Fatalf("setFolder() error = %v", err)
	}
	if len(cfg.Folders) != 1 || cfg.Folders["ordnera"] == "" {
		t.Fatalf("setFolder() folders = %#v, want one lowercase alias", cfg.Folders)
	}
}

func TestLoadConfigRejectsInvalidScheme(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configFile, []byte(`{"scheme":"invalid_schema"}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, err := loadConfig(configFile); err == nil {
		t.Fatal("loadConfig() error = nil, want invalid-scheme error")
	}
}

func TestConfigUsesDefaultScheme(t *testing.T) {
	if got := (config{}).urlScheme(); got != defaultURLScheme {
		t.Errorf("urlScheme() = %q, want %q", got, defaultURLScheme)
	}
}

func TestFolderForTargetResolvesSubdirectory(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, "FREIEPRE", "Projekt mit Leerzeichen")
	if err := os.MkdirAll(want, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	want, err := filepath.EvalSymlinks(want)
	if err != nil {
		t.Fatalf("EvalSymlinks() error = %v", err)
	}
	cfg := config{Folders: map[string]string{"TFS": root}}
	got, err := folderForTarget(cfg, "TFS", "FREIEPRE/Projekt mit Leerzeichen")
	if err != nil {
		t.Fatalf("folderForTarget() error = %v", err)
	}
	if got != want {
		t.Errorf("folderForTarget() = %q, want %q", got, want)
	}
}

func TestFolderForTargetCreatesMissingSubdirectory(t *testing.T) {
	root := t.TempDir()
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("EvalSymlinks() error = %v", err)
	}
	want := filepath.Join(root, "FREIEPRE", "Neues Projekt")
	cfg := config{Folders: map[string]string{"TFS": root}}

	got, err := folderForTarget(cfg, "TFS", "FREIEPRE/Neues Projekt")
	if err != nil {
		t.Fatalf("folderForTarget() error = %v", err)
	}
	if got != want {
		t.Errorf("folderForTarget() = %q, want %q", got, want)
	}
	info, err := os.Stat(got)
	if err != nil {
		t.Fatalf("created folder is not accessible: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("created path %q is not a directory", got)
	}
}

func TestFolderForTargetRejectsSymlinkOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "outside")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("Symlink konnte nicht erstellt werden: %v", err)
	}
	cfg := config{Folders: map[string]string{"TFS": root}}
	if _, err := folderForTarget(cfg, "TFS", "outside"); err == nil {
		t.Fatal("folderForTarget() error = nil, want symlink escape error")
	}
}
