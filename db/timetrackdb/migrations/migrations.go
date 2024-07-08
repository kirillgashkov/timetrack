package migrations

import (
	"embed"
	"io/fs"
)

//go:embed *.sql
var migrations embed.FS

func FS() fs.FS {
	return migrations
}
