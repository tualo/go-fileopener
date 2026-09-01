//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const desktopFileName = "tualo-fileopener.desktop"

func registerProtocol(executable string) error {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dataDir = filepath.Join(home, ".local", "share")
	}
	applicationsDir := filepath.Join(dataDir, "applications")
	if err := os.MkdirAll(applicationsDir, 0o700); err != nil {
		return fmt.Errorf("Anwendungsordner konnte nicht erstellt werden: %w", err)
	}
	entry := fmt.Sprintf("[Desktop Entry]\nType=Application\nName=Tualo File Opener\nExec=%s open %%u\nNoDisplay=true\nTerminal=false\nMimeType=x-scheme-handler/%s;\n", quoteDesktopArgument(executable), urlScheme)
	if err := os.WriteFile(filepath.Join(applicationsDir, desktopFileName), []byte(entry), 0o644); err != nil {
		return fmt.Errorf("Desktop-Eintrag konnte nicht geschrieben werden: %w", err)
	}
	if err := exec.Command("xdg-mime", "default", desktopFileName, "x-scheme-handler/"+urlScheme).Run(); err != nil {
		return fmt.Errorf("Schema konnte mit xdg-mime nicht registriert werden: %w", err)
	}
	if updateDatabase, err := exec.LookPath("update-desktop-database"); err == nil {
		_ = exec.Command(updateDatabase, applicationsDir).Run()
	}
	return nil
}

func unregisterProtocol() error {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dataDir = filepath.Join(home, ".local", "share")
	}
	path := filepath.Join(dataDir, "applications", desktopFileName)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Desktop-Eintrag konnte nicht entfernt werden: %w", err)
	}
	return nil
}

func quoteDesktopArgument(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "`", "\\`", "$", "\\$")
	return `"` + replacer.Replace(value) + `"`
}
