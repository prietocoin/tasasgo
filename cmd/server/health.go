package main

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"tasasgo/internal/config"
)

// probeTimeout acota la comprobación de base de datos dentro del handler. Un
// health check que espera indefinidamente deja de ser un health check: debe
// responder rápido incluso cuando la dependencia está caída.
const probeTimeout = 2 * time.Second

// healthHandler construye el handler de /health inyectando sus dependencias por
// cierre en lugar de leer variables globales, para poder probarlo con una
// configuración explícita.
//
// El payload expone la instancia de Evolution resuelta y el modo simulación: son
// los dos valores que más se confunden al desplegar, y verlos aquí permite
// confirmar qué instancia está activa sin abrir el archivo de entorno y sin
// recompilar el binario.
func healthHandler(cfg *config.Config, pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "OK",
			"service":   "TasasHub Engine ⚡",
			"instance":  cfg.EvolutionInstance,
			"database":  databaseStatus(pool),
			"simulated": cfg.SimulationMode(),
			"timestamp": time.Now().UTC(),
		})
	}
}

// databaseStatus responde "ok" o "unreachable" sin propagar el error.
//
// Que la base se haya caído no convierte al proceso en no saludable: el servicio
// sigue atendiendo, y el detalle del fallo pertenece al log, no a un payload
// público que cualquiera puede consultar.
func databaseStatus(pool *pgxpool.Pool) string {
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return "unreachable"
	}
	return "ok"
}
