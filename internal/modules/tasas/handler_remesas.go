package tasas

import (
	"github.com/gofiber/fiber/v2"

	"tasasgo/internal/modules/tasas/store"
)

// NotificarRemesas responde POST /notificar-remesas.
//
// Registra una notificación del lote vigente: el flujo de remesas cuelga de los
// disparadores de esa tabla, no de esta llamada, así que aquí no hay más trabajo que
// dejar la fila.
//
// El original devolvía la fila insertada completa; el store solo devuelve el
// identificador, que es lo que el disparador necesita y lo único que esta capa puede
// afirmar sin conocer el resto del esquema.
func (d *Deps) NotificarRemesas(c *fiber.Ctx) error {
	lotes, err := store.RecentLotes(c.Context(), d.Pool)
	if err != nil {
		return errorInterno(c, err)
	}
	id := "T000"
	if len(lotes) > 0 {
		id = identificadorDe(lotes[0])
	}
	notificacion, err := store.RecordNotification(c.Context(), d.Pool, id)
	if err != nil {
		return errorInterno(c, err)
	}
	return ok(c, fiber.Map{"notificacion": fiber.Map{"id_tasa": notificacion}})
}
