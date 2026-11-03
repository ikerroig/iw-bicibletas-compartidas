// Package migrations embebe las migraciones SQL para que el migrador y los
// tests de integración apliquen exactamente el mismo esquema compartido.
package migrations

import "embed"

// FS contiene todos los ficheros *.sql de esta carpeta.
//
//go:embed *.sql
var FS embed.FS
