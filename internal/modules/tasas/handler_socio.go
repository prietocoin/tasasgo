package tasas

import (
	"github.com/gofiber/fiber/v2"

	"tasasgo/internal/modules/tasas/dominio"
	"tasasgo/internal/modules/tasas/store"
)

// cambiosDeSocio es el cuerpo que envía el panel al guardar un socio.
//
// Activo, ModoPrueba y Rol son punteros porque un campo ausente no debe escribirse.
// El panel no envía el rol, y con un string simple cada guardado de monedas lo
// habría borrado.
type cambiosDeSocio struct {
	Monedas    map[string]*dominio.Moneda `json:"monedas"`
	Activo     *bool                      `json:"activo"`
	ModoPrueba *bool                      `json:"modo_prueba"`
	Rol        *string                    `json:"rol"`
}

// GuardarSocio responde PUT /socios/:socio.
//
// Un socio inexistente devuelve 404: el original respondía éxito con un socio vacío
// y el panel mostraba "guardado" sin haber guardado nada.
func (d *Deps) GuardarSocio(c *fiber.Ctx) error {
	var cambios cambiosDeSocio
	if err := parsearCuerpo(c, &cambios); err != nil {
		return fallo(c, fiber.StatusBadRequest, "cuerpo de socio inválido")
	}
	socio, err := store.UpdateSocio(c.Context(), d.Pool, c.Params("socio"), store.SocioUpdate{
		Monedas:    cambios.Monedas,
		Activo:     cambios.Activo,
		ModoPrueba: cambios.ModoPrueba,
		Rol:        cambios.Rol,
	})
	if err != nil {
		return errorDeSocio(c, err)
	}
	return ok(c, fiber.Map{"socio": socio})
}
