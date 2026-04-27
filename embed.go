package embedded

import "embed"

//go:embed all:assets
var Assets embed.FS

//go:embed all:templates
var Templates embed.FS
