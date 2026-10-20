package tasas

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/pashagolub/pgxmock/v4"

	"tasasgo/internal/config"
	"tasasgo/internal/modules/tasas/evolution"
)

// columnasLote y columnasSocio son las columnas que devuelven las lecturas simuladas.
var columnasLote = []string{"id_tasa", "tasas"}
var columnasSocio = []string{"nombre", "id_grupo", "moneda_base", "monedas", "activo", "modo_prueba", "rol"}

// texto y booleano arman las columnas anulables, porque es como pgx entrega una
// columna que puede ser nula: un valor suelto no puede escanearse en un puntero.
func texto(valor string) *string { return &valor }
func booleano(valor bool) *bool  { return &valor }

// nuevaApp monta el módulo con dobles: pool simulado, Evolution en modo simulación
// (sin red) y el origen HOO nulo, que es su estado hasta la Fase 6. El PNG se
// dibuja de verdad, porque el render es una función pura sobre recursos embebidos.
func nuevaApp(t *testing.T) (*fiber.App, pgxmock.PgxPoolIface) {
	t.Helper()
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("no se pudo crear el doble del pool: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	deps := NewDeps(pool, evolution.New("", "", "Glauk"), nil, &config.Config{JIDGrupoPrueba: "prueba@g.us"})
	app := fiber.New()
	Register(app, deps)
	return app, pool
}

// pedir ejecuta una petición y devuelve estado, tipo de contenido y cuerpo, que es
// todo lo que estas pruebas comparan: no abre un puerto ni sale a la red.
func pedir(t *testing.T, app *fiber.App, metodo, ruta, cuerpo string) (int, string, string) {
	t.Helper()
	peticion := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	peticion.Header.Set("Content-Type", "application/json")
	respuesta, err := app.Test(peticion)
	if err != nil {
		t.Fatalf("la petición falló: %v", err)
	}
	leido, _ := io.ReadAll(respuesta.Body)
	return respuesta.StatusCode, respuesta.Header.Get("Content-Type"), string(leido)
}

func TestLecturasYRenderDeLaCartelera(t *testing.T) {
	app, pool := nuevaApp(t)
	pool.ExpectQuery(`FROM tasas_glaukov`).WillReturnRows(pgxmock.NewRows(columnasLote).AddRow(texto("T060"), []byte(`{"COP":4050}`)))
	pool.ExpectQuery(`FROM perfiles_glaukov`).WithArgs("Ana").WillReturnRows(pgxmock.NewRows(columnasSocio).AddRow("Ana", nil, nil, []byte(`{}`), booleano(true), nil, nil))
	pool.ExpectQuery(`FROM tasas_glaukov`).WillReturnRows(pgxmock.NewRows(columnasLote).AddRow(texto("T060"), []byte(`{"COP":4050}`)))
	pool.ExpectQuery(`FROM perfiles_glaukov`).WithArgs("Nadie").WillReturnRows(pgxmock.NewRows(columnasSocio))

	estado, _, cuerpo := pedir(t, app, http.MethodGet, "/api/v1/tasas/monitor", "")
	if estado != http.StatusOK || !strings.Contains(cuerpo, `"id_tasa":"T060"`) || !strings.Contains(cuerpo, `"success":true`) {
		t.Fatalf("se esperaba el lote vigente en /monitor, llegó %d: %s", estado, cuerpo)
	}
	estado, tipo, _ := pedir(t, app, http.MethodGet, "/api/v1/tasas/render/Ana", "")
	if estado != http.StatusOK || tipo != "image/png" {
		t.Fatalf("se esperaba un PNG, llegó %d como %s", estado, tipo)
	}
	estado, _, cuerpo = pedir(t, app, http.MethodGet, "/api/v1/tasas/render/Nadie", "")
	if estado != http.StatusNotFound || cuerpo != "Socio no encontrado" {
		t.Fatalf("se esperaba el 404 en texto plano, llegó %d: %s", estado, cuerpo)
	}
}

func TestDispararLoteOrdenaElDetalleYSaltaApagados(t *testing.T) {
	app, pool := nuevaApp(t)
	pool.ExpectQuery(`FROM tasas_glaukov`).WillReturnRows(pgxmock.NewRows(columnasLote).AddRow(texto("T060"), []byte(`{"COP":4050}`)))
	pool.ExpectQuery(`FROM perfiles_glaukov`).WillReturnRows(pgxmock.NewRows(columnasSocio).
		AddRow("Zoe", nil, nil, []byte(`{}`), booleano(true), nil, nil).
		AddRow("Ana", nil, nil, []byte(`{}`), booleano(true), nil, nil).
		AddRow("Beto", nil, nil, []byte(`{}`), booleano(false), nil, nil))

	estado, _, cuerpo := pedir(t, app, http.MethodPost, "/api/v1/tasas/disparar-lote", `{"modo_prueba_global":true}`)
	if estado != http.StatusOK || !strings.Contains(cuerpo, `"detalle":[{"socio":"Ana","status":"ok"},{"socio":"Zoe","status":"ok"}]`) {
		t.Fatalf("se esperaba el detalle ordenado y sin el socio apagado, llegó %d: %s", estado, cuerpo)
	}
}
