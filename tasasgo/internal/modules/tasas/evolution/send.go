package evolution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SendMedia publica una imagen en base64 en la instancia configurada.
//
// En modo simulación devuelve un estado sintético sin tocar la red, de modo que
// un despliegue sin credenciales sigue siendo ejecutable y verificable.
//
// El estado HTTP se valida ANTES de decodificar el cuerpo: cuando Evolution
// rechaza una petición responde HTML, y leerlo como JSON ocultaría la causa real
// (credencial inválida o instancia inexistente).
func (c *Client) SendMedia(ctx context.Context, req SendMediaRequest) (*SendMediaResponse, error) {
	if c.Simulated() {
		return &SendMediaResponse{Status: SimulationStatus}, nil
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("serializar petición: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("construir petición: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("apikey", c.apiKey)

	response, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("enviar media: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("evolution respondió estado %d", response.StatusCode)
	}

	result := &SendMediaResponse{}
	if err := json.NewDecoder(response.Body).Decode(result); err != nil {
		return nil, fmt.Errorf("decodificar respuesta: %w", err)
	}
	return result, nil
}

// endpoint compone la URL del envío incluyendo la instancia en la ruta.
//
// La instancia es el valor que cambia según el despliegue (Glauk u otra versión
// de pruebas), por eso su construcción se aísla aquí y no dentro del envío.
func (c *Client) endpoint() string {
	return fmt.Sprintf("%s/message/sendMedia/%s", c.baseURL, c.instance)
}
