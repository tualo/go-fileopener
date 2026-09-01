//go:build darwin

package main

import "os/exec"

func showError(err error) {
	script := `on run argv
set errorMessage to item 1 of argv
display dialog "Tualo File Opener konnte den Ordner nicht oeffnen." & return & return & errorMessage with title "Tualo File Opener" buttons {"OK"} default button "OK" with icon stop
end run`
	_ = exec.Command("osascript", "-e", script, "--", err.Error()).Run()
}
