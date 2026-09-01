//go:build darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func registerProtocol(executable string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	appPath := filepath.Join(home, "Applications", "Tualo File Opener.app")
	if err := os.MkdirAll(filepath.Dir(appPath), 0o755); err != nil {
		return fmt.Errorf("Applications-Ordner konnte nicht erstellt werden: %w", err)
	}
	_ = os.RemoveAll(appPath)
	script := fmt.Sprintf("on open location theURL\n do shell script quoted form of %q & \" open \" & quoted form of theURL\nend open location", executable)
	if output, err := exec.Command("/usr/bin/osacompile", "-o", appPath, "-e", script).CombinedOutput(); err != nil {
		return fmt.Errorf("macOS-App konnte nicht erstellt werden: %w (%s)", err, strings.TrimSpace(string(output)))
	}

	plist := filepath.Join(appPath, "Contents", "Info.plist")
	plistBuddy := "/usr/libexec/PlistBuddy"
	commands := []string{
		"Add :CFBundleIdentifier string de.tualo.fileopener",
		"Add :CFBundleURLTypes array",
		"Add :CFBundleURLTypes:0 dict",
		"Add :CFBundleURLTypes:0:CFBundleURLName string Tualo File Opener URL",
		"Add :CFBundleURLTypes:0:CFBundleURLSchemes array",
		"Add :CFBundleURLTypes:0:CFBundleURLSchemes:0 string " + urlScheme,
	}
	for _, command := range commands {
		if output, err := exec.Command(plistBuddy, "-c", command, plist).CombinedOutput(); err != nil {
			return fmt.Errorf("App-Registrierung konnte nicht konfiguriert werden: %w (%s)", err, strings.TrimSpace(string(output)))
		}
	}
	if output, err := exec.Command("/usr/bin/codesign", "--force", "--deep", "--sign", "-", appPath).CombinedOutput(); err != nil {
		return fmt.Errorf("macOS-App konnte nicht signiert werden: %w (%s)", err, strings.TrimSpace(string(output)))
	}
	lsregister := "/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"
	if output, err := exec.Command(lsregister, "-f", appPath).CombinedOutput(); err != nil {
		return fmt.Errorf("Schema konnte nicht bei Launch Services registriert werden: %w (%s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func unregisterProtocol() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	appPath := filepath.Join(home, "Applications", "Tualo File Opener.app")
	lsregister := "/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"
	_ = exec.Command(lsregister, "-u", appPath).Run()
	if err := os.RemoveAll(appPath); err != nil {
		return fmt.Errorf("macOS-App konnte nicht entfernt werden: %w", err)
	}
	return nil
}
