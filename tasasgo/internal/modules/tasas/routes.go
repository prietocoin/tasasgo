package tasas

import "github.com/gofiber/fiber/v2"

// Register monta la superficie HTTP del módulo bajo /api/v1/tasas.
//
// Las trece rutas son las que declaraba tasas.routes.js, con los mismos métodos y el
// mismo orden de segmentos. Se conservan dos detalles del original: /render atiende
// GET y POST con el mismo handler, y /disparar-lote apunta al mismo código que
// /publicar-lote, porque en el servicio Node eran alias.
//
// Recibe las dependencias ya construidas: este archivo no decide de dónde salen.
func Register(app *fiber.App, deps *Deps) {
	grupo := app.Group("/api/v1/tasas")

	grupo.Get("/monitor", deps.Monitor)
	grupo.Get("/socios", deps.Socios)
	grupo.Get("/hoo-borrador", deps.BorradorHoo)

	grupo.Put("/socios/:socio", deps.GuardarSocio)
	grupo.Post("/guardar-lote", deps.GuardarLote)
	grupo.Post("/publicar-lote", deps.PublicarLote)

	grupo.Post("/notificar-remesas", deps.NotificarRemesas)

	grupo.Get("/calcular/:socio", deps.Calcular)
	grupo.Get("/render/:socio", deps.Render)
	grupo.Post("/render", deps.Render)
	grupo.Post("/disparar", deps.Disparar)
	grupo.Post("/disparar-lote", deps.PublicarLote)
	grupo.Delete("/lote/:id_tasa", deps.EliminarLote)
}
