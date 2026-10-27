// Package migrations incrusta las migraciones goose en el binario.
// FS acepta "" o "." como raíz para ser compatible con goose.SetBaseFS.
package migrations

import (
	"embed"
	"io/fs"
)

//go:embed *.sql
var files embed.FS

// FS es el filesystem de migraciones con raíz normalizada.
var FS fs.FS = rootFS{files}

type rootFS struct{ fsys embed.FS }

func (r rootFS) Open(name string) (fs.File, error) {
	return r.fsys.Open(normalize(name))
}

func (r rootFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return r.fsys.ReadDir(normalize(name))
}

func normalize(name string) string {
	if name == "" {
		return "."
	}
	return name
}
