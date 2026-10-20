package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"tasasgo/internal/config"
	"tasasgo/internal/modules/tasas"
	"tasasgo/internal/modules/tasas/evolution"
	"tasasgo/internal/modules/tasas/hoo"
	"tasasgo/internal/web"
)

// raizWeb es el directorio de los archivos del panel, común a los dos entornos:
// en desarrollo el servicio se ejecuta desde la carpeta del módulo y en el
// contenedor el directorio de trabajo es "/", donde el Dockerfile copia public/.
// El panel deja de necesitar un servidor Node propio por ser, precisamente, estos
// archivos servidos desde aquí.
const raizWeb = "./public"

// registerRoutes declara la superficie HTTP del servicio.
//
// No contiene lógica de negocio: compone las dependencias reales, le entrega el
// mapa de rutas al módulo de tasas y monta el panel. El PNG se dibuja en memoria.
//
// El orden importa y es la única sutileza del archivo: el panel va antes que el
// 404 porque tiene que poder responder lo que nadie reclamó, pero después de la
// API porque responde a cualquier ruta y taparía /api/v1/tasas.
func registerRoutes(app *fiber.App, cfg *config.Config, pool *pgxpool.Pool) {
	app.Get("/health", healthHandler(cfg, pool))
	tasas.Register(app, tasas.NewDeps(
		pool,
		evolution.New(cfg.EvolutionAPIURL, cfg.EvolutionAPIKey, cfg.EvolutionInstance),
		hoo.New(cfg.HooURL),
		cfg,
	))
	web.Register(app, raizWeb)
	app.Use(notFound)
}

// notFound responde en JSON a cualquier ruta desconocida.
//
// Fiber responde texto plano por defecto, y el dashboard espera cuerpos JSON
// homogéneos: conservar el mismo formato que el servicio Node evita que el
// cliente tenga que interpretar dos idiomas distintos según quién responda.
func notFound(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"success": false,
		"error":   "ruta no encontrada: " + c.Path(),
	})
}
