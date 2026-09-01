//go:build linux

package main

import _ "embed"

//go:embed assets/app-icon.png
var appIcon []byte

const appIconName = "app-icon.png"
