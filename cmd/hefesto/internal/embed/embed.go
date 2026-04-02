// Package embed provides embedded filesystem access to Hefesto configuration files.
package embed

import "embed"

// ConfigFiles contains all configuration files from HefestoOpenCode
// embedded directly into the binary for self-contained distribution.
//
//go:embed all:config
var ConfigFiles embed.FS
