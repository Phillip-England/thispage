package templates

import "embed"

// AdminFS contains the built-in admin templates.
//go:embed *.html
var AdminFS embed.FS
