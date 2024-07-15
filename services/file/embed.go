package file

import (
	"embed"
)

//go:embed all:templateRepo
var StaticFiles embed.FS
