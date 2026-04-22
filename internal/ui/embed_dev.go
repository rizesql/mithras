//go:build dev

package ui

import "io/fs"

var Embed fs.FS = nil // or os.DirFS(".") or something fake
