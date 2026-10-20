package tasas

import (
	"github.com/gofiber/fiber/v2"
)

// nombreDeRender obtiene el socio del endpoint de imagen.
//
// El mismo handler sirve dos entradas: GET /render/:socio lleva el nombre en la ruta
// y POST /render lo lleva en el cuerpo. Se resuelve aquí, en un solo lugar, para que
// el handler no tenga que saber por dónde llegó.
func nombreDeRender(c *fiber.Ctx) string {
	if nombre := c.Params("socio"); nombre != "" {
		return nombre
	}
	var cuerpo struct {
		Socio string `json:"socio"`
	}
	if err := c.BodyParser(&cuerpo); err != nil {
		return ""
	}
	return cuerpo.Socio
}

// sinSocio responde 404 en texto plano.
//
// Este endpoint es el único que no habla JSON: su cliente es un navegador que pinta
// la imagen, y el original respondía cadenas sueltas. Conservarlas es parte de la
// paridad, no un descuido.
func sinSocio(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).SendString("Socio no encontrado")
}

// sinImagen responde 500 en texto plano, por el mismo motivo.
func sinImagen(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).SendString("Error generando PNG")
}
