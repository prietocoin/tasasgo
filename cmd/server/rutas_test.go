package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"tasasgo/internal/config"
)

// pedir ejecuta la petición contra la aplicación sin abrir un puerto.
func pedir(t *testing.T, app *fiber.App, ruta string) (*http.Response, string) {
	t.Helper()
	respuesta, err := app.Test(httptest.NewRequest(http.MethodGet, ruta, nil))
	if err != nil {
		t.Fatalf("GET %s: %v", ruta, err)
	}
	t.Cleanup(func() { respuesta.Body.Close() })
	cuerpo, err := io.ReadAll(respuesta.Body)
	if err != nil {
		t.Fatalf("GET %s: leyendo el cuerpo: %v", ruta, err)
	}
	return respuesta, string(cuerpo)
}

// TestElMontajeRealSirveElPanel apunta al ensamblado de cmd/server, no al del
// paquete web: aquí se ve si el directorio que el servicio espera es el que el
// despliegue le entrega, y si el panel convive con la API y con el 404.
//
// El proceso de prueba arranca en la carpeta del paquete, así que se mueve a la
// del módulo, que es desde donde se ejecuta el servicio de verdad. El pool va en
// nil a propósito: el panel no lo toca y la suite no necesita una base de datos.
func TestElMontajeRealSirveElPanel(t *testing.T) {
	t.Chdir("../..")
	app := newApp(&config.Config{}, nil)

	casos := []struct {
		ruta     string
		estado   int
		contiene string
	}{
		{"/", http.StatusOK, "TasasHub Engine"},
		{"/manifest.json", http.StatusOK, `"short_name": "TasasHub"`},
		{"/sw.js", http.StatusOK, "addEventListener"},
		{"/flags/ve.png", http.StatusOK, "PNG"},
		{"/api/v1/tasas/inexistente", http.StatusNotFound, `"success":false`},
		{"/no-existe", http.StatusNotFound, `"success":false`},
	}
	for _, caso := range casos {
		respuesta, cuerpo := pedir(t, app, caso.ruta)
		if respuesta.StatusCode != caso.estado {
			t.Errorf("%s: estado = %d, se esperaba %d", caso.ruta, respuesta.StatusCode, caso.estado)
			continue
		}
		if !strings.Contains(cuerpo, caso.contiene) {
			t.Errorf("%s: el cuerpo no contiene %q", caso.ruta, caso.contiene)
		}
	}
}
