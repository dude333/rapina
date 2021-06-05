// +build !dev

package server

import "embed"

//go:embed templates
var _fs embed.FS
