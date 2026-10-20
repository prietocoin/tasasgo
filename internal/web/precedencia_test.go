package web

import (
	"net/http"
	"strings"
	"testing"
)

// TestLaApiTienePrecedencia es la razón de este paquete, no un detalle: el
// servicio de archivos responde a cualquier ruta, así que si se montara antes
// que la API el panel taparía /api/v1/tasas y el dashboard dejaría de funcionar
// sin que ninguna otra prueba se queje.
func TestLaApiTienePrecedencia(t *testing.T) {
	casos := []struct{ metodo, ruta string }{
		{http.MethodGet, "/api/v1/tasas/monitor"},
		{http.MethodPost, "/api/v1/tasas/publicar-lote"},
	}
	for _, caso := range casos {
		respuesta, cuerpo := pedir(t, montar(t), caso.metodo, caso.ruta)
		if respuesta.StatusCode != http.StatusOK {
			t.Errorf("%s %s: estado = %d, se esperaba 200", caso.metodo, caso.ruta, respuesta.StatusCode)
		}
		if !strings.Contains(string(cuerpo), `"origen":"api"`) {
			t.Errorf("%s %s: respondió %s, se esperaba el handler de la API", caso.metodo, caso.ruta, cuerpo)
		}
	}
}

// TestLoDesconocidoSigueSiendoEl404DelServicio comprueba las dos caras del
// mismo acuerdo: el panel no inventa una página para una ruta que no existe, y
// la API tampoco pierde su 404 en JSON por culpa del servicio de archivos.
func TestLoDesconocidoSigueSiendoEl404DelServicio(t *testing.T) {
	casos := []string{"/no-existe.html", "/api/v1/tasas/inexistente"}
	for _, ruta := range casos {
		respuesta, cuerpo := pedir(t, montar(t), http.MethodGet, ruta)
		if respuesta.StatusCode != http.StatusNotFound {
			t.Errorf("%s: estado = %d, se esperaba 404", ruta, respuesta.StatusCode)
		}
		if tipo := respuesta.Header.Get("Content-Type"); !strings.Contains(tipo, "application/json") {
			t.Errorf("%s: Content-Type = %q, se esperaba JSON", ruta, tipo)
		}
		if !strings.Contains(string(cuerpo), `"success":false`) {
			t.Errorf("%s: cuerpo = %s, se esperaba el error del servicio", ruta, cuerpo)
		}
	}
}
