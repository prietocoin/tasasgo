package renderer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tasasgo/internal/modules/tasas/dominio"
)

// carteleraDePrueba trae una fila completa y otra a la que le faltan nombre de
// país, bandera y tendencia, para comprobar que la normalización las completa.
func carteleraDePrueba() dominio.Cartelera {
	return dominio.Cartelera{
		NombreSocio: "Ana",
		LoteTasa:    "T001",
		Fecha:       "17-oct",
		Hora:        "09:30",
		Tarjetas: []dominio.TarjetaPais{
			{Code: "USD", NombrePais: "Estados Unidos", Bandera: "us", Compra: 40.1, Venta: 41.2, Trend: "up"},
			{Code: "EUR", Compra: 44, Venta: 45},
		},
	}
}

func TestRenderDevuelveElPNGDelSidecar(t *testing.T) {
	esperado := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}
	var recibido Payload
	servidor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/render" {
			t.Errorf("petición inesperada: %s %s", r.Method, r.URL.Path)
		}
		cuerpo, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(cuerpo, &recibido); err != nil {
			t.Errorf("el cuerpo no es la cartelera esperada: %v", err)
		}
		_, _ = w.Write(esperado)
	}))
	defer servidor.Close()

	png, err := NewSidecar(servidor.URL).Render(context.Background(), carteleraDePrueba())
	if err != nil {
		t.Fatalf("Render devolvió error: %v", err)
	}
	if string(png) != string(esperado) {
		t.Fatalf("PNG distinto: %q", png)
	}
	if recibido.NombreSocio != "Ana" || len(recibido.Tarjetas) != 2 {
		t.Errorf("cartelera mal serializada: %+v", recibido)
	}
	segunda := recibido.Tarjetas[1]
	if segunda.Pais != "EUR" || segunda.Bandera != "eur" || segunda.Trend != "equal" {
		t.Errorf("normalización incompleta: %+v", segunda)
	}
}

func TestRenderConvierteElErrorDelSidecarEnStatusError(t *testing.T) {
	servidor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "satori explotó", http.StatusInternalServerError)
	}))
	defer servidor.Close()

	_, err := NewSidecar(servidor.URL).Render(context.Background(), carteleraDePrueba())

	var status *StatusError
	if !errors.As(err, &status) {
		t.Fatalf("se esperaba StatusError, llegó: %v", err)
	}
	if status.Code != http.StatusInternalServerError || status.Body == "" {
		t.Errorf("StatusError incompleto: %+v", status)
	}
}

func TestRenderRespetaElPlazoDeEspera(t *testing.T) {
	servidor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer servidor.Close()

	cliente := NewSidecar(servidor.URL)
	cliente.timeout = 20 * time.Millisecond

	_, err := cliente.Render(context.Background(), carteleraDePrueba())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("se esperaba plazo agotado, llegó: %v", err)
	}
}
