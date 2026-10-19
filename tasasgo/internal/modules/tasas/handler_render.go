package tasas

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"tasasgo/internal/modules/tasas/renderer"
	"tasasgo/internal/modules/tasas/store"
)

// Calcular responde GET /calcular/:socio con la cartelera ya normalizada.
//
// Devuelve el mismo objeto que se le manda al renderer, así que el panel ve
// exactamente lo que se publicaría con ese lote: una sola definición de la
// cartelera para la imagen y para la API.
func (d *Deps) Calcular(c *fiber.Ctx) error {
	cartelera, err := d.carteleraDe(c.Context(), c.Params("socio"))
	if err != nil {
		return errorDeSocio(c, err)
	}
	return ok(c, fiber.Map{"datos": renderer.Normalizar(cartelera)})
}

// Render responde GET /render/:socio y POST /render con la imagen PNG.
//
// Un socio inexistente responde 404 y un fallo del renderer 500, ambos en texto
// plano (ver respuesta_render.go): es el único endpoint que no habla JSON.
func (d *Deps) Render(c *fiber.Ctx) error {
	cartelera, err := d.carteleraDe(c.Context(), nombreDeRender(c))
	if errors.Is(err, store.ErrNotFound) {
		return sinSocio(c)
	}
	if err != nil {
		return sinImagen(c)
	}
	png, err := d.Renderer.Render(c.Context(), cartelera)
	if err != nil {
		return sinImagen(c)
	}
	return c.Type("png").Send(png)
}
