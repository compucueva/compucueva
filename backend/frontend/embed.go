package frontend

import "embed"

//go:embed *.html *.css *.js
var Static embed.FS
