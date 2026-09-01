//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"syscall"
)

const (
	windowsApplicationName = "Tualo File Opener"
	windowsCapabilitiesKey = `HKCU\Software\Tualo\FileOpener\Capabilities`
	windowsProgID          = "TualoFileOpener.URL"
)

func registerProtocol(executable, icon, scheme string) error {
	schemeKey := `HKCU\Software\Classes\` + scheme
	progIDKey := `HKCU\Software\Classes\` + windowsProgID
	commands := [][]string{
		{"add", schemeKey, "/ve", "/d", "URL:Tualo File Opener", "/f"},
		{"add", schemeKey, "/v", "URL Protocol", "/d", "", "/f"},
		{"add", schemeKey + `\shell\open\command`, "/ve", "/d", `"` + executable + `" open "%1"`, "/f"},
		{"add", progIDKey, "/ve", "/d", "URL:Tualo File Opener", "/f"},
		{"add", progIDKey, "/v", "URL Protocol", "/d", "", "/f"},
		{"add", progIDKey + `\DefaultIcon`, "/ve", "/d", icon, "/f"},
		{"add", progIDKey + `\shell\open\command`, "/ve", "/d", `"` + executable + `" open "%1"`, "/f"},
		{"add", windowsCapabilitiesKey, "/v", "ApplicationName", "/d", windowsApplicationName, "/f"},
		{"add", windowsCapabilitiesKey, "/v", "ApplicationDescription", "/d", "Oeffnet konfigurierte lokale Ordner ueber URL-Links.", "/f"},
		{"add", windowsCapabilitiesKey, "/v", "ApplicationIcon", "/d", icon, "/f"},
		{"add", windowsCapabilitiesKey + `\URLAssociations`, "/v", scheme, "/d", windowsProgID, "/f"},
		{"add", `HKCU\Software\RegisteredApplications`, "/v", windowsApplicationName, "/d", `Software\Tualo\FileOpener\Capabilities`, "/f"},
	}
	for _, args := range commands {
		if output, err := exec.Command("reg.exe", args...).CombinedOutput(); err != nil {
			return fmt.Errorf("Schema konnte nicht registriert werden: %w (%s)", err, output)
		}
	}
	notifyAssociationChanged()
	return nil
}

func unregisterProtocol(scheme string) error {
	values := [][2]string{
		{windowsCapabilitiesKey + `\URLAssociations`, scheme},
		{`HKCU\Software\RegisteredApplications`, windowsApplicationName},
	}
	for _, value := range values {
		if err := deleteRegistryValue(value[0], value[1]); err != nil {
			return err
		}
	}
	keys := []string{
		windowsCapabilitiesKey,
		`HKCU\Software\Classes\` + windowsProgID,
		`HKCU\Software\Classes\` + scheme,
	}
	for _, key := range keys {
		if err := deleteRegistryKey(key); err != nil {
			return err
		}
	}
	notifyAssociationChanged()
	return nil
}

func deleteRegistryValue(key, name string) error {
	if err := exec.Command("reg.exe", "query", key, "/v", name).Run(); err != nil {
		return nil
	}
	output, err := exec.Command("reg.exe", "delete", key, "/v", name, "/f").CombinedOutput()
	if err != nil {
		return fmt.Errorf("Registry-Wert konnte nicht entfernt werden: %w (%s)", err, output)
	}
	return nil
}

func deleteRegistryKey(key string) error {
	if err := exec.Command("reg.exe", "query", key).Run(); err != nil {
		return nil
	}
	output, err := exec.Command("reg.exe", "delete", key, "/f").CombinedOutput()
	if err != nil {
		return fmt.Errorf("Registry-Schluessel konnte nicht entfernt werden: %w (%s)", err, output)
	}
	return nil
}

func notifyAssociationChanged() {
	const (
		shcneAssocChanged = 0x08000000
		shcnfIDList       = 0x0000
		shcnfFlush        = 0x1000
	)
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shChangeNotify := shell32.NewProc("SHChangeNotify")
	_, _, _ = shChangeNotify.Call(
		uintptr(shcneAssocChanged),
		uintptr(shcnfIDList|shcnfFlush),
		0,
		0,
	)
}
