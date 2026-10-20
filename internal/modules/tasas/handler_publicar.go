package tasas

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// PublicarLote responde POST /publicar-lote y POST /disparar-lote.
//
// Las dos rutas comparten handler porque el original también las compartía:
// dispararLoteGeneral solo llamaba a publicarLoteSolo. La respuesta lleva el lote
// publicado, cuántos socios se procesaron y el detalle socio a socio, ya ordenado
// por nombre para que no dependa del azar del pool.
func (d *Deps) PublicarLote(c *fiber.Ctx) error {
	var cuerpo struct {
		ModoPruebaGlobal bool `json:"modo_prueba_global"`
	}
	if err := parsearCuerpo(c, &cuerpo); err != nil {
		return fallo(c, fiber.StatusBadRequest, "cuerpo de publicación inválido")
	}
	datos, err := d.publicar(c.Context(), cuerpo.ModoPruebaGlobal)
	if errors.Is(err, errSinLotes) {
		return fallo(c, fiber.StatusBadRequest, err.Error())
	}
	if err != nil {
		return errorInterno(c, err)
	}
	return ok(c, datos)
}
