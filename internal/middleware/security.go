// Package middleware reúne la seguridad transversal que el servicio aplica a todas
// sus respuestas, para que ninguna ruta tenga que acordarse de ella.
//
// Replica lo que hacía server.js en Node: CORS abierto y una política de marcos que
// permite embeber el panel en el laboratorio. Vive en su propio paquete porque es una
// decisión de despliegue, no de negocio.
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// politicaDeMarcos es la regla de embeber: el propio origen y el laboratorio.
const politicaDeMarcos = "frame-ancestors 'self' https://lab.jairokov.com"

// Register monta CORS y la política de marcos.
//
// CORS queda abierto porque el panel se sirve desde otro origen mientras Node y Go
// coexisten; la política de marcos es la que decide quién puede mostrarlo en un
// iframe, y esa decisión no es del navegador de cada socio.
func Register(app *fiber.App) {
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
	}))
	app.Use(marcosPermitidos)
}

// marcosPermitidos fija el CSP y elimina X-Frame-Options.
//
// El original borraba esa cabecera a propósito: su DENY por defecto anularía la
// política y el laboratorio dejaría de poder mostrar el panel.
func marcosPermitidos(c *fiber.Ctx) error {
	c.Set("Content-Security-Policy", politicaDeMarcos)
	c.Response().Header.Del("X-Frame-Options")
	return c.Next()
}
