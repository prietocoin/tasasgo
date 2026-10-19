package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// pedir monta una aplicación mínima con el middleware ya aplicado y consulta una ruta
// cualquiera, que es la forma de comprobar que las cabeceras salen en todas las
// respuestas y no solo en las que alguien se acordó de decorar.
//
// La capa previa pone X-Frame-Options a propósito: nada en Fiber la añade por su
// cuenta, y la eliminación solo se puede comprobar si algo llegó a ponerla.
func pedir(t *testing.T) *http.Response {
	t.Helper()
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Frame-Options", "DENY")
		return c.Next()
	})
	Register(app)
	app.Get("/panel", func(c *fiber.Ctx) error { return c.SendString("ok") })

	peticion := httptest.NewRequest(http.MethodGet, "/panel", nil)
	peticion.Header.Set("Origin", "https://lab.jairokov.com")
	respuesta, err := app.Test(peticion)
	if err != nil {
		t.Fatalf("la petición falló: %v", err)
	}
	return respuesta
}

func TestCabecerasDeSeguridad(t *testing.T) {
	respuesta := pedir(t)

	if csp := respuesta.Header.Get("Content-Security-Policy"); csp != "frame-ancestors 'self' https://lab.jairokov.com" {
		t.Fatalf("el CSP debe limitar los marcos al laboratorio, llegó %q", csp)
	}
	if marco := respuesta.Header.Get("X-Frame-Options"); marco != "" {
		t.Fatalf("X-Frame-Options debe eliminarse o el iframe deja de funcionar, llegó %q", marco)
	}
	if origen := respuesta.Header.Get("Access-Control-Allow-Origin"); origen != "*" {
		t.Fatalf("CORS debe quedar abierto mientras Node y Go coexisten, llegó %q", origen)
	}
}
