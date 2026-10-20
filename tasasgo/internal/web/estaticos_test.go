package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// raizDelPanel apunta al directorio que se publica de verdad, no a una copia de
// prueba: lo que hay que demostrar es que el panel viaja completo, y un fixture
// podría seguir pasando mientras el directorio real se queda sin un archivo.
func raizDelPanel() string { return filepath.Join("..", "..", "public") }

// montar arma la aplicación en el mismo orden que cmd/server: primero la API,
// después los archivos del panel y, al final, el 404 en JSON.
func montar(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Get("/api/v1/tasas/monitor", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true, "origen": "api"})
	})
	app.Post("/api/v1/tasas/publicar-lote", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true, "origen": "api"})
	})
	Register(app, raizDelPanel())
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false})
	})
	return app
}

// pedir ejecuta la petición contra la aplicación sin abrir un puerto.
func pedir(t *testing.T, app *fiber.App, metodo, ruta string) (*http.Response, []byte) {
	t.Helper()
	respuesta, err := app.Test(httptest.NewRequest(metodo, ruta, nil))
	if err != nil {
		t.Fatalf("%s %s: %v", metodo, ruta, err)
	}
	t.Cleanup(func() { respuesta.Body.Close() })
	cuerpo, err := io.ReadAll(respuesta.Body)
	if err != nil {
		t.Fatalf("%s %s: leyendo el cuerpo: %v", metodo, ruta, err)
	}
	return respuesta, cuerpo
}

// TestLaRaizSirveElPanel comprueba que "/" devuelve el dashboard y no un listado
// del directorio: es la dirección que se abre en el laboratorio.
func TestLaRaizSirveElPanel(t *testing.T) {
	respuesta, cuerpo := pedir(t, montar(t), http.MethodGet, "/")
	if respuesta.StatusCode != http.StatusOK {
		t.Errorf("estado = %d, se esperaba 200", respuesta.StatusCode)
	}
	if tipo := respuesta.Header.Get("Content-Type"); !strings.Contains(tipo, "text/html") {
		t.Errorf("Content-Type = %q, se esperaba HTML", tipo)
	}
	if !strings.Contains(string(cuerpo), "TasasHub Engine") {
		t.Error("el cuerpo no parece el panel")
	}
}

// TestLosArchivosDelPanelSeSirven recorre lo que el navegador pide al abrir el
// panel: el HTML referencia el manifiesto, el service worker y las banderas, y
// ninguno de los tres puede faltar sin que la interfaz quede a medias.
func TestLosArchivosDelPanelSeSirven(t *testing.T) {
	casos := []struct {
		ruta  string
		tipo  string
		bytes []byte
	}{
		{"/manifest.json", "application/json", nil},
		{"/sw.js", "javascript", nil},
		{"/flags/ar.png", "image/png", []byte{0x89, 'P', 'N', 'G'}},
	}
	for _, caso := range casos {
		respuesta, cuerpo := pedir(t, montar(t), http.MethodGet, caso.ruta)
		if respuesta.StatusCode != http.StatusOK {
			t.Errorf("%s: estado = %d, se esperaba 200", caso.ruta, respuesta.StatusCode)
			continue
		}
		if tipo := respuesta.Header.Get("Content-Type"); !strings.Contains(tipo, caso.tipo) {
			t.Errorf("%s: Content-Type = %q, se esperaba %q", caso.ruta, tipo, caso.tipo)
		}
		if caso.bytes != nil && !strings.HasPrefix(string(cuerpo), string(caso.bytes)) {
			t.Errorf("%s: el cuerpo no empieza por la firma %q", caso.ruta, caso.bytes)
		}
	}
}
