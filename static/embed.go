package static

import "embed"

// AdminFS contains the built-in admin static assets.
//go:embed *.css
var AdminFS embed.FS
