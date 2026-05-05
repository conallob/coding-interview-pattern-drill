package server

import "embed"

//go:embed static/index.html
var staticFiles embed.FS
