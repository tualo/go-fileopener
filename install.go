package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func installProtocol() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Programmpfad konnte nicht ermittelt werden: %w", err)
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return fmt.Errorf("Programmpfad konnte nicht aufgeloest werden: %w", err)
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("Konfigurationsordner konnte nicht ermittelt werden: %w", err)
	}
	name := "fileopener"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	installedExecutable := filepath.Join(configDir, "tualo-fileopener", "bin", name)
	if filepath.Clean(executable) != filepath.Clean(installedExecutable) {
		if err := copyExecutable(executable, installedExecutable); err != nil {
			return err
		}
	}
	return registerProtocol(installedExecutable)
}

func copyExecutable(source, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return fmt.Errorf("Installationsordner konnte nicht erstellt werden: %w", err)
	}
	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("Programm konnte nicht gelesen werden: %w", err)
	}
	defer input.Close()

	temporaryPath := destination + ".tmp"
	output, err := os.OpenFile(temporaryPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("Programm konnte nicht installiert werden: %w", err)
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("Programm konnte nicht kopiert werden: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("Programm konnte nicht gespeichert werden: %w", closeErr)
	}
	if err := os.Rename(temporaryPath, destination); err != nil {
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("Programm konnte nicht aktiviert werden: %w", err)
	}
	return nil
}
