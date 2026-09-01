//go:build linux

package main

import "os/exec"

func showError(err error) {
	message := "Tualo File Opener konnte den Ordner nicht oeffnen.\n\n" + err.Error()
	if zenity, lookupErr := exec.LookPath("zenity"); lookupErr == nil {
		_ = exec.Command(zenity, "--error", "--title=Tualo File Opener", "--text="+message).Run()
		return
	}
	if notifySend, lookupErr := exec.LookPath("notify-send"); lookupErr == nil {
		_ = exec.Command(notifySend, "--urgency=critical", "Tualo File Opener", message).Run()
	}
}
