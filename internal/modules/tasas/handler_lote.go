package tasas

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"tasasgo/internal/modules/tasas/store"
)

// selloFecha es el formato del sello de creación, con milisegundos y zona Z.
//
// Es el que produce `new Date().toISOString()` en el original: con el RFC3339 de Go
// el valor guardado cambiaría de forma y las comparaciones del panel fallarían.
const selloFecha = "2006-01-02T15:04:05.000Z"

// GuardarLote responde POST /guardar-lote: guarda un lote nuevo sin publicarlo, con
// el identificador derivado del último lote y la fecha sellada aquí si el panel no la
// mandó, que es lo que hacía el original.
func (d *Deps) GuardarLote(c *fiber.Ctx) error {
	var cuerpo struct {
		Tasas map[string]any `json:"tasas"`
	}
	if err := parsearCuerpo(c, &cuerpo); err != nil || cuerpo.Tasas == nil {
		return fallo(c, fiber.StatusBadRequest, "Estructura de tasas inválida.")
	}
	lotes, err := store.RecentLotes(c.Context(), d.Pool)
	if err != nil {
		return errorInterno(c, err)
	}
	ultimo := ""
	if len(lotes) > 0 {
		ultimo = lotes[0].IDTasa
	}
	if sello, esTexto := cuerpo.Tasas["CREATED_AT"].(string); !esTexto || sello == "" {
		cuerpo.Tasas["CREATED_AT"] = time.Now().UTC().Format(selloFecha)
	}
	id := store.NextLoteID(ultimo)
	lote, err := store.InsertLote(c.Context(), d.Pool, id, cuerpo.Tasas)
	if err != nil {
		return errorInterno(c, err)
	}
	return ok(c, fiber.Map{"id_tasa": id, "registro": lote})
}

// EliminarLote responde DELETE /lote/:id_tasa.
func (d *Deps) EliminarLote(c *fiber.Ctx) error {
	id := c.Params("id_tasa")
	if id == "" {
		return fallo(c, fiber.StatusBadRequest, "Debe proporcionar un ID de lote válido.")
	}
	borrado, err := store.DeleteLote(c.Context(), d.Pool, id)
	if err != nil {
		return errorInterno(c, err)
	}
	if !borrado {
		return noEncontrado(c, "No se encontró el lote "+id+".")
	}
	return ok(c, fiber.Map{"message": "Lote " + id + " eliminado correctamente."})
}
