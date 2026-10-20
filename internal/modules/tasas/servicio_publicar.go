package tasas

import (
	"context"
	"errors"
	"sort"

	"github.com/gofiber/fiber/v2"

	"tasasgo/internal/modules/tasas/calculator"
	"tasasgo/internal/modules/tasas/store"
)

// errSinLotes es el fallo de publicar sin nada guardado, y es centinela para que el
// handler responda 400 y no un 500: no es un error del servidor, llega antes que los datos.
var errSinLotes = errors.New("No hay lotes guardados en tasas_glaukov para publicar.")

// publicar envía el lote vigente a todos los socios activos, comparando contra el
// mismo lote anterior. Los resultados se ordenan por nombre porque el pool termina en
// orden impredecible y el panel no debe ver la lista cambiar entre publicaciones.
func (d *Deps) publicar(ctx context.Context, modoPruebaGlobal bool) (fiber.Map, error) {
	lotes, err := store.RecentLotes(ctx, d.Pool)
	if err != nil {
		return nil, err
	}
	if len(lotes) == 0 {
		return nil, errSinLotes
	}
	actual, anterior := loteVigente(lotes)

	socios, err := store.AllSocios(ctx, d.Pool)
	if err != nil {
		return nil, err
	}
	detalle := d.enParalelo(ctx, activosDe(socios), func(ctx context.Context, socio store.Socio) resultado {
		cartelera := calculator.Build(socio.Perfil, actual, anterior, d.Now())
		if _, err := d.enviarCartelera(ctx, socio, cartelera, modoPruebaGlobal); err != nil {
			return resultado{Socio: socio.Nombre, Status: "error", Error: err.Error()}
		}
		return resultado{Socio: socio.Nombre, Status: "ok"}
	})
	sort.Slice(detalle, func(i, j int) bool { return detalle[i].Socio < detalle[j].Socio })

	return fiber.Map{"id_tasa": actual.IDTasa, "procesados": len(detalle), "detalle": detalle}, nil
}

// activosDe descarta los socios apagados.
//
// Se filtra antes de publicar, como el original: no aparecen en el detalle y tampoco
// cuentan como procesados.
func activosDe(socios []store.Socio) []store.Socio {
	activos := make([]store.Socio, 0, len(socios))
	for _, socio := range socios {
		if socio.Activo != nil && !*socio.Activo {
			continue
		}
		activos = append(activos, socio)
	}
	return activos
}
