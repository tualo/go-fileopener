package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

func openFolder(path string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", path)
	case "windows":
		command = exec.Command("explorer.exe", path)
	case "linux":
		command = exec.Command("xdg-open", path)
	default:
		return fmt.Errorf("Betriebssystem %q wird nicht unterstuetzt", runtime.GOOS)
	}
	if err := command.Start(); err != nil {
		return fmt.Errorf("Ordner konnte nicht geoeffnet werden: %w", err)
	}
	return nil
}
