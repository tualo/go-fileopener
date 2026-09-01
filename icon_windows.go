//go:build windows

package main

import _ "embed"

//go:embed assets/app-icon.ico
var appIcon []byte

const appIconName = "app-icon.ico"
