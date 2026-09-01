//go:build darwin

package main

import _ "embed"

//go:embed assets/app-icon.icns
var appIcon []byte

const appIconName = "app-icon.icns"
