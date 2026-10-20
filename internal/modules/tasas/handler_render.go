package tasas

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"tasasgo/internal/modules/tasas/render"
	"tasasgo/internal/modules/tasas/store"
)

// Calcular responde GET /calcular/:socio con la cartelera ya normalizada.
//
// Devuelve exactamente la vista que alimenta al dibujo, así que el panel ve lo
// que se publicaría con ese lote: una sola definición de la cartelera para la
// imagen y para la API.
func (d *Deps) Calcular(c *fiber.Ctx) error {
	cartelera, err := d.carteleraDe(c.Context(), c.Params("socio"))
	if err != nil {
		return errorDeSocio(c, err)
	}
	return ok(c, fiber.Map{"datos": render.Normalizar(cartelera)})
}

// Render responde GET /render/:socio y POST /render con la imagen PNG.
//
// El PNG se dibuja en memoria, sin servicio externo. Un socio inexistente responde
// 404 y un fallo del dibujo 500, ambos en texto plano (ver respuesta_render.go): es
// el único endpoint que no habla JSON.
func (d *Deps) Render(c *fiber.Ctx) error {
	cartelera, err := d.carteleraDe(c.Context(), nombreDeRender(c))
	if errors.Is(err, store.ErrNotFound) {
		return sinSocio(c)
	}
	if err != nil {
		return sinImagen(c)
	}
	png, err := render.GenerarCartelera(cartelera)
	if err != nil {
		return sinImagen(c)
	}
	return c.Type("png").Send(png)
}
