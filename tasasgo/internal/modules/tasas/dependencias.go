// Package tasas es la capa HTTP del módulo: traduce peticiones de Fiber a llamadas
// del store, del calculador y del render, y devuelve las respuestas con la misma
// forma que el servicio Node.
//
// No escribe SQL, no calcula tasas y no dibuja: recibe esas piezas ya construidas y
// las conecta. Los handlers son métodos de Deps, que se inyecta desde cmd/server,
// de modo que no exista estado global y que la suite pueda montar la aplicación
// entera con dobles en memoria.
package tasas

import (
	"time"

	"tasasgo/internal/config"
	"tasasgo/internal/modules/tasas/evolution"
	"tasasgo/internal/modules/tasas/hoo"
	"tasasgo/internal/modules/tasas/store"
)

// Deps reúne las piezas que todos los handlers comparten.
//
// Config se necesita por un ajuste que solo se usa aquí: el grupo de pruebas al que
// se redirige la publicación mientras el modo prueba esté activo.
//
// Now es una función y no el reloj del sistema para que las pruebas de los endpoints
// puedan fijar el instante: la fecha publicada es parte de la respuesta.
//
// Hoo puede ser nulo mientras no exista el cliente real (ver handler_hoo.go): el
// handler responde un error explícito en lugar de provocar un pánico.
type Deps struct {
	Pool      store.Querier
	Evolution *evolution.Client
	Hoo       hoo.Borrador
	Config    *config.Config
	Now       func() time.Time
}

// NewDeps construye las dependencias con el reloj del sistema.
func NewDeps(pool store.Querier, evo *evolution.Client, borrador hoo.Borrador, cfg *config.Config) *Deps {
	return &Deps{Pool: pool, Evolution: evo, Hoo: borrador, Config: cfg, Now: time.Now}
}
