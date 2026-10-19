package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"tasasgo/internal/config"
	"tasasgo/internal/middleware"
)

// bodyLimit replica el límite de 10 MB que el servicio Node aplica al JSON
// entrante. Las carteleras viajan como base64 y un límite menor rompería el envío
// de imágenes grandes justo cuando todo lo demás ya funciona.
const bodyLimit = 10 * 1024 * 1024

// newApp construye el servidor HTTP con sus ajustes, su manejo de errores y sus
// rutas ya montadas.
//
// Se separa de main para que el ensamblado sea inspeccionable y verificable sin
// abrir un puerto, y para que main quede reducido a decidir el orden de arranque.
func newApp(cfg *config.Config, pool *pgxpool.Pool) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "TasasHub Engine ⚡",
		BodyLimit:    bodyLimit,
		ErrorHandler: jsonErrorHandler,
	})
	middleware.Register(app)
	registerRoutes(app, cfg, pool)
	return app
}

// jsonErrorHandler fuerza que todo error se responda en JSON, con la misma forma
// que usa el servicio Node ({success, error}).
//
// Sin esto, fiber responde texto plano y el dashboard recibe dos formatos
// distintos según qué servicio haya fallado, lo que obliga a adivinar en el
// cliente en lugar de leer un campo.
func jsonErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if httpError, ok := err.(*fiber.Error); ok {
		code = httpError.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error":   err.Error(),
	})
}
