package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"tasasgo/internal/config"
	"tasasgo/internal/modules/tasas"
	"tasasgo/internal/modules/tasas/evolution"
	"tasasgo/internal/modules/tasas/hoo"
	"tasasgo/internal/modules/tasas/renderer"
)

// registerRoutes declara la superficie HTTP del servicio.
//
// No contiene lógica de negocio: compone las dependencias reales y le entrega el
// mapa de rutas al módulo de tasas. El frontend y el renderer PNG siguen siendo
// servidos por el servicio Node, así que este proceso no monta archivos estáticos.
func registerRoutes(app *fiber.App, cfg *config.Config, pool *pgxpool.Pool) {
	app.Get("/health", healthHandler(cfg, pool))
	tasas.Register(app, tasas.NewDeps(
		pool,
		renderer.NewSidecar(cfg.RendererURL),
		evolution.New(cfg.EvolutionAPIURL, cfg.EvolutionAPIKey, cfg.EvolutionInstance),
		hoo.New(cfg.HooURL),
		cfg,
	))
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
