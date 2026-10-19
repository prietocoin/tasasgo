package renderer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// maxErrorBody limita cuánto de la respuesta de error se conserva: suficiente para
// diagnosticar sin arrastrar respuestas enormes dentro de un error.
const maxErrorBody = 2 << 10

// post envía el cuerpo ya serializado y devuelve la respuesta. Un código distinto
// de 200 se convierte en StatusError; cualquier fallo de red o de plazo vuelve
// envuelto con %w para no perder la causa original.
func (s *Sidecar) post(ctx context.Context, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("construir petición HTTP: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llamar al renderer: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		trozo, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		return nil, &StatusError{Code: resp.StatusCode, Body: string(trozo)}
	}
	return io.ReadAll(resp.Body)
}
