package tasas

import (
	"context"

	"tasasgo/internal/modules/tasas/calculator"
	"tasasgo/internal/modules/tasas/dominio"
	"tasasgo/internal/modules/tasas/store"
)

// carteleraDe calcula la cartelera de un socio a partir del lote vigente.
//
// Unifica lo que el controlador original repetía en cuatro endpoints: buscar el
// socio, leer los dos últimos lotes y calcular. Al vivir en un solo lugar, la
// definición de "lote vigente" no puede divergir entre /calcular, /render y
// /disparar.
//
// Un socio inexistente devuelve store.ErrNotFound, y cada handler lo traduce al
// formato que su endpoint publica.
func (d *Deps) carteleraDe(ctx context.Context, nombre string) (dominio.Cartelera, error) {
	socio, err := store.SocioByName(ctx, d.Pool, nombre)
	if err != nil {
		return dominio.Cartelera{}, err
	}
	lotes, err := store.RecentLotes(ctx, d.Pool)
	if err != nil {
		return dominio.Cartelera{}, err
	}
	actual, anterior := loteVigente(lotes)
	return calculator.Build(socio.Perfil, actual, anterior, d.Now()), nil
}

// loteVigente devuelve el lote vigente y el anterior.
//
// Con un solo lote publicado el anterior es el mismo: es lo que hacía el original
// (`rows[1] || rows[0]`), y sin eso no habría contra qué comparar la tendencia. Con
// la tabla vacía devuelve dos lotes vacíos, y el cálculo los trata como T000.
func loteVigente(lotes []dominio.Lote) (dominio.Lote, dominio.Lote) {
	if len(lotes) == 0 {
		return dominio.Lote{}, dominio.Lote{}
	}
	if len(lotes) == 1 {
		return lotes[0], lotes[0]
	}
	return lotes[0], lotes[1]
}
