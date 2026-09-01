//go:build windows

package main

import (
	"fmt"
	"os/exec"
)

func registerProtocol(executable string) error {
	key := `HKCU\Software\Classes\` + urlScheme
	commands := [][]string{
		{"add", key, "/ve", "/d", "URL:Tualo File Opener", "/f"},
		{"add", key, "/v", "URL Protocol", "/d", "", "/f"},
		{"add", key + `\shell\open\command`, "/ve", "/d", `"` + executable + `" open "%1"`, "/f"},
	}
	for _, args := range commands {
		if output, err := exec.Command("reg.exe", args...).CombinedOutput(); err != nil {
			return fmt.Errorf("Schema konnte nicht registriert werden: %w (%s)", err, output)
		}
	}
	return nil
}

func unregisterProtocol() error {
	output, err := exec.Command("reg.exe", "delete", `HKCU\Software\Classes\`+urlScheme, "/f").CombinedOutput()
	if err != nil {
		return fmt.Errorf("Registrierung konnte nicht entfernt werden: %w (%s)", err, output)
	}
	return nil
}
