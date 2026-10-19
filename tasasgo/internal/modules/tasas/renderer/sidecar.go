package renderer

import (
	"context"
	"net/http"
	"strings"
	"time"

	"tasasgo/internal/modules/tasas/dominio"
)

// defaultTimeout acota la espera al sidecar, que tarda segundos por imagen.
const defaultTimeout = 30 * time.Second

// Sidecar es la implementación de Renderer que habla con el servicio Node por
// HTTP; su estado es mínimo (la URL resuelta, un cliente y el plazo) y nunca
// global, de modo que cada prueba construye el suyo contra un servidor local.
type Sidecar struct {
	url     string
	client  *http.Client
	timeout time.Duration
}

// NewSidecar apunta el cliente a la base del sidecar y compone la ruta /render; la
// ruta se escribe aquí, en un único lugar, para que nadie más tenga que conocerla.
func NewSidecar(baseURL string) *Sidecar {
	return &Sidecar{
		url:     strings.TrimRight(baseURL, "/") + "/render",
		client:  &http.Client{},
		timeout: defaultTimeout,
	}
}

// Sidecar satisface el contrato del paquete: si Render cambiara de firma, esta
// afirmación dejaría de compilar.
var _ Renderer = (*Sidecar)(nil)

// Render serializa la cartelera y la envía por POST al sidecar, devolviendo los
// bytes PNG de la respuesta.
//
// El éxito se define como HTTP 200 con el PNG en el cuerpo; cualquier otro código
// llega como StatusError, y cualquier fallo de red o de plazo como error envuelto
// con %w, para que quien llama pueda distinguir un caso del otro.
//
// El plazo aplicado es el más corto entre el que trae quien llama y el propio del
// cliente: un handler puede imponer su presupuesto sin que el cliente deje de
// protegerse de una espera infinita.
func (s *Sidecar) Render(ctx context.Context, cartelera dominio.Cartelera) ([]byte, error) {
	body, err := serializar(cartelera)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.post(ctx, body)
}
