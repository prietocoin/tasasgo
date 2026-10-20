package evolution

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSendMediaSimulation verifica que sin URL configurada no se abre ninguna
// conexión de red. Es la garantía de que el servicio arranca y opera en un
// entorno sin WhatsApp ni credenciales.
func TestSendMediaSimulation(t *testing.T) {
	client := New("", "clave-de-prueba", "GLAUK")

	response, err := client.SendMedia(context.Background(), SendMediaRequest{Number: "120363@g.us"})
	if err != nil {
		t.Fatalf("no se esperaba error en modo simulación: %v", err)
	}
	if response.Status != SimulationStatus {
		t.Fatalf("estado inesperado: se esperaba %q y llegó %q", SimulationStatus, response.Status)
	}
	if !client.Simulated() {
		t.Fatal("el cliente debería reportarse en modo simulación")
	}
}

// TestSendMediaRequestContract fija los tres hechos centrales del diagnóstico:
// un solo header de autenticación, la instancia viajando en la URL y el payload
// plano con base64 puro.
func TestSendMediaRequestContract(t *testing.T) {
	var headers http.Header
	var path string
	var decoded SendMediaRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = r.Header.Clone()
		path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&decoded)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	client := New(server.URL, "clave-de-prueba", "GLAUK")
	request := SendMediaRequest{
		Number: "120363@g.us", MediaType: "image", MimeType: "image/png",
		FileName: "cartelera.png", Media: "iVBORw0KGgo=", Caption: "prueba",
	}

	response, err := client.SendMedia(context.Background(), request)
	if err != nil {
		t.Fatalf("envío inesperadamente fallido: %v", err)
	}
	if response.Status != "success" {
		t.Fatalf("estado inesperado: %q", response.Status)
	}
	if path != "/message/sendMedia/GLAUK" {
		t.Fatalf("la instancia debería viajar en la ruta, se recibió %q", path)
	}
	if got := headers.Get("apikey"); got != "clave-de-prueba" {
		t.Fatalf("header apikey incorrecto: %q", got)
	}
	for _, forbidden := range []string{"Global-Api-Key", "Authorization"} {
		if headers.Get(forbidden) != "" {
			t.Fatalf("el header %s no debería enviarse", forbidden)
		}
	}
	if decoded.Media != "iVBORw0KGgo=" || decoded.MediaType != "image" {
		t.Fatalf("payload alterado en el camino: %+v", decoded)
	}
}

// TestSendMediaReportsHTTPError comprueba que una respuesta de error se reporta
// como error de transporte y no como fallo de decodificación. Validar el estado
// antes de leer el cuerpo es lo que permite atribuir un 401 a la credencial.
func TestSendMediaReportsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("<html>no autorizado</html>"))
	}))
	defer server.Close()

	client := New(server.URL, "clave-incorrecta", "GLAUK")

	if _, err := client.SendMedia(context.Background(), SendMediaRequest{}); err == nil {
		t.Fatal("se esperaba error ante una respuesta 401")
	}
}
