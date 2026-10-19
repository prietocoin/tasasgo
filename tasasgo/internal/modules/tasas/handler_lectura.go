package tasas

import (
	"github.com/gofiber/fiber/v2"

	"tasasgo/internal/modules/tasas/dominio"
	"tasasgo/internal/modules/tasas/store"
)

// Monitor responde GET /monitor con el lote vigente, que es lo que el panel muestra
// como "tasas actuales".
//
// Con la tabla vacía responde T000 y un objeto sin tasas: el original devolvía el
// mismo hueco (`row.id_tasa || 'T000'`) y el panel lo dibuja como un lote sin
// publicar, en lugar de mostrar un error.
func (d *Deps) Monitor(c *fiber.Ctx) error {
	lotes, err := store.RecentLotes(c.Context(), d.Pool)
	if err != nil {
		return errorInterno(c, err)
	}
	if len(lotes) == 0 {
		return ok(c, fiber.Map{"id_tasa": "T000", "tasas": fiber.Map{}})
	}
	return ok(c, fiber.Map{"id_tasa": identificadorDe(lotes[0]), "tasas": lotes[0].Tasas})
}

// Socios responde GET /socios con la lista que el panel pinta.
//
// Aplica el rol por defecto (el original respondía SOCIO cuando la columna venía
// nula) y garantiza que monedas sea un objeto y no un nulo, porque el panel itera
// sobre él sin comprobarlo.
func (d *Deps) Socios(c *fiber.Ctx) error {
	socios, err := store.AllSocios(c.Context(), d.Pool)
	if err != nil {
		return errorInterno(c, err)
	}
	for i := range socios {
		if socios[i].Rol == "" {
			socios[i].Rol = "SOCIO"
		}
		if socios[i].Monedas == nil {
			socios[i].Monedas = map[string]*dominio.Moneda{}
		}
	}
	return ok(c, fiber.Map{"socios": socios})
}

// identificadorDe devuelve el identificador visible del lote, con T000 cuando no
// viene: la misma referencia que publica el cálculo.
func identificadorDe(lote dominio.Lote) string {
	if lote.IDTasa == "" {
		return "T000"
	}
	return lote.IDTasa
}
